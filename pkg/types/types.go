package types

import (
	"context"
	"time"
)

// Role represents the role of a message in a conversation
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a single message in a conversation
type Message struct {
	Role      Role                   `json:"role"`
	Content   string                 `json:"content"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	ToolID    string                 `json:"tool_call_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a request to call a tool/function
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall represents the details of a function to call
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Response represents a response from the LLM
type Response struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Metadata  *Metadata  `json:"metadata,omitempty"`
}

// Metadata contains information about the LLM response
type Metadata struct {
	Model            string `json:"model"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

// ChatOptions contains options for chat completion
type ChatOptions struct {
	Temperature      float64                `json:"temperature,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	Tools            []ToolDefinition       `json:"tools,omitempty"`
	SystemPrompt     string                 `json:"system,omitempty"`
	AdditionalParams map[string]interface{} `json:"-"`
}

// ToolDefinition defines a tool that can be called by the LLM
type ToolDefinition struct {
	Type     string             `json:"type"` // "function"
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition defines the schema for a function
type FunctionDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  *JSONSchema `json:"parameters"`
}

// JSONSchema represents a JSON schema for function parameters
type JSONSchema struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
	Metadata  *Metadata  `json:"metadata,omitempty"`
	Error     error      `json:"-"`
}

// StreamHandler is a callback function for handling streaming chunks
type StreamHandler func(chunk StreamChunk) error

// LLMProvider is the interface for LLM providers (Ollama, OpenAI, etc.)
type LLMProvider interface {
	// Chat sends messages and gets a response
	Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error)

	// Stream sends messages and streams the response via callback
	Stream(ctx context.Context, messages []Message, options *ChatOptions, handler StreamHandler) error
}

// Tool is the interface for tools that can be used by the agent
type Tool interface {
	// Name returns the name of the tool
	Name() string

	// Description returns a description of what the tool does
	Description() string

	// Parameters returns the JSON schema for the tool's parameters
	Parameters() *JSONSchema

	// Execute runs the tool with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Memory is the interface for managing conversation history
type Memory interface {
	// Add adds a message to memory
	Add(message Message) error

	// GetHistory returns recent messages up to the limit
	GetHistory(limit int) ([]Message, error)

	// Clear clears all messages from memory
	Clear() error

	// Size returns the number of messages in memory
	Size() int
}

// AdvancedMemory extends Memory with semantic search and advanced features
type AdvancedMemory interface {
	Memory // Embed basic interface

	// Semantic search using vector embeddings
	SearchSemantic(ctx context.Context, query string, limit int) ([]Message, error)

	// Add message with pre-computed embedding
	AddWithEmbedding(ctx context.Context, message Message, embedding []float32) error

	// Get messages by category
	GetByCategory(ctx context.Context, category MessageCategory, limit int) ([]Message, error)

	// Get most important messages
	GetMostImportant(ctx context.Context, limit int) ([]Message, error)

	// Hybrid search combining keyword and semantic search
	HybridSearch(ctx context.Context, query string, limit int) ([]Message, error)

	// Get memory statistics
	GetStats(ctx context.Context) (*MemoryStats, error)

	// Archive old/unimportant messages
	Archive(ctx context.Context, olderThan time.Duration) error

	// Export memory for backup
	Export(ctx context.Context, path string) error
}
