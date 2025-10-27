package web

import (
	"testing"
)

func TestScrapeTool_Metadata(t *testing.T) {
	tool := NewScrapeTool(ScrapeConfig{})

	if tool.Name() != "web_scrape" {
		t.Errorf("Expected name 'web_scrape', got '%s'", tool.Name())
	}

	if tool.Category() != "web" {
		t.Errorf("Expected category 'web', got '%s'", tool.Category())
	}

	if !tool.IsSafe() {
		t.Error("Expected tool to be safe")
	}

	if tool.Description() == "" {
		t.Error("Description should not be empty")
	}
}

func TestScrapeTool_Parameters(t *testing.T) {
	tool := NewScrapeTool(ScrapeConfig{})
	schema := tool.Parameters()

	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check required parameters
	if len(schema.Required) != 2 || schema.Required[0] != "url" || schema.Required[1] != "selector" {
		t.Errorf("Expected required parameters ['url', 'selector'], got %v", schema.Required)
	}

	// Check properties
	expectedProps := []string{"url", "selector", "extract", "all", "headers"}
	for _, prop := range expectedProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Missing property '%s'", prop)
		}
	}
}

func TestScrapeTool_ExtractParams(t *testing.T) {
	tool := NewScrapeTool(ScrapeConfig{})

	tests := []struct {
		name         string
		params       map[string]interface{}
		wantURL      string
		wantSelector string
		wantExtract  string
		wantAll      bool
		wantError    bool
		errMsg       string
	}{
		{
			name:         "valid params - minimal",
			params:       map[string]interface{}{"url": "https://example.com", "selector": "div.content"},
			wantURL:      "https://example.com",
			wantSelector: "div.content",
			wantExtract:  "text",
			wantAll:      false,
			wantError:    false,
		},
		{
			name:         "valid params - full",
			params:       map[string]interface{}{"url": "https://example.com", "selector": "a", "extract": "href", "all": true},
			wantURL:      "https://example.com",
			wantSelector: "a",
			wantExtract:  "href",
			wantAll:      true,
			wantError:    false,
		},
		{
			name:      "missing url",
			params:    map[string]interface{}{"selector": "div"},
			wantError: true,
			errMsg:    "url parameter is required",
		},
		{
			name:      "empty url",
			params:    map[string]interface{}{"url": "", "selector": "div"},
			wantError: true,
			errMsg:    "url cannot be empty",
		},
		{
			name:      "invalid url type",
			params:    map[string]interface{}{"url": 123, "selector": "div"},
			wantError: true,
			errMsg:    "url must be a string",
		},
		{
			name:      "missing selector",
			params:    map[string]interface{}{"url": "https://example.com"},
			wantError: true,
			errMsg:    "selector parameter is required",
		},
		{
			name:      "empty selector",
			params:    map[string]interface{}{"url": "https://example.com", "selector": ""},
			wantError: true,
			errMsg:    "selector cannot be empty",
		},
		{
			name:      "invalid selector type",
			params:    map[string]interface{}{"url": "https://example.com", "selector": 123},
			wantError: true,
			errMsg:    "selector must be a string",
		},
		{
			name:      "invalid extract type",
			params:    map[string]interface{}{"url": "https://example.com", "selector": "div", "extract": 123},
			wantError: true,
			errMsg:    "extract must be a string",
		},
		{
			name:      "invalid all type",
			params:    map[string]interface{}{"url": "https://example.com", "selector": "div", "all": "yes"},
			wantError: true,
			errMsg:    "all must be a boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr, selector, extract, extractAll, _, err := tool.extractParams(tt.params)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if urlStr != tt.wantURL {
				t.Errorf("Expected URL '%s', got '%s'", tt.wantURL, urlStr)
			}

			if selector != tt.wantSelector {
				t.Errorf("Expected selector '%s', got '%s'", tt.wantSelector, selector)
			}

			if extract != tt.wantExtract {
				t.Errorf("Expected extract '%s', got '%s'", tt.wantExtract, extract)
			}

			if extractAll != tt.wantAll {
				t.Errorf("Expected all '%v', got '%v'", tt.wantAll, extractAll)
			}
		})
	}
}

func TestScrapeTool_BuildResult(t *testing.T) {
	tool := NewScrapeTool(ScrapeConfig{})

	tests := []struct {
		name        string
		url         string
		selector    string
		extract     string
		extractAll  bool
		results     []string
		wantSuccess bool
		wantCount   int
		checkResult bool
		wantResult  string
	}{
		{
			name:        "single result - extract first",
			url:         "https://example.com",
			selector:    "h1",
			extract:     "text",
			extractAll:  false,
			results:     []string{"Title"},
			wantSuccess: true,
			wantCount:   1,
			checkResult: true,
			wantResult:  "Title",
		},
		{
			name:        "multiple results - extract all",
			url:         "https://example.com",
			selector:    "a",
			extract:     "href",
			extractAll:  true,
			results:     []string{"link1", "link2", "link3"},
			wantSuccess: true,
			wantCount:   3,
			checkResult: false,
		},
		{
			name:        "no results",
			url:         "https://example.com",
			selector:    "div.notfound",
			extract:     "text",
			extractAll:  false,
			results:     []string{},
			wantSuccess: true,
			wantCount:   0,
			checkResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.buildResult(tt.url, tt.selector, tt.extract, tt.extractAll, tt.results)

			if success, ok := result["success"].(bool); !ok || success != tt.wantSuccess {
				t.Errorf("Expected success=%v, got %v", tt.wantSuccess, result["success"])
			}

			if count, ok := result["count"].(int); !ok || count != tt.wantCount {
				t.Errorf("Expected count=%d, got %v", tt.wantCount, result["count"])
			}

			if tt.checkResult {
				if resultStr, ok := result["result"].(string); !ok || resultStr != tt.wantResult {
					t.Errorf("Expected result='%s', got %v", tt.wantResult, result["result"])
				}
			}
		})
	}
}

func TestScrapeTool_Config(t *testing.T) {
	config := ScrapeConfig{
		Timeout:         60,
		MaxResponseSize: 10 * 1024 * 1024,
		AllowedDomains:  []string{"example.com"},
		UserAgent:       "test-scraper",
		AllowPrivateIPs: true,
		RateLimit:       2,
	}

	tool := NewScrapeTool(config)

	if tool.config.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %v", tool.config.Timeout)
	}

	if tool.config.MaxResponseSize != 10*1024*1024 {
		t.Errorf("Expected max response size 10485760, got %d", tool.config.MaxResponseSize)
	}

	if len(tool.config.AllowedDomains) != 1 || tool.config.AllowedDomains[0] != "example.com" {
		t.Errorf("Expected allowed domains ['example.com'], got %v", tool.config.AllowedDomains)
	}

	if tool.config.UserAgent != "test-scraper" {
		t.Errorf("Expected user agent 'test-scraper', got '%s'", tool.config.UserAgent)
	}

	if !tool.config.AllowPrivateIPs {
		t.Error("Expected AllowPrivateIPs to be true")
	}

	if tool.config.RateLimit != 2 {
		t.Errorf("Expected rate limit 2, got %v", tool.config.RateLimit)
	}
}
