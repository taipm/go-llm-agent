package system

import (
	"context"
	"testing"
)

func TestInfoTool_Metadata(t *testing.T) {
	tool := NewInfoTool()

	if tool.Name() != "system_info" {
		t.Errorf("Expected name 'system_info', got '%s'", tool.Name())
	}

	if tool.Category() != "system" {
		t.Errorf("Expected category 'system', got '%s'", tool.Category())
	}

	if !tool.IsSafe() {
		t.Error("Expected system_info to be safe (read-only)")
	}

	if tool.RequiresAuth() {
		t.Error("Expected system_info to not require auth")
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestInfoTool_Parameters(t *testing.T) {
	tool := NewInfoTool()
	params := tool.Parameters()

	if params == nil {
		t.Fatal("Parameters returned nil")
	}

	if params.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", params.Type)
	}

	// Check 'type' parameter exists
	typeParam, ok := params.Properties["type"]
	if !ok {
		t.Fatal("Expected 'type' parameter")
	}

	if typeParam.Type != "string" {
		t.Errorf("Expected type parameter to be string, got '%s'", typeParam.Type)
	}

	// Verify enum values
	expectedEnum := []interface{}{"cpu", "memory", "disk", "os", "network", "all"}
	if len(typeParam.Enum) != len(expectedEnum) {
		t.Errorf("Expected %d enum values, got %d", len(expectedEnum), len(typeParam.Enum))
	}

	// Check 'path' parameter exists (optional)
	pathParam, ok := params.Properties["path"]
	if !ok {
		t.Fatal("Expected 'path' parameter")
	}

	if pathParam.Type != "string" {
		t.Errorf("Expected path parameter to be string, got '%s'", pathParam.Type)
	}

	// Verify required fields
	if len(params.Required) != 1 || params.Required[0] != "type" {
		t.Errorf("Expected required: ['type'], got %v", params.Required)
	}
}

func TestInfoTool_ExtractType(t *testing.T) {
	tool := NewInfoTool()

	tests := []struct {
		name      string
		params    map[string]interface{}
		expected  string
		shouldErr bool
	}{
		{
			name:      "valid cpu type",
			params:    map[string]interface{}{"type": "cpu"},
			expected:  "cpu",
			shouldErr: false,
		},
		{
			name:      "valid memory type",
			params:    map[string]interface{}{"type": "memory"},
			expected:  "memory",
			shouldErr: false,
		},
		{
			name:      "valid disk type",
			params:    map[string]interface{}{"type": "disk"},
			expected:  "disk",
			shouldErr: false,
		},
		{
			name:      "missing type",
			params:    map[string]interface{}{},
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "empty type",
			params:    map[string]interface{}{"type": ""},
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "invalid type - not string",
			params:    map[string]interface{}{"type": 123},
			expected:  "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.extractType(tt.params)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestInfoTool_GetCPUInfo(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "cpu",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"type", "logical_cores", "usage_percent"}
	for _, field := range requiredFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}

	// Verify type field
	if resultMap["type"] != "cpu" {
		t.Errorf("Expected type 'cpu', got '%v'", resultMap["type"])
	}

	// Verify logical_cores is a positive number
	if cores, ok := resultMap["logical_cores"].(int); ok {
		if cores <= 0 {
			t.Errorf("Expected positive logical_cores, got %d", cores)
		}
	}
}

func TestInfoTool_GetMemoryInfo(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "memory",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"type", "total_bytes", "available_bytes", "used_bytes", "usage_percent"}
	for _, field := range requiredFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}

	// Verify type field
	if resultMap["type"] != "memory" {
		t.Errorf("Expected type 'memory', got '%v'", resultMap["type"])
	}

	// Verify total_bytes is positive
	if total, ok := resultMap["total_bytes"].(uint64); ok {
		if total == 0 {
			t.Error("Expected non-zero total_bytes")
		}
	}

	// Verify GB conversions exist
	gbFields := []string{"total_gb", "available_gb", "used_gb"}
	for _, field := range gbFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}
}

func TestInfoTool_GetDiskInfo(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "disk",
		"path": "/",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"type", "path", "total_bytes", "free_bytes", "used_bytes", "usage_percent"}
	for _, field := range requiredFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}

	// Verify type field
	if resultMap["type"] != "disk" {
		t.Errorf("Expected type 'disk', got '%v'", resultMap["type"])
	}

	// Verify path
	if path, ok := resultMap["path"].(string); ok {
		if path == "" {
			t.Error("Expected non-empty path")
		}
	}
}

func TestInfoTool_GetDiskInfo_DefaultPath(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	// Test without providing path parameter
	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "disk",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Should default to root path
	if _, ok := resultMap["path"]; !ok {
		t.Error("Expected path field in result")
	}
}

func TestInfoTool_GetOSInfo(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "os",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"type", "hostname", "os", "platform", "go_version", "go_os", "go_arch"}
	for _, field := range requiredFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}

	// Verify type field
	if resultMap["type"] != "os" {
		t.Errorf("Expected type 'os', got '%v'", resultMap["type"])
	}

	// Verify hostname is not empty
	if hostname, ok := resultMap["hostname"].(string); ok {
		if hostname == "" {
			t.Error("Expected non-empty hostname")
		}
	}

	// Verify go_version starts with "go"
	if goVer, ok := resultMap["go_version"].(string); ok {
		if len(goVer) < 2 {
			t.Errorf("Expected go_version to be at least 2 chars, got '%s'", goVer)
		}
	}
}

func TestInfoTool_GetNetworkInfo(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "network",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"type", "interfaces", "count"}
	for _, field := range requiredFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}

	// Verify type field
	if resultMap["type"] != "network" {
		t.Errorf("Expected type 'network', got '%v'", resultMap["type"])
	}

	// Verify interfaces is an array
	if interfaces, ok := resultMap["interfaces"].([]map[string]interface{}); ok {
		// Should have at least one interface (loopback)
		if len(interfaces) == 0 {
			t.Error("Expected at least one network interface")
		}

		// Check first interface has required fields
		if len(interfaces) > 0 {
			iface := interfaces[0]
			if _, ok := iface["name"]; !ok {
				t.Error("Expected 'name' field in network interface")
			}
		}
	}
}

func TestInfoTool_GetAllInfo(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "all",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check that all info types are present
	infoTypes := []string{"cpu", "memory", "disk", "os", "network"}
	for _, infoType := range infoTypes {
		if _, ok := resultMap[infoType]; !ok {
			t.Errorf("Expected '%s' in all info result", infoType)
		}
	}

	// Verify type field
	if resultMap["type"] != "all" {
		t.Errorf("Expected type 'all', got '%v'", resultMap["type"])
	}
}

func TestInfoTool_InvalidType(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "invalid",
	})

	if err == nil {
		t.Error("Expected error for invalid type")
	}

	if result != nil {
		t.Error("Expected nil result for invalid type")
	}
}

func TestInfoTool_InvalidDiskPath(t *testing.T) {
	tool := NewInfoTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"type": "disk",
		"path": "/nonexistent/path/that/should/not/exist",
	})

	// Should return an error for non-existent path
	if err == nil {
		t.Error("Expected error for non-existent disk path")
	}

	if result != nil {
		t.Error("Expected nil result for invalid disk path")
	}
}
