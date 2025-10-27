package web

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// Config contains configuration for web operations
type Config struct {
	// Timeout for HTTP requests
	Timeout time.Duration

	// MaxResponseSize limits the maximum response size (in bytes)
	MaxResponseSize int64

	// AllowedDomains restricts requests to these domains (empty = allow all)
	AllowedDomains []string

	// FollowRedirects enables following HTTP redirects
	FollowRedirects bool

	// MaxRedirects limits the number of redirects to follow
	MaxRedirects int

	// UserAgent sets the User-Agent header for requests
	UserAgent string

	// AllowPrivateIPs allows requests to private IP addresses (DANGEROUS)
	AllowPrivateIPs bool
}

// DefaultConfig provides sensible defaults for web operations
var DefaultConfig = Config{
	Timeout:         30 * time.Second,
	MaxResponseSize: 1 * 1024 * 1024, // 1MB
	AllowedDomains:  []string{},
	FollowRedirects: true,
	MaxRedirects:    10,
	UserAgent:       "go-llm-agent/1.0",
	AllowPrivateIPs: false,
}

// FetchTool performs HTTP GET requests
type FetchTool struct {
	tools.BaseTool
	config Config
	client *http.Client
}

// NewFetchTool creates a new web fetch tool with the given configuration
func NewFetchTool(config Config) *FetchTool {
	// Create HTTP client with custom transport for security
	transport := &http.Transport{
		DialContext: createSecureDialContext(config.AllowPrivateIPs),
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	// Configure redirect policy
	if !config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if config.MaxRedirects > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		}
	}

	return &FetchTool{
		BaseTool: tools.NewBaseTool(
			"web_fetch",
			"Fetch content from a URL using HTTP GET. Returns the response body, status code, and headers.",
			tools.CategoryWeb,
			false, // no auth required
			true,  // safe operation (read-only)
		),
		config: config,
		client: client,
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *FetchTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"url": {
				Type:        "string",
				Description: "The URL to fetch (must be HTTP or HTTPS)",
			},
			"headers": {
				Type:        "object",
				Description: "Optional custom HTTP headers as key-value pairs",
			},
		},
		Required: []string{"url"},
	}
}

// Execute fetches content from a URL
func (t *FetchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	urlStr, headers, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Validate URL
	parsedURL, err := t.validateURL(urlStr)
	if err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", t.config.UserAgent)

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, fmt.Sprint(value))
	}

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body with size limit
	body, err := t.readResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}

	// Build result
	return t.buildResult(resp, body), nil
}

// extractParams extracts and validates parameters
func (t *FetchTool) extractParams(params map[string]interface{}) (string, map[string]interface{}, error) {
	urlStr, ok := params["url"].(string)
	if !ok || urlStr == "" {
		return "", nil, fmt.Errorf("url parameter is required and must be a non-empty string")
	}

	headers := make(map[string]interface{})
	if headersParam, ok := params["headers"].(map[string]interface{}); ok {
		headers = headersParam
	}

	return urlStr, headers, nil
}

// validateURL validates and parses the URL
func (t *FetchTool) validateURL(urlStr string) (*url.URL, error) {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("only HTTP and HTTPS schemes are allowed, got: %s", parsedURL.Scheme)
	}

	// Check host
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("URL must have a host")
	}

	// Check against allowed domains
	if len(t.config.AllowedDomains) > 0 {
		if !t.isDomainAllowed(parsedURL.Host) {
			return nil, fmt.Errorf("domain %s is not in allowed domains list", parsedURL.Host)
		}
	}

	// Check for localhost/private IPs in hostname
	if !t.config.AllowPrivateIPs {
		host := parsedURL.Hostname()
		if isLocalhost(host) {
			return nil, fmt.Errorf("requests to localhost are not allowed")
		}
	}

	return parsedURL, nil
}

// isDomainAllowed checks if a domain is in the allowed list
func (t *FetchTool) isDomainAllowed(host string) bool {
	// Extract hostname without port
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	for _, allowed := range t.config.AllowedDomains {
		if hostname == allowed || strings.HasSuffix(hostname, "."+allowed) {
			return true
		}
	}

	return false
}

// readResponseBody reads the response body with size limit
func (t *FetchTool) readResponseBody(body io.Reader) (string, error) {
	if t.config.MaxResponseSize > 0 {
		body = io.LimitReader(body, t.config.MaxResponseSize+1)
	}

	content, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if t.config.MaxResponseSize > 0 && int64(len(content)) > t.config.MaxResponseSize {
		return "", fmt.Errorf("response size (%d bytes) exceeds maximum allowed size (%d bytes)",
			len(content), t.config.MaxResponseSize)
	}

	return string(content), nil
}

// buildResult creates the result map
func (t *FetchTool) buildResult(resp *http.Response, body string) map[string]interface{} {
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return map[string]interface{}{
		"success":     true,
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     headers,
		"body":        body,
		"body_size":   len(body),
		"url":         resp.Request.URL.String(),
	}
}

// isPrivateIP checks if an IP is private
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip) {
			return true
		}
	}

	return false
}

// isLocalhost checks if a hostname is localhost
func isLocalhost(host string) bool {
	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		host == "0.0.0.0" ||
		strings.HasPrefix(host, "127.") ||
		strings.HasPrefix(host, "0.")
}
