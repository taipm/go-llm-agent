package web

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// ScrapeConfig holds configuration for web scraping
type ScrapeConfig struct {
	Timeout         time.Duration
	MaxResponseSize int64
	AllowedDomains  []string
	UserAgent       string
	AllowPrivateIPs bool
	RateLimit       time.Duration // Minimum delay between requests
}

// ScrapeTool implements web scraping with CSS selectors
type ScrapeTool struct {
	tools.BaseTool
	config      ScrapeConfig
	fetchTool   *FetchTool
	lastRequest time.Time
}

// NewScrapeTool creates a new scraping tool with custom configuration
func NewScrapeTool(config ScrapeConfig) *ScrapeTool {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxResponseSize == 0 {
		config.MaxResponseSize = 5 * 1024 * 1024 // 5MB default for HTML
	}
	if config.UserAgent == "" {
		config.UserAgent = "GoLLMAgent/1.0"
	}

	// Create fetch tool for HTTP requests
	fetchConfig := Config{
		Timeout:         config.Timeout,
		MaxResponseSize: config.MaxResponseSize,
		AllowedDomains:  config.AllowedDomains,
		UserAgent:       config.UserAgent,
		AllowPrivateIPs: config.AllowPrivateIPs,
		FollowRedirects: true,
		MaxRedirects:    5,
	}

	return &ScrapeTool{
		BaseTool: tools.NewBaseTool(
			"web_scrape",
			"Scrape and extract content from web pages using CSS selectors. Returns text content and HTML attributes.",
			tools.CategoryWeb,
			false, // no auth required
			true,  // safe operation
		),
		config:    config,
		fetchTool: NewFetchTool(fetchConfig),
	}
}

// Parameters returns the tool parameter schema
func (t *ScrapeTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"url": {
				Type:        "string",
				Description: "The URL of the web page to scrape",
			},
			"selector": {
				Type:        "string",
				Description: "CSS selector to extract elements (e.g., 'div.content', 'h1', 'a[href]')",
			},
			"extract": {
				Type:        "string",
				Description: "What to extract: 'text' (default), 'html', or an attribute name (e.g., 'href', 'src')",
			},
			"all": {
				Type:        "boolean",
				Description: "If true, extract all matching elements. If false, extract only the first match (default: false)",
			},
			"headers": {
				Type:        "object",
				Description: "Optional HTTP headers to include in the request",
			},
		},
		Required: []string{"url", "selector"},
	}
}

// Execute performs web scraping
func (t *ScrapeTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Rate limiting
	if t.config.RateLimit > 0 {
		elapsed := time.Since(t.lastRequest)
		if elapsed < t.config.RateLimit {
			time.Sleep(t.config.RateLimit - elapsed)
		}
	}
	t.lastRequest = time.Now()

	// Extract and validate parameters
	urlStr, selector, extract, extractAll, headers, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Fetch the web page using fetch tool
	fetchParams := map[string]interface{}{
		"url": urlStr,
	}
	if headers != nil && len(headers) > 0 {
		fetchParams["headers"] = headers
	}

	fetchResult, err := t.fetchTool.Execute(ctx, fetchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	// Extract HTML body from fetch result
	resultMap, ok := fetchResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected fetch result format")
	}

	htmlBody, ok := resultMap["body"].(string)
	if !ok {
		return nil, fmt.Errorf("no HTML body in fetch result")
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract content based on selector
	results := []string{}
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		var value string

		switch extract {
		case "text", "":
			value = strings.TrimSpace(s.Text())
		case "html":
			html, _ := s.Html()
			value = strings.TrimSpace(html)
		default:
			// Extract attribute
			value, _ = s.Attr(extract)
		}

		if value != "" {
			results = append(results, value)
		}

		// If not extracting all, stop after first match
		if !extractAll && len(results) > 0 {
			return
		}
	})

	// Build result
	return t.buildResult(urlStr, selector, extract, extractAll, results), nil
}

// extractParams extracts and validates parameters
func (t *ScrapeTool) extractParams(params map[string]interface{}) (string, string, string, bool, map[string]interface{}, error) {
	// Extract URL
	urlVal, ok := params["url"]
	if !ok {
		return "", "", "", false, nil, fmt.Errorf("url parameter is required")
	}

	urlStr, ok := urlVal.(string)
	if !ok {
		return "", "", "", false, nil, fmt.Errorf("url must be a string")
	}

	if urlStr == "" {
		return "", "", "", false, nil, fmt.Errorf("url cannot be empty")
	}

	// Extract selector
	selectorVal, ok := params["selector"]
	if !ok {
		return "", "", "", false, nil, fmt.Errorf("selector parameter is required")
	}

	selector, ok := selectorVal.(string)
	if !ok {
		return "", "", "", false, nil, fmt.Errorf("selector must be a string")
	}

	if selector == "" {
		return "", "", "", false, nil, fmt.Errorf("selector cannot be empty")
	}

	// Extract 'extract' (optional, defaults to 'text')
	extract := "text"
	if extractVal, ok := params["extract"]; ok && extractVal != nil {
		extract, ok = extractVal.(string)
		if !ok {
			return "", "", "", false, nil, fmt.Errorf("extract must be a string")
		}
	}

	// Extract 'all' (optional, defaults to false)
	extractAll := false
	if allVal, ok := params["all"]; ok && allVal != nil {
		extractAll, ok = allVal.(bool)
		if !ok {
			return "", "", "", false, nil, fmt.Errorf("all must be a boolean")
		}
	}

	// Extract headers (optional)
	var headers map[string]interface{}
	if headersVal, ok := params["headers"]; ok && headersVal != nil {
		headers, ok = headersVal.(map[string]interface{})
		if !ok {
			return "", "", "", false, nil, fmt.Errorf("headers must be an object")
		}
	}

	return urlStr, selector, extract, extractAll, headers, nil
}

// buildResult creates the result object
func (t *ScrapeTool) buildResult(url, selector, extract string, extractAll bool, results []string) map[string]interface{} {
	result := map[string]interface{}{
		"success":  true,
		"url":      url,
		"selector": selector,
		"extract":  extract,
		"count":    len(results),
	}

	if extractAll || len(results) > 1 {
		result["results"] = results
	} else if len(results) == 1 {
		result["result"] = results[0]
	} else {
		result["result"] = nil
		result["results"] = []string{}
	}

	return result
}
