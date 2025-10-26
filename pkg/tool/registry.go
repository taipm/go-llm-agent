package tool

import (
	"context"
	"fmt"
	"sync"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// Registry manages available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]types.Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]types.Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool types.Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Unregister removes a tool from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool %s not found", name)
	}

	delete(r.tools, name)
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (types.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// List returns all registered tool names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetDefinitions returns tool definitions for LLM
func (r *Registry) GetDefinitions() []types.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definitions := make([]types.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, types.ToolDefinition{
			Type: "function",
			Function: types.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		})
	}
	return definitions
}

// Execute runs a tool with the given parameters
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	// TODO: Add parameter validation against schema
	result, err := tool.Execute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("tool %s execution failed: %w", name, err)
	}

	return result, nil
}

// Size returns the number of registered tools
func (r *Registry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}
