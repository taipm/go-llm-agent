package tools

import (
	"context"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// Tool is the interface that all built-in tools must implement
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Description returns a human-readable description of what the tool does
	Description() string

	// Parameters returns the JSON schema defining the tool's input parameters
	Parameters() *types.JSONSchema

	// Execute runs the tool with the given parameters
	// Returns the tool's output or an error if execution fails
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)

	// Category returns the category this tool belongs to
	Category() ToolCategory

	// RequiresAuth returns true if this tool requires authentication/authorization
	RequiresAuth() bool

	// IsSafe returns true if this tool is safe to use without restrictions
	// Unsafe tools (file delete, system exec) may require additional confirmation
	IsSafe() bool
}

// ToolCategory represents the functional category of a tool
type ToolCategory string

const (
	// CategoryFile represents tools that operate on files and directories
	CategoryFile ToolCategory = "file"

	// CategoryWeb represents tools that make HTTP requests or scrape web content
	CategoryWeb ToolCategory = "web"

	// CategorySystem represents tools that interact with the operating system
	CategorySystem ToolCategory = "system"

	// CategoryData represents tools that process or transform data
	CategoryData ToolCategory = "data"

	// CategoryMath represents tools that perform mathematical operations
	CategoryMath ToolCategory = "math"

	// CategoryDateTime represents tools that work with dates and times
	CategoryDateTime ToolCategory = "datetime"

	// CategoryDatabase represents tools that interact with databases
	CategoryDatabase ToolCategory = "database"
)

// BaseTool provides common functionality for all tools
// Tools can embed this struct to inherit default implementations
type BaseTool struct {
	name         string
	description  string
	category     ToolCategory
	requiresAuth bool
	isSafe       bool
}

// NewBaseTool creates a new BaseTool with the given properties
func NewBaseTool(name, description string, category ToolCategory, requiresAuth, isSafe bool) BaseTool {
	return BaseTool{
		name:         name,
		description:  description,
		category:     category,
		requiresAuth: requiresAuth,
		isSafe:       isSafe,
	}
}

// Name implements Tool.Name
func (b *BaseTool) Name() string {
	return b.name
}

// Description implements Tool.Description
func (b *BaseTool) Description() string {
	return b.description
}

// Category implements Tool.Category
func (b *BaseTool) Category() ToolCategory {
	return b.category
}

// RequiresAuth implements Tool.RequiresAuth
func (b *BaseTool) RequiresAuth() bool {
	return b.requiresAuth
}

// IsSafe implements Tool.IsSafe
func (b *BaseTool) IsSafe() bool {
	return b.isSafe
}

// ToToolDefinition converts a Tool to a types.ToolDefinition for use with LLM providers
func ToToolDefinition(tool Tool) types.ToolDefinition {
	return types.ToolDefinition{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		},
	}
}

// ToToolDefinitions converts multiple tools to tool definitions
func ToToolDefinitions(tools []Tool) []types.ToolDefinition {
	definitions := make([]types.ToolDefinition, len(tools))
	for i, tool := range tools {
		definitions[i] = ToToolDefinition(tool)
	}
	return definitions
}
