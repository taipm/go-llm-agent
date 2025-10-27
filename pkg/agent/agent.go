package agent

import (
	"context"
	"fmt"

	"github.com/taipm/go-llm-agent/pkg/builtin"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/tool"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// Agent orchestrates LLM, tools, and memory
type Agent struct {
	provider types.LLMProvider
	tools    *tool.Registry
	memory   types.Memory
	options  *Options
	logger   Logger
}

// Options contains configuration for the agent
type Options struct {
	SystemPrompt  string
	Temperature   float64
	MaxTokens     int
	MaxIterations int // Maximum tool calling iterations
}

// DefaultOptions returns default agent options
func DefaultOptions() *Options {
	return &Options{
		SystemPrompt:  "You are a helpful AI assistant.",
		Temperature:   0.7,
		MaxTokens:     2000,
		MaxIterations: 10,
	}
}

// New creates a new agent with default memory (100 messages) and all builtin tools
func New(provider types.LLMProvider, opts ...Option) *Agent {
	agent := &Agent{
		provider: provider,
		tools:    tool.NewRegistry(),
		memory:   memory.NewBuffer(100), // Default memory with 100 messages
		options:  DefaultOptions(),
		logger:   NewConsoleLogger(), // Default logger
	}

	// Load all builtin tools by default
	registry := builtin.GetRegistry()
	for _, tool := range registry.All() {
		agent.tools.Register(tool)
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// Option is a function that configures the agent
type Option func(*Agent)

// WithMemory sets the memory for the agent
func WithMemory(memory types.Memory) Option {
	return func(a *Agent) {
		a.memory = memory
	}
}

// WithSystemPrompt sets the system prompt
func WithSystemPrompt(prompt string) Option {
	return func(a *Agent) {
		a.options.SystemPrompt = prompt
	}
}

// WithTemperature sets the temperature
func WithTemperature(temp float64) Option {
	return func(a *Agent) {
		a.options.Temperature = temp
	}
}

// WithMaxTokens sets the max tokens
func WithMaxTokens(max int) Option {
	return func(a *Agent) {
		a.options.MaxTokens = max
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) Option {
	return func(a *Agent) {
		a.logger = logger
	}
}

// WithLogLevel sets the log level
func WithLogLevel(level LogLevel) Option {
	return func(a *Agent) {
		a.logger.SetLevel(level)
	}
}

// DisableLogging disables all logging
func DisableLogging() Option {
	return func(a *Agent) {
		a.logger = &NoopLogger{}
	}
}

// WithoutBuiltinTools disables automatic loading of builtin tools
func WithoutBuiltinTools() Option {
	return func(a *Agent) {
		// Clear the tools registry
		a.tools = tool.NewRegistry()
	}
}

// AddTool registers a tool with the agent
func (a *Agent) AddTool(t types.Tool) error {
	return a.tools.Register(t)
}

// RemoveTool unregisters a tool
func (a *Agent) RemoveTool(name string) error {
	return a.tools.Unregister(name)
}

// ToolCount returns the number of registered tools
func (a *Agent) ToolCount() int {
	return a.tools.Size()
}

// Chat sends a message and returns the response
func (a *Agent) Chat(ctx context.Context, message string) (string, error) {
	// Log user message
	LogUserMessage(a.logger, message)

	// Add user message to memory if available
	userMsg := types.Message{
		Role:    types.RoleUser,
		Content: message,
	}

	if a.memory != nil {
		if err := a.memory.Add(userMsg); err != nil {
			return "", fmt.Errorf("failed to add message to memory: %w", err)
		}
		a.logger.Debug("💾 Saved user message to memory")
	}

	// Get conversation history
	var messages []types.Message
	if a.memory != nil {
		history, err := a.memory.GetHistory(0) // Get all
		if err != nil {
			return "", fmt.Errorf("failed to get history: %w", err)
		}
		messages = history
		a.logger.Debug("💾 Retrieved %d messages from memory", len(messages))
	} else {
		messages = []types.Message{userMsg}
	}

	// Prepare chat options
	chatOpts := &types.ChatOptions{
		SystemPrompt: a.options.SystemPrompt,
		Temperature:  a.options.Temperature,
		MaxTokens:    a.options.MaxTokens,
	}

	// Add tools if available
	if a.tools.Size() > 0 {
		chatOpts.Tools = a.tools.GetDefinitions()
	}

	// Run agent loop with tool calling
	response, err := a.runLoop(ctx, messages, chatOpts)
	if err != nil {
		a.logger.Error("Agent execution failed: %v", err)
		return "", err
	}

	// Log final response
	LogResponse(a.logger, response)

	// Note: runLoop already saves the final response to memory
	return response, nil
}

// runLoop executes the agent loop with tool calling
func (a *Agent) runLoop(ctx context.Context, messages []types.Message, opts *types.ChatOptions) (string, error) {
	currentMessages := make([]types.Message, len(messages))
	copy(currentMessages, messages)

	for iteration := 0; iteration < a.options.MaxIterations; iteration++ {
		LogIteration(a.logger, iteration, a.options.MaxIterations)

		// Log thinking
		LogThinking(a.logger)

		// Call LLM
		response, err := a.provider.Chat(ctx, currentMessages, opts)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		// If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			a.logger.Debug("No tool calls, returning response")

			// Save final assistant response to memory
			if a.memory != nil {
				finalMsg := types.Message{
					Role:    types.RoleAssistant,
					Content: response.Content,
				}
				if err := a.memory.Add(finalMsg); err != nil {
					return "", fmt.Errorf("failed to add final response to memory: %w", err)
				}
				a.logger.Debug("💾 Saved assistant response to memory")
			}
			return response.Content, nil
		}

		// Log tool calls
		a.logger.Info("🔧 Agent wants to call %d tool(s): %s", len(response.ToolCalls), FormatToolCalls(response.ToolCalls))

		// Execute tool calls
		assistantMsg := types.Message{
			Role:      types.RoleAssistant,
			Content:   response.Content,
			ToolCalls: response.ToolCalls,
		}
		currentMessages = append(currentMessages, assistantMsg)

		// Save assistant message to memory
		if a.memory != nil {
			if err := a.memory.Add(assistantMsg); err != nil {
				return "", fmt.Errorf("failed to add assistant message to memory: %w", err)
			}
			a.logger.Debug("💾 Saved assistant message with %d tool calls to memory", len(response.ToolCalls))
		}

		// Execute each tool
		for _, toolCall := range response.ToolCalls {
			// Log tool call
			LogToolCall(a.logger, toolCall.Function.Name, toolCall.Function.Arguments)

			result, err := a.tools.Execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				// Return error as tool result
				result = map[string]interface{}{
					"error": err.Error(),
				}
				LogToolResult(a.logger, toolCall.Function.Name, false, err)
			} else {
				LogToolResult(a.logger, toolCall.Function.Name, true, result)
			}

			// Add tool result to messages
			toolMsg := types.Message{
				Role:    types.RoleTool,
				Content: fmt.Sprintf("%v", result),
				ToolID:  toolCall.ID,
			}
			currentMessages = append(currentMessages, toolMsg)

			// Save tool result to memory
			if a.memory != nil {
				if err := a.memory.Add(toolMsg); err != nil {
					return "", fmt.Errorf("failed to add tool message to memory: %w", err)
				}
			}
		}

		if a.memory != nil {
			a.logger.Debug("💾 Saved %d tool results to memory", len(response.ToolCalls))
		}

		// Continue loop to let LLM process tool results
	}

	return "", fmt.Errorf("max iterations (%d) reached", a.options.MaxIterations)
}

// Reset clears the conversation history
func (a *Agent) Reset() error {
	if a.memory != nil {
		return a.memory.Clear()
	}
	return nil
}

// GetHistory returns the conversation history
func (a *Agent) GetHistory() ([]types.Message, error) {
	if a.memory == nil {
		return []types.Message{}, nil
	}
	return a.memory.GetHistory(0)
}

// ChatStream sends a message and streams the response via callback
func (a *Agent) ChatStream(ctx context.Context, message string, handler types.StreamHandler) error {
	// Add user message to memory if available
	userMsg := types.Message{
		Role:    types.RoleUser,
		Content: message,
	}

	if a.memory != nil {
		if err := a.memory.Add(userMsg); err != nil {
			return fmt.Errorf("failed to add message to memory: %w", err)
		}
	}

	// Get conversation history
	var messages []types.Message
	if a.memory != nil {
		history, err := a.memory.GetHistory(0)
		if err != nil {
			return fmt.Errorf("failed to get history: %w", err)
		}
		messages = history
	} else {
		messages = []types.Message{userMsg}
	}

	// Prepare chat options
	chatOpts := &types.ChatOptions{
		SystemPrompt: a.options.SystemPrompt,
		Temperature:  a.options.Temperature,
		MaxTokens:    a.options.MaxTokens,
	}

	// Add tools if available
	if a.tools.Size() > 0 {
		chatOpts.Tools = a.tools.GetDefinitions()
	}

	// Accumulate full response for memory
	var fullResponse string
	var toolCalls []types.ToolCall

	// Wrap handler to accumulate response
	wrappedHandler := func(chunk types.StreamChunk) error {
		// Accumulate content
		fullResponse += chunk.Content

		// Store tool calls from final chunk
		if chunk.Done && len(chunk.ToolCalls) > 0 {
			toolCalls = chunk.ToolCalls
		}

		// Call user's handler
		return handler(chunk)
	}

	// Stream response
	if err := a.provider.Stream(ctx, messages, chatOpts, wrappedHandler); err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	// Handle tool calls if present
	if len(toolCalls) > 0 {
		// Add assistant message with tool calls to memory
		if a.memory != nil {
			assistantMsg := types.Message{
				Role:      types.RoleAssistant,
				Content:   fullResponse,
				ToolCalls: toolCalls,
			}
			if err := a.memory.Add(assistantMsg); err != nil {
				return fmt.Errorf("failed to add assistant message to memory: %w", err)
			}
		}

		// Execute tools and continue (non-streaming for now, can enhance later)
		// This is a simplified version - for full streaming with tools,
		// we'd need a more complex loop
		for _, tc := range toolCalls {
			result, err := a.tools.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				result = map[string]interface{}{"error": err.Error()}
			}

			toolMsg := types.Message{
				Role:    types.RoleTool,
				Content: fmt.Sprintf("%v", result),
				ToolID:  tc.ID,
			}

			if a.memory != nil {
				a.memory.Add(toolMsg)
			}
		}

		// Get final response after tool execution (non-streaming)
		finalResp, err := a.Chat(ctx, "")
		if err != nil {
			return err
		}

		// Send final response as a single chunk
		finalChunk := types.StreamChunk{
			Content: finalResp,
			Done:    true,
		}
		return handler(finalChunk)
	}

	// Add assistant response to memory
	if a.memory != nil {
		assistantMsg := types.Message{
			Role:    types.RoleAssistant,
			Content: fullResponse,
		}
		if err := a.memory.Add(assistantMsg); err != nil {
			return fmt.Errorf("failed to add response to memory: %w", err)
		}
	}

	return nil
}
