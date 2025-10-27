package system

import (
	"context"
	"runtime"
	"testing"
)

func TestProcessesTool_Metadata(t *testing.T) {
	tool := NewProcessesTool()

	if tool.Name() != "system_processes" {
		t.Errorf("Expected name 'system_processes', got '%s'", tool.Name())
	}

	if tool.Category() != "system" {
		t.Errorf("Expected category 'system', got '%s'", tool.Category())
	}

	if !tool.IsSafe() {
		t.Error("Expected system_processes to be safe (read-only)")
	}

	if tool.RequiresAuth() {
		t.Error("Expected system_processes to not require auth")
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestProcessesTool_Parameters(t *testing.T) {
	tool := NewProcessesTool()
	params := tool.Parameters()

	if params == nil {
		t.Fatal("Parameters returned nil")
	}

	if params.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", params.Type)
	}

	// Check parameters exist
	expectedParams := []string{"limit", "sort_by", "name_filter", "min_cpu", "min_memory"}
	for _, param := range expectedParams {
		if _, ok := params.Properties[param]; !ok {
			t.Errorf("Expected parameter '%s'", param)
		}
	}
}

func TestProcessesTool_Execute(t *testing.T) {
	tool := NewProcessesTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"limit": 10,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"type", "count", "total", "processes"}
	for _, field := range requiredFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}

	// Verify type
	if resultMap["type"] != "processes" {
		t.Errorf("Expected type 'processes', got '%v'", resultMap["type"])
	}

	// Verify processes array
	processes, ok := resultMap["processes"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected processes to be array")
	}

	// Should have some processes
	if len(processes) == 0 {
		t.Error("Expected at least one process")
	}

	// Check first process has required fields
	if len(processes) > 0 {
		proc := processes[0]
		procFields := []string{"pid", "name", "cpu_percent", "memory_mb", "status"}
		for _, field := range procFields {
			if _, ok := proc[field]; !ok {
				t.Errorf("Expected field '%s' in process", field)
			}
		}
	}
}

func TestProcessesTool_WithLimit(t *testing.T) {
	tool := NewProcessesTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"limit": 5,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	processes := resultMap["processes"].([]map[string]interface{})

	if len(processes) > 5 {
		t.Errorf("Expected at most 5 processes, got %d", len(processes))
	}
}

func TestProcessesTool_SortByCPU(t *testing.T) {
	tool := NewProcessesTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"limit":   10,
		"sort_by": "cpu",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	processes := resultMap["processes"].([]map[string]interface{})

	// Verify sorted by CPU (descending)
	for i := 1; i < len(processes); i++ {
		cpu1, _ := processes[i-1]["cpu_percent"].(float64)
		cpu2, _ := processes[i]["cpu_percent"].(float64)
		if cpu1 < cpu2 {
			t.Errorf("Processes not sorted by CPU: %f < %f", cpu1, cpu2)
		}
	}
}

func TestProcessesTool_SortByMemory(t *testing.T) {
	tool := NewProcessesTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"limit":   10,
		"sort_by": "memory",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	processes := resultMap["processes"].([]map[string]interface{})

	// Verify sorted by memory (descending)
	for i := 1; i < len(processes); i++ {
		mem1, _ := processes[i-1]["memory_mb"].(float64)
		mem2, _ := processes[i]["memory_mb"].(float64)
		if mem1 < mem2 {
			t.Errorf("Processes not sorted by memory: %f < %f", mem1, mem2)
		}
	}
}

func TestProcessesTool_NameFilter(t *testing.T) {
	tool := NewProcessesTool()
	ctx := context.Background()

	// Look for a common process (should exist on most systems)
	var searchName string
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		searchName = "sh" // shell processes
	} else {
		searchName = "exe" // Windows executables
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"name_filter": searchName,
		"limit":       10,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	processes := resultMap["processes"].([]map[string]interface{})

	// All returned processes should match the filter
	for _, proc := range processes {
		name, _ := proc["name"].(string)
		if !contains(name, searchName) {
			t.Errorf("Process name '%s' doesn't contain '%s'", name, searchName)
		}
	}
}

func TestAppsTool_Metadata(t *testing.T) {
	tool := NewAppsTool()

	if tool.Name() != "system_apps" {
		t.Errorf("Expected name 'system_apps', got '%s'", tool.Name())
	}

	if tool.Category() != "system" {
		t.Errorf("Expected category 'system', got '%s'", tool.Category())
	}

	if !tool.IsSafe() {
		t.Error("Expected system_apps to be safe (read-only)")
	}

	if tool.RequiresAuth() {
		t.Error("Expected system_apps to not require auth")
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestAppsTool_Parameters(t *testing.T) {
	tool := NewAppsTool()
	params := tool.Parameters()

	if params == nil {
		t.Fatal("Parameters returned nil")
	}

	if params.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", params.Type)
	}

	// Check parameters exist
	expectedParams := []string{"limit", "name_filter", "source"}
	for _, param := range expectedParams {
		if _, ok := params.Properties[param]; !ok {
			t.Errorf("Expected parameter '%s'", param)
		}
	}
}

func TestAppsTool_Execute(t *testing.T) {
	tool := NewAppsTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"limit": 10,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Check required fields
	requiredFields := []string{"type", "count", "platform", "sources", "applications"}
	for _, field := range requiredFields {
		if _, ok := resultMap[field]; !ok {
			t.Errorf("Expected field '%s' in result", field)
		}
	}

	// Verify type
	if resultMap["type"] != "applications" {
		t.Errorf("Expected type 'applications', got '%v'", resultMap["type"])
	}

	// Verify platform
	if resultMap["platform"] != runtime.GOOS {
		t.Errorf("Expected platform '%s', got '%v'", runtime.GOOS, resultMap["platform"])
	}

	// Verify applications array
	apps, ok := resultMap["applications"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected applications to be array")
	}

	// Should have some apps (depends on platform)
	t.Logf("Found %d applications", len(apps))

	// If apps found, check structure
	if len(apps) > 0 {
		app := apps[0]
		if _, ok := app["name"]; !ok {
			t.Error("Expected 'name' field in application")
		}
	}
}

func TestAppsTool_WithLimit(t *testing.T) {
	tool := NewAppsTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"limit": 5,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	apps := resultMap["applications"].([]map[string]interface{})

	if len(apps) > 5 {
		t.Errorf("Expected at most 5 applications, got %d", len(apps))
	}
}

func TestAppsTool_NameFilter(t *testing.T) {
	tool := NewAppsTool()
	ctx := context.Background()

	// Use a generic filter that might match something
	result, err := tool.Execute(ctx, map[string]interface{}{
		"name_filter": "a",
		"limit":       10,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	apps := resultMap["applications"].([]map[string]interface{})

	// All returned apps should contain 'a' in name
	for _, app := range apps {
		name, _ := app["name"].(string)
		if !contains(name, "a") {
			t.Errorf("Application name '%s' doesn't contain 'a'", name)
		}
	}
}

// Helper functions tests
func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"HELLO WORLD", "world", true},
		{"hello world", "WORLD", true},
		{"hello", "bye", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		got := contains(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HELLO", "hello"},
		{"World", "world"},
		{"123ABC", "123abc"},
		{"", ""},
		{"already lowercase", "already lowercase"},
	}

	for _, tt := range tests {
		got := toLower(tt.input)
		if got != tt.want {
			t.Errorf("toLower(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
