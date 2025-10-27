package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// PostConfig holds configuration for POST requests
type PostConfig struct {
	Timeout         time.Duration
	MaxResponseSize int64
	AllowedDomains  []string
	FollowRedirects bool
	MaxRedirects    int
	UserAgent       string
	AllowPrivateIPs bool
}

// PostTool implements HTTP POST functionality
type PostTool struct {
	tools.BaseTool
	config PostConfig
	client *http.Client
}

// NewPostTool creates a new POST tool with custom configuration
func NewPostTool(config PostConfig) *PostTool {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxResponseSize == 0 {
		config.MaxResponseSize = 1024 * 1024 // 1MB default
	}
	if config.MaxRedirects == 0 {
		config.MaxRedirects = 5
	}
	if config.UserAgent == "" {
		config.UserAgent = "GoLLMAgent/1.0"
	}

	// Create HTTP client with custom transport for security
	transport := &http.Transport{
		DialContext: createSecureDialContext(config.AllowPrivateIPs),
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

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

	return &PostTool{
		BaseTool: tools.NewBaseTool(
			"web_post",
			"Send HTTP POST requests to web servers with JSON or form data. Supports custom headers and security controls.",
			tools.CategoryWeb,
			false, // no auth required
			true,  // safe operation
		),
		config: config,
		client: client,
	}
}

// Parameters returns the tool parameter schema
func (t *PostTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"url": {
				Type:        "string",
				Description: "The URL to send the POST request to",
			},
			"body": {
				Type:        "object",
				Description: "The request body data (will be sent as JSON)",
			},
			"headers": {
				Type:        "object",
				Description: "Optional HTTP headers to include in the request",
			},
			"form_data": {
				Type:        "object",
				Description: "Optional form data (if provided, body will be ignored and Content-Type will be application/x-www-form-urlencoded)",
			},
		},
		Required: []string{"url"},
	}
}

// Execute performs the POST request
func (t *PostTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract and validate parameters
	urlStr, body, formData, headers, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Validate URL using fetch tool's validation
	fetchTool := &FetchTool{config: Config{
		AllowedDomains:  t.config.AllowedDomains,
		AllowPrivateIPs: t.config.AllowPrivateIPs,
	}}

	parsedURL, err := fetchTool.validateURL(urlStr)
	if err != nil {
		return nil, err
	}

	// Prepare request body and content type
	var reqBody io.Reader
	var contentType string

	if formData != nil {
		// Form data takes precedence
		formValues := url.Values{}
		for k, v := range formData {
			formValues.Set(k, fmt.Sprint(v))
		}
		reqBody = strings.NewReader(formValues.Encode())
		contentType = "application/x-www-form-urlencoded"
	} else if body != nil {
		// JSON body
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
		contentType = "application/json"
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsedURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type if we have a body
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Set user agent
	req.Header.Set("User-Agent", t.config.UserAgent)

	// Set custom headers (may override defaults)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := t.readResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}

	// Build result
	return t.buildResult(resp, responseBody, parsedURL.String()), nil
}

// extractParams extracts and validates parameters
func (t *PostTool) extractParams(params map[string]interface{}) (string, map[string]interface{}, map[string]interface{}, map[string]string, error) {
	// Extract URL
	urlVal, ok := params["url"]
	if !ok {
		return "", nil, nil, nil, fmt.Errorf("url parameter is required")
	}

	urlStr, ok := urlVal.(string)
	if !ok {
		return "", nil, nil, nil, fmt.Errorf("url must be a string")
	}

	if urlStr == "" {
		return "", nil, nil, nil, fmt.Errorf("url cannot be empty")
	}

	// Extract body (optional)
	var body map[string]interface{}
	if bodyVal, ok := params["body"]; ok && bodyVal != nil {
		body, ok = bodyVal.(map[string]interface{})
		if !ok {
			return "", nil, nil, nil, fmt.Errorf("body must be an object")
		}
	}

	// Extract form_data (optional)
	var formData map[string]interface{}
	if formVal, ok := params["form_data"]; ok && formVal != nil {
		formData, ok = formVal.(map[string]interface{})
		if !ok {
			return "", nil, nil, nil, fmt.Errorf("form_data must be an object")
		}
	}

	// Extract headers (optional)
	headers := make(map[string]string)
	if headersVal, ok := params["headers"]; ok && headersVal != nil {
		headersMap, ok := headersVal.(map[string]interface{})
		if !ok {
			return "", nil, nil, nil, fmt.Errorf("headers must be an object")
		}

		for k, v := range headersMap {
			headers[k] = fmt.Sprint(v)
		}
	}

	return urlStr, body, formData, headers, nil
}

// readResponseBody reads the response body with size limit (reuses fetch tool logic)
func (t *PostTool) readResponseBody(body io.Reader) (string, error) {
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

// buildResult creates the result object
func (t *PostTool) buildResult(resp *http.Response, body string, finalURL string) map[string]interface{} {
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
		"url":         finalURL,
	}
}
