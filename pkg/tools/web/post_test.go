package web

import (
	"context"
	"testing"
)

func TestPostTool_Metadata(t *testing.T) {
	tool := NewPostTool(PostConfig{})

	if tool.Name() != "web_post" {
		t.Errorf("Expected name 'web_post', got '%s'", tool.Name())
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

func TestPostTool_Parameters(t *testing.T) {
	tool := NewPostTool(PostConfig{})
	schema := tool.Parameters()

	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check required parameters
	if len(schema.Required) != 1 || schema.Required[0] != "url" {
		t.Errorf("Expected required parameter 'url', got %v", schema.Required)
	}

	// Check properties
	expectedProps := []string{"url", "body", "headers", "form_data"}
	for _, prop := range expectedProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Missing property '%s'", prop)
		}
	}
}

func TestPostTool_ExtractParams(t *testing.T) {
	tool := NewPostTool(PostConfig{})

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantURL   string
		wantBody  bool
		wantForm  bool
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid params with URL and body",
			params:    map[string]interface{}{"url": "https://example.com", "body": map[string]interface{}{"key": "value"}},
			wantURL:   "https://example.com",
			wantBody:  true,
			wantForm:  false,
			wantError: false,
		},
		{
			name:      "valid params with URL and form_data",
			params:    map[string]interface{}{"url": "https://example.com", "form_data": map[string]interface{}{"key": "value"}},
			wantURL:   "https://example.com",
			wantBody:  false,
			wantForm:  true,
			wantError: false,
		},
		{
			name:      "valid params with URL only",
			params:    map[string]interface{}{"url": "https://example.com"},
			wantURL:   "https://example.com",
			wantBody:  false,
			wantForm:  false,
			wantError: false,
		},
		{
			name:      "missing URL",
			params:    map[string]interface{}{},
			wantError: true,
			errMsg:    "url parameter is required",
		},
		{
			name:      "empty URL",
			params:    map[string]interface{}{"url": ""},
			wantError: true,
			errMsg:    "url cannot be empty",
		},
		{
			name:      "invalid URL type",
			params:    map[string]interface{}{"url": 123},
			wantError: true,
			errMsg:    "url must be a string",
		},
		{
			name:      "invalid body type",
			params:    map[string]interface{}{"url": "https://example.com", "body": "not an object"},
			wantError: true,
			errMsg:    "body must be an object",
		},
		{
			name:      "invalid form_data type",
			params:    map[string]interface{}{"url": "https://example.com", "form_data": "not an object"},
			wantError: true,
			errMsg:    "form_data must be an object",
		},
		{
			name:      "invalid headers type",
			params:    map[string]interface{}{"url": "https://example.com", "headers": "not an object"},
			wantError: true,
			errMsg:    "headers must be an object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr, body, formData, _, err := tool.extractParams(tt.params)

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

			if tt.wantBody && body == nil {
				t.Error("Expected body to be set")
			}
			if !tt.wantBody && body != nil {
				t.Error("Expected body to be nil")
			}

			if tt.wantForm && formData == nil {
				t.Error("Expected form_data to be set")
			}
			if !tt.wantForm && formData != nil {
				t.Error("Expected form_data to be nil")
			}
		})
	}
}

func TestPostTool_Execute_Validation(t *testing.T) {
	tool := NewPostTool(PostConfig{})
	ctx := context.Background()

	tests := []struct {
		name   string
		params map[string]interface{}
		errMsg string
	}{
		{
			name:   "missing url parameter",
			params: map[string]interface{}{},
			errMsg: "url parameter is required",
		},
		{
			name:   "empty url",
			params: map[string]interface{}{"url": ""},
			errMsg: "url cannot be empty",
		},
		{
			name:   "invalid url type",
			params: map[string]interface{}{"url": 123},
			errMsg: "url must be a string",
		},
		{
			name:   "localhost rejected",
			params: map[string]interface{}{"url": "http://localhost:8080"},
			errMsg: "requests to localhost are not allowed",
		},
		{
			name:   "invalid scheme",
			params: map[string]interface{}{"url": "ftp://example.com"},
			errMsg: "only HTTP and HTTPS schemes are allowed, got: ftp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Execute(ctx, tt.params)
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				return
			}

			if err.Error() != tt.errMsg {
				t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestPostTool_Config(t *testing.T) {
	config := PostConfig{
		Timeout:         60,
		MaxResponseSize: 2048,
		AllowedDomains:  []string{"example.com"},
		UserAgent:       "test-agent",
		AllowPrivateIPs: true,
	}

	tool := NewPostTool(config)

	if tool.config.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %v", tool.config.Timeout)
	}

	if tool.config.MaxResponseSize != 2048 {
		t.Errorf("Expected max response size 2048, got %d", tool.config.MaxResponseSize)
	}

	if len(tool.config.AllowedDomains) != 1 || tool.config.AllowedDomains[0] != "example.com" {
		t.Errorf("Expected allowed domains ['example.com'], got %v", tool.config.AllowedDomains)
	}

	if tool.config.UserAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got '%s'", tool.config.UserAgent)
	}

	if !tool.config.AllowPrivateIPs {
		t.Error("Expected AllowPrivateIPs to be true")
	}
}
