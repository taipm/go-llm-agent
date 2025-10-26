package tool

import (
	"context"
	"testing"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// Mock tool for testing
type mockTool struct {
	name        string
	description string
	executed    bool
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"input": {
				Type:        "string",
				Description: "Test input",
			},
		},
		Required: []string{"input"},
	}
}

func (m *mockTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	m.executed = true
	return map[string]interface{}{"result": "success"}, nil
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	tool1 := &mockTool{name: "tool1", description: "Test tool 1"}
	err := registry.Register(tool1)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	if registry.Size() != 1 {
		t.Errorf("Expected registry size 1, got %d", registry.Size())
	}

	// Try to register duplicate
	err = registry.Register(tool1)
	if err == nil {
		t.Error("Expected error when registering duplicate tool")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	tool1 := &mockTool{name: "tool1", description: "Test tool 1"}

	registry.Register(tool1)

	// Get existing tool
	retrieved, err := registry.Get("tool1")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	if retrieved.Name() != "tool1" {
		t.Errorf("Expected tool name 'tool1', got %s", retrieved.Name())
	}

	// Get non-existing tool
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent tool")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	tool1 := &mockTool{name: "tool1", description: "Test tool 1"}

	registry.Register(tool1)

	err := registry.Unregister("tool1")
	if err != nil {
		t.Fatalf("Failed to unregister tool: %v", err)
	}

	if registry.Size() != 0 {
		t.Errorf("Expected registry size 0, got %d", registry.Size())
	}

	// Try to unregister non-existent tool
	err = registry.Unregister("nonexistent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent tool")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	tool1 := &mockTool{name: "tool1", description: "Test tool 1"}
	tool2 := &mockTool{name: "tool2", description: "Test tool 2"}

	registry.Register(tool1)
	registry.Register(tool2)

	names := registry.List()
	if len(names) != 2 {
		t.Errorf("Expected 2 tool names, got %d", len(names))
	}
}

func TestRegistry_Execute(t *testing.T) {
	registry := NewRegistry()
	tool1 := &mockTool{name: "tool1", description: "Test tool 1"}

	registry.Register(tool1)

	ctx := context.Background()
	params := map[string]interface{}{"input": "test"}

	result, err := registry.Execute(ctx, "tool1", params)
	if err != nil {
		t.Fatalf("Failed to execute tool: %v", err)
	}

	if !tool1.executed {
		t.Error("Tool was not executed")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["result"] != "success" {
		t.Errorf("Expected result 'success', got %v", resultMap["result"])
	}
}

func TestRegistry_GetDefinitions(t *testing.T) {
	registry := NewRegistry()

	tool1 := &mockTool{name: "tool1", description: "Test tool 1"}
	tool2 := &mockTool{name: "tool2", description: "Test tool 2"}

	registry.Register(tool1)
	registry.Register(tool2)

	definitions := registry.GetDefinitions()
	if len(definitions) != 2 {
		t.Errorf("Expected 2 definitions, got %d", len(definitions))
	}

	for _, def := range definitions {
		if def.Type != "function" {
			t.Errorf("Expected type 'function', got %s", def.Type)
		}
	}
}
