package datetime

import (
	"context"
	"testing"
	"time"
)

func TestCalcTool_Metadata(t *testing.T) {
	tool := NewCalcTool()

	if tool.Name() != "datetime_calc" {
		t.Errorf("Expected name 'datetime_calc', got '%s'", tool.Name())
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

func TestCalcTool_Parameters(t *testing.T) {
	tool := NewCalcTool()
	schema := tool.Parameters()

	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check required parameters
	if len(schema.Required) != 2 {
		t.Errorf("Expected 2 required parameters, got %d", len(schema.Required))
	}

	// Check properties
	expectedProps := []string{"datetime", "operation", "duration", "target_datetime", "format", "timezone", "output_format"}
	for _, prop := range expectedProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Missing property '%s'", prop)
		}
	}
}

func TestCalcTool_AddDuration(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":  "2024-01-15T10:00:00Z",
		"operation": "add",
		"duration":  "2h30m",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["result"].(string)

	// Adding 2h30m to 10:00:00 should give 12:30:00
	if formatted != "2024-01-15T12:30:00Z" {
		t.Errorf("Expected '2024-01-15T12:30:00Z', got '%s'", formatted)
	}
}

func TestCalcTool_SubtractDuration(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":  "2024-01-15T10:00:00Z",
		"operation": "subtract",
		"duration":  "1h30m",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["result"].(string)

	// Subtracting 1h30m from 10:00:00 should give 08:30:00
	if formatted != "2024-01-15T08:30:00Z" {
		t.Errorf("Expected '2024-01-15T08:30:00Z', got '%s'", formatted)
	}
}

func TestCalcTool_AddDays(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":  "2024-01-15T10:00:00Z",
		"operation": "add",
		"duration":  "7d",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["result"].(string)

	// Adding 7 days should give 2024-01-22
	if formatted != "2024-01-22T10:00:00Z" {
		t.Errorf("Expected '2024-01-22T10:00:00Z', got '%s'", formatted)
	}
}

func TestCalcTool_DiffDates(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":        "2024-01-20T10:00:00Z",
		"operation":       "diff",
		"target_datetime": "2024-01-15T10:00:00Z",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	diff := resultMap["difference"].(map[string]interface{})

	// Difference should be 5 days = 120 hours
	hours := diff["hours"].(float64)
	if hours != 120.0 {
		t.Errorf("Expected 120 hours, got %f", hours)
	}

	days := diff["days"].(float64)
	if days != 5.0 {
		t.Errorf("Expected 5 days, got %f", days)
	}
}

func TestCalcTool_ParseDuration(t *testing.T) {
	tool := NewCalcTool()

	tests := []struct {
		name     string
		duration string
		want     time.Duration
		wantErr  bool
	}{
		{"hours", "2h", 2 * time.Hour, false},
		{"minutes", "30m", 30 * time.Minute, false},
		{"seconds", "45s", 45 * time.Second, false},
		{"milliseconds", "500ms", 500 * time.Millisecond, false},
		{"combined", "1h30m", 90 * time.Minute, false},
		{"days", "1d", 24 * time.Hour, false},
		{"multiple days", "7d", 7 * 24 * time.Hour, false},
		{"decimal days", "1.5d", 36 * time.Hour, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.parseDuration(tt.duration)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("Expected %v, got %v", tt.want, got)
				}
			}
		})
	}
}

func TestCalcTool_ExtractParams(t *testing.T) {
	tool := NewCalcTool()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "valid add operation",
			params: map[string]interface{}{
				"datetime":  "2024-01-15T10:00:00Z",
				"operation": "add",
				"duration":  "2h",
			},
			wantError: false,
		},
		{
			name: "valid diff operation",
			params: map[string]interface{}{
				"datetime":        "2024-01-20T10:00:00Z",
				"operation":       "diff",
				"target_datetime": "2024-01-15T10:00:00Z",
			},
			wantError: false,
		},
		{
			name:      "missing datetime",
			params:    map[string]interface{}{"operation": "add"},
			wantError: true,
			errMsg:    "datetime parameter is required",
		},
		{
			name:      "missing operation",
			params:    map[string]interface{}{"datetime": "2024-01-15T10:00:00Z"},
			wantError: true,
			errMsg:    "operation parameter is required",
		},
		{
			name:      "empty datetime",
			params:    map[string]interface{}{"datetime": "", "operation": "add"},
			wantError: true,
			errMsg:    "datetime cannot be empty",
		},
		{
			name:      "invalid datetime type",
			params:    map[string]interface{}{"datetime": 123, "operation": "add"},
			wantError: true,
			errMsg:    "datetime must be a string",
		},
		{
			name:      "invalid operation type",
			params:    map[string]interface{}{"datetime": "2024-01-15T10:00:00Z", "operation": 123},
			wantError: true,
			errMsg:    "operation must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, _, _, _, _, _, err := tool.extractParams(tt.params)

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

func TestCalcTool_MissingDurationForAdd(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":  "2024-01-15T10:00:00Z",
		"operation": "add",
		// Missing duration
	})

	if err == nil {
		t.Error("Expected error for missing duration, got nil")
	}
}

func TestCalcTool_MissingTargetForDiff(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":  "2024-01-15T10:00:00Z",
		"operation": "diff",
		// Missing target_datetime
	})

	if err == nil {
		t.Error("Expected error for missing target_datetime, got nil")
	}
}

func TestCalcTool_InvalidOperation(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":  "2024-01-15T10:00:00Z",
		"operation": "invalid",
	})

	if err == nil {
		t.Error("Expected error for invalid operation, got nil")
	}
}

func TestCalcTool_WithTimezone(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":  "2024-01-15T10:00:00+09:00",
		"operation": "add",
		"duration":  "1h",
		"format":    "RFC3339",
		"timezone":  "Asia/Tokyo",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["result"].(string)

	// Should maintain Tokyo timezone (+09:00)
	if formatted != "2024-01-15T11:00:00+09:00" {
		t.Errorf("Expected '2024-01-15T11:00:00+09:00', got '%s'", formatted)
	}
}

func TestCalcTool_CustomOutputFormat(t *testing.T) {
	tool := NewCalcTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"datetime":      "2024-01-15T10:00:00Z",
		"operation":     "add",
		"duration":      "2h",
		"output_format": "RFC1123",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	formatted := resultMap["result"].(string)

	// Should be in RFC1123 format
	if formatted != "Mon, 15 Jan 2024 12:00:00 UTC" {
		t.Errorf("Expected 'Mon, 15 Jan 2024 12:00:00 UTC', got '%s'", formatted)
	}
}
