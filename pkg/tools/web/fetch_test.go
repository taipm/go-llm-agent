package web

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
)

func TestFetchTool_Metadata(t *testing.T) {
	tool := NewFetchTool(DefaultConfig)

	if tool.Name() != "web_fetch" {
		t.Errorf("expected name web_fetch, got %s", tool.Name())
	}

	if tool.Category() != tools.CategoryWeb {
		t.Errorf("expected category Web, got %s", tool.Category())
	}

	if !tool.IsSafe() {
		t.Error("expected web_fetch to be safe (read-only)")
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestFetchTool_Parameters(t *testing.T) {
	tool := NewFetchTool(DefaultConfig)
	params := tool.Parameters()

	if params == nil {
		t.Fatal("expected parameters to be non-nil")
	}

	if params.Type != "object" {
		t.Error("expected type to be object")
	}

	if params.Properties["url"] == nil {
		t.Error("expected url property")
	}

	if params.Properties["headers"] == nil {
		t.Error("expected headers property")
	}

	if len(params.Required) != 1 || params.Required[0] != "url" {
		t.Errorf("expected only url to be required, got %v", params.Required)
	}
}

func TestFetchTool_ValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid HTTPS URL",
			config:  DefaultConfig,
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			config:  DefaultConfig,
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "reject FTP scheme",
			config:  DefaultConfig,
			url:     "ftp://example.com",
			wantErr: true,
			errMsg:  "only HTTP and HTTPS schemes are allowed",
		},
		{
			name:    "reject file scheme",
			config:  DefaultConfig,
			url:     "file:///etc/passwd",
			wantErr: true,
			errMsg:  "only HTTP and HTTPS schemes are allowed",
		},
		{
			name:    "reject localhost",
			config:  DefaultConfig,
			url:     "http://localhost:8080",
			wantErr: true,
			errMsg:  "requests to localhost are not allowed",
		},
		{
			name:    "reject 127.0.0.1",
			config:  DefaultConfig,
			url:     "http://127.0.0.1:8080",
			wantErr: true,
			errMsg:  "requests to localhost are not allowed",
		},
		{
			name:    "reject ::1",
			config:  DefaultConfig,
			url:     "http://[::1]:8080",
			wantErr: true,
			errMsg:  "requests to localhost are not allowed",
		},
		{
			name:    "reject 0.0.0.0",
			config:  DefaultConfig,
			url:     "http://0.0.0.0:8080",
			wantErr: true,
			errMsg:  "requests to localhost are not allowed",
		},
		{
			name:    "reject empty URL",
			config:  DefaultConfig,
			url:     "",
			wantErr: true,
		},
		{
			name:    "reject URL without host",
			config:  DefaultConfig,
			url:     "http://",
			wantErr: true,
			errMsg:  "URL must have a host",
		},
		{
			name: "reject domain not in whitelist",
			config: Config{
				AllowedDomains: []string{"example.com"},
			},
			url:     "https://forbidden.com",
			wantErr: true,
			errMsg:  "domain forbidden.com is not in allowed domains list",
		},
		{
			name: "allow domain in whitelist",
			config: Config{
				AllowedDomains: []string{"example.com"},
			},
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name: "allow subdomain when parent in whitelist",
			config: Config{
				AllowedDomains: []string{"example.com"},
			},
			url:     "https://api.example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewFetchTool(tt.config)
			_, err := tool.validateURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestFetchTool_ExtractParams(t *testing.T) {
	tool := NewFetchTool(DefaultConfig)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantURL string
		wantErr bool
	}{
		{
			name: "valid params with URL only",
			params: map[string]interface{}{
				"url": "https://example.com",
			},
			wantURL: "https://example.com",
			wantErr: false,
		},
		{
			name: "valid params with URL and headers",
			params: map[string]interface{}{
				"url": "https://example.com",
				"headers": map[string]interface{}{
					"Authorization": "Bearer token",
				},
			},
			wantURL: "https://example.com",
			wantErr: false,
		},
		{
			name:    "missing URL",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "empty URL",
			params: map[string]interface{}{
				"url": "",
			},
			wantErr: true,
		},
		{
			name: "invalid URL type",
			params: map[string]interface{}{
				"url": 123,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, headers, err := tool.extractParams(tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("extractParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if url != tt.wantURL {
					t.Errorf("url = %v, want %v", url, tt.wantURL)
				}
				if headers == nil {
					t.Error("headers should not be nil")
				}
			}
		})
	}
}

func TestFetchTool_IsDomainAllowed(t *testing.T) {
	tests := []struct {
		name           string
		allowedDomains []string
		host           string
		want           bool
	}{
		{
			name:           "exact domain match",
			allowedDomains: []string{"example.com"},
			host:           "example.com",
			want:           true,
		},
		{
			name:           "subdomain match",
			allowedDomains: []string{"example.com"},
			host:           "api.example.com",
			want:           true,
		},
		{
			name:           "deep subdomain match",
			allowedDomains: []string{"example.com"},
			host:           "v1.api.example.com",
			want:           true,
		},
		{
			name:           "domain not in list",
			allowedDomains: []string{"example.com"},
			host:           "forbidden.com",
			want:           false,
		},
		{
			name:           "similar but different domain",
			allowedDomains: []string{"example.com"},
			host:           "notexample.com",
			want:           false,
		},
		{
			name:           "multiple domains in whitelist",
			allowedDomains: []string{"example.com", "test.com"},
			host:           "test.com",
			want:           true,
		},
		{
			name:           "host with port",
			allowedDomains: []string{"example.com"},
			host:           "example.com:8080",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				AllowedDomains: tt.allowedDomains,
			}
			tool := NewFetchTool(config)

			got := tool.isDomainAllowed(tt.host)
			if got != tt.want {
				t.Errorf("isDomainAllowed(%s) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"loopback IPv4", "127.0.0.1", true},
		{"loopback IPv4 range", "127.0.0.2", true},
		{"loopback IPv6", "::1", true},
		{"private 10.x", "10.0.0.1", true},
		{"private 10.x range", "10.255.255.255", true},
		{"private 172.16.x", "172.16.0.1", true},
		{"private 172.16.x range", "172.31.255.255", true},
		{"private 192.168.x", "192.168.0.1", true},
		{"private 192.168.x range", "192.168.255.255", true},
		{"public IP", "8.8.8.8", false},
		{"public IP", "1.1.1.1", false},
		{"public IP", "93.184.216.34", false}, // example.com
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}

			got := isPrivateIP(ip)
			if got != tt.want {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"127.255.255.255", true},
		{"::1", true},
		{"0.0.0.0", true},
		{"0.1.2.3", true},
		{"example.com", false},
		{"192.168.1.1", false},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := isLocalhost(tt.host)
			if got != tt.want {
				t.Errorf("isLocalhost(%s) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestFetchTool_Execute_Validation(t *testing.T) {
	tool := NewFetchTool(DefaultConfig)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing url parameter",
			params:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "url parameter is required",
		},
		{
			name: "empty url",
			params: map[string]interface{}{
				"url": "",
			},
			wantErr: true,
			errMsg:  "url parameter is required",
		},
		{
			name: "invalid url type",
			params: map[string]interface{}{
				"url": 123,
			},
			wantErr: true,
			errMsg:  "url parameter is required",
		},
		{
			name: "localhost rejected",
			params: map[string]interface{}{
				"url": "http://localhost:8080",
			},
			wantErr: true,
			errMsg:  "localhost",
		},
		{
			name: "invalid scheme",
			params: map[string]interface{}{
				"url": "ftp://example.com",
			},
			wantErr: true,
			errMsg:  "only HTTP and HTTPS schemes are allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_, err := tool.Execute(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestFetchTool_Config(t *testing.T) {
	config := Config{
		Timeout:          5 * time.Second,
		MaxResponseSize:  1024,
		AllowedDomains:   []string{"example.com"},
		FollowRedirects:  false,
		MaxRedirects:     5,
		UserAgent:        "test-agent",
		AllowPrivateIPs:  true,
	}

	tool := NewFetchTool(config)

	if tool.config.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", tool.config.Timeout, 5*time.Second)
	}

	if tool.config.MaxResponseSize != 1024 {
		t.Errorf("MaxResponseSize = %v, want %v", tool.config.MaxResponseSize, 1024)
	}

	if len(tool.config.AllowedDomains) != 1 {
		t.Errorf("AllowedDomains length = %v, want %v", len(tool.config.AllowedDomains), 1)
	}

	if tool.config.UserAgent != "test-agent" {
		t.Errorf("UserAgent = %v, want %v", tool.config.UserAgent, "test-agent")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
