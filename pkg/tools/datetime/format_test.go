package datetime

import (
	"context"
	"testing"
	"time"
)

func TestFormatTool_Metadata(t *testing.T) {
	tool := NewFormatTool()

	if tool.Name() != "datetime_format" {
		t.Errorf("Expected name 'datetime_format', got '%s'", tool.Name())
	}

	if tool.Category() != "datetime" {
		t.Errorf("Expected category 'datetime', got '%s'", tool.Category())
	}

	if !tool.IsSafe() {
		t.Error("Expected tool to be safe")
	}

	if tool.Description() == "" {
		t.Error("Description should not be empty")
	}
}

func TestFormatTool_Parameters(t *testing.T) {
	tool := NewFormatTool()
	schema := tool.Parameters()

	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check required parameters
	if len(schema.Required) != 1 || schema.Required[0] != "datetime" {
		t.Errorf("Expected required parameter 'datetime', got %v", schema.Required)
	}

	// Check properties
	expectedProps := []string{"datetime", "from_format", "from_custom_format", "to_format", "to_custom_format", "from_timezone", "to_timezone"}
	for _, prop := range expectedProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Missing property '%s'", prop)
		}
	}
}

func TestFormatTool_ExtractParams(t *testing.T) {
	tool := NewFormatTool()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid minimal params",
			params:    map[string]interface{}{"datetime": "2024-01-15T10:30:00Z"},
			wantError: false,
		},
		{
			name: "valid full params",
			params: map[string]interface{}{
				"datetime":           "2024-01-15T10:30:00Z",
				"from_format":        "RFC3339",
				"to_format":          "RFC1123",
				"from_timezone":      "UTC",
				"to_timezone":        "America/New_York",
				"from_custom_format": "2006-01-02",
				"to_custom_format":   "01/02/2006",
			},
			wantError: false,
		},
		{
			name:      "missing datetime",
			params:    map[string]interface{}{},
			wantError: true,
			errMsg:    "datetime parameter is required",
		},
		{
			name:      "empty datetime",
			params:    map[string]interface{}{"datetime": ""},
			wantError: true,
			errMsg:    "datetime cannot be empty",
		},
		{
			name:      "invalid datetime type",
			params:    map[string]interface{}{"datetime": 123},
			wantError: true,
			errMsg:    "datetime must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, _, _, _, err := tool.extractParams(tt.params)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFormatTool_RFC3339_to_RFC1123(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":    "2024-01-15T10:30:00Z",
		"from_format": "RFC3339",
		"to_format":   "RFC1123",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	formatted, ok := resultMap["formatted"].(string)
	if !ok {
		t.Fatal("Expected formatted to be a string")
	}

	// The output should be in RFC1123 format
	if formatted != "Mon, 15 Jan 2024 10:30:00 UTC" {
		t.Errorf("Expected 'Mon, 15 Jan 2024 10:30:00 UTC', got '%s'", formatted)
	}
}

func TestFormatTool_UnixTimestamp(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	// Convert Unix timestamp to RFC3339
	// 1705305600 = 2024-01-15 08:00:00 UTC
	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":    "1705305600",
		"from_format": "Unix",
		"to_format":   "RFC3339",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["formatted"].(string)

	// Check that it's a valid RFC3339 format
	if formatted == "" {
		t.Error("Expected formatted to be non-empty")
	}

	// Verify Unix timestamp in result
	unix := resultMap["unix"].(int64)
	if unix != 1705305600 {
		t.Errorf("Expected unix 1705305600, got %d", unix)
	}
}

func TestFormatTool_TimezoneConversion(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":      "2024-01-15T10:00:00Z",
		"from_format":   "RFC3339",
		"from_timezone": "UTC",
		"to_format":     "RFC3339",
		"to_timezone":   "Asia/Tokyo",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["formatted"].(string)

	// Tokyo is UTC+9, so 10:00 UTC should be 19:00 JST
	if formatted != "2024-01-15T19:00:00+09:00" {
		t.Errorf("Expected '2024-01-15T19:00:00+09:00', got '%s'", formatted)
	}
}

func TestFormatTool_CustomFormat(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":          "15/01/2024 10:30",
		"from_format":       "custom",
		"from_custom_format": "02/01/2006 15:04",
		"to_format":         "custom",
		"to_custom_format":  "2006-01-02 15:04:05",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["formatted"].(string)

	if formatted != "2024-01-15 10:30:00" {
		t.Errorf("Expected '2024-01-15 10:30:00', got '%s'", formatted)
	}
}

func TestFormatTool_ToUnixFormats(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	tests := []struct {
		name       string
		toFormat   string
		wantPrefix string // Unix timestamps are long, just check prefix
	}{
		{
			name:       "to Unix",
			toFormat:   "Unix",
			wantPrefix: "17053056",
		},
		{
			name:       "to UnixMilli",
			toFormat:   "UnixMilli",
			wantPrefix: "1705305600",
		},
		{
			name:       "to UnixNano",
			toFormat:   "UnixNano",
			wantPrefix: "1705305600000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, map[string]interface{}{
				"datetime":    "2024-01-15T08:00:00Z",
				"from_format": "RFC3339",
				"to_format":   tt.toFormat,
			})

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			resultMap := result.(map[string]interface{})
			formatted := resultMap["formatted"].(string)

			if len(formatted) < len(tt.wantPrefix) || formatted[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("Expected formatted to start with '%s', got '%s'", tt.wantPrefix, formatted)
			}
		})
	}
}

func TestFormatTool_InvalidTimezone(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":      "2024-01-15T10:00:00Z",
		"from_timezone": "Invalid/Timezone",
	})

	if err == nil {
		t.Error("Expected error for invalid timezone, got nil")
	}
}

func TestFormatTool_InvalidFormat(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":    "2024-01-15T10:00:00Z",
		"from_format": "InvalidFormat",
	})

	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

func TestFormatTool_MissingCustomFormat(t *testing.T) {
	tool := NewFormatTool()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":    "2024-01-15",
		"from_format": "custom",
		// Missing from_custom_format
	})

	if err == nil {
		t.Error("Expected error for missing custom format, got nil")
	}
}

func TestFormatTool_GetGoFormat(t *testing.T) {
	tool := NewFormatTool()

	tests := []struct {
		name    string
		format  string
		want    string
		wantErr bool
	}{
		{"RFC3339", "RFC3339", time.RFC3339, false},
		{"RFC1123", "RFC1123", time.RFC1123, false},
		{"Kitchen", "Kitchen", time.Kitchen, false},
		{"Invalid", "InvalidFormat", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.getGoFormat(tt.format)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("Expected '%s', got '%s'", tt.want, got)
				}
			}
		})
	}
}
