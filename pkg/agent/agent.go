package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taipm/go-llm-agent/pkg/builtin"
	"github.com/taipm/go-llm-agent/pkg/learning"
	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/reasoning"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// Agent orchestrates LLM, tools, and memory with automatic reasoning
type Agent struct {
	provider types.LLMProvider
	tools    *tools.Registry
	memory   types.Memory
	options  *Options
	logger   logger.Logger

	// Reasoning engines (lazy initialized)
	reactAgent *reasoning.ReActAgent
	cotAgent   *reasoning.CoTAgent
	reflector  *reasoning.Reflector
	planner    *reasoning.Planner

	// Learning system (lazy initialized)
	experienceStore *learning.ExperienceStore
	toolSelector    *learning.ToolSelector
	conversationID  string // Current session ID

	// Auto-reasoning settings
	enableAutoReasoning bool
}

// Options contains configuration for the agent
type Options struct {
	SystemPrompt     string
	Temperature      float64
	MaxTokens        int
	MaxIterations    int     // Maximum tool calling iterations
	MinConfidence    float64 // Minimum confidence for reflection (0.0 = disabled)
	EnableReflection bool    // Enable self-reflection verification
	EnableLearning   bool    // Enable experience tracking and learning
}

// DefaultOptions returns default agent options
func DefaultOptions() *Options {
	return &Options{
		SystemPrompt:     "You are a helpful AI assistant.",
		Temperature:      0.7,
		MaxTokens:        2000,
		MaxIterations:    10,
		MinConfidence:    0.7,  // Default: require 70% confidence
		EnableReflection: true, // Enable reflection by default
		EnableLearning:   true, // Enable learning by default
	}
}

// New creates a new agent with default memory (100 messages) and all builtin tools
func New(provider types.LLMProvider, opts ...Option) *Agent {
	// Create default logger with DEBUG level for detailed reasoning
	defaultLogger := logger.NewConsoleLogger()
	defaultLogger.SetLevel(logger.LogLevelDebug)

	// Try to create VectorMemory for best learning experience
	// Fallback to BufferMemory if Qdrant not available
	defaultMemory := tryCreateVectorMemory(defaultLogger)

	agent := &Agent{
		provider:            provider,
		tools:               tools.NewRegistry(),
		memory:              defaultMemory,
		options:             DefaultOptions(),
		logger:              defaultLogger,       // Default logger with DEBUG level
		enableAutoReasoning: true,                // Enable auto reasoning by default
		conversationID:      uuid.New().String(), // Generate unique session ID
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

// tryCreateVectorMemory attempts to create VectorMemory with default settings
// Falls back to BufferMemory if Qdrant is not available
func tryCreateVectorMemory(log logger.Logger) types.Memory {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to create VectorMemory with default Qdrant settings
	vectorMem, err := memory.NewVectorMemory(ctx, memory.VectorMemoryConfig{
		QdrantURL:      "localhost:6334",                 // Default Qdrant address
		CollectionName: "agent_memory",                   // Default collection name
		Embedder:       memory.NewOllamaEmbedder("", ""), // Default Ollama embedder
		CacheSize:      100,
	})

	if err != nil {
		// Qdrant not available, use BufferMemory instead
		log.Info("ℹ️  Using BufferMemory (Qdrant not available)")
		log.Info("💡 For full learning with semantic search, start Qdrant:")
		log.Info("   docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant")
		return memory.NewBuffer(100)
	}

	log.Info("✅ VectorMemory initialized - full learning enabled")
	return vectorMem
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
func WithLogger(log logger.Logger) Option {
	return func(a *Agent) {
		a.logger = log
	}
}

// WithLogLevel sets the log level
func WithLogLevel(level logger.LogLevel) Option {
	return func(a *Agent) {
		a.logger.SetLevel(level)
	}
}

// DisableLogging disables all logging
func DisableLogging() Option {
	return func(a *Agent) {
		a.logger = &logger.NoopLogger{}
	}
}

// WithAutoReasoning enables/disables automatic reasoning detection
func WithAutoReasoning(enabled bool) Option {
	return func(a *Agent) {
		a.enableAutoReasoning = enabled
	}
}

// WithoutAutoReasoning disables automatic reasoning (use simple LLM calls only)
func WithoutAutoReasoning() Option {
	return func(a *Agent) {
		a.enableAutoReasoning = false
	}
}

// WithoutBuiltinTools disables automatic loading of builtin tools
func WithoutBuiltinTools() Option {
	return func(a *Agent) {
		// Clear the tools registry
		a.tools = tools.NewRegistry()
	}
}

// WithReflection enables or disables self-reflection
func WithReflection(enabled bool) Option {
	return func(a *Agent) {
		a.options.EnableReflection = enabled
	}
}

// WithMinConfidence sets the minimum confidence threshold for reflection
func WithMinConfidence(threshold float64) Option {
	return func(a *Agent) {
		a.options.MinConfidence = threshold
	}
}

// WithLearning enables experience tracking and learning
// Note: Requires AdvancedMemory (e.g., VectorMemory) to work properly
// If using BufferMemory, learning will log a warning but continue to work with limited functionality
func WithLearning(enabled bool) Option {
	return func(a *Agent) {
		a.options.EnableLearning = enabled

		if enabled {
			// Helpful log for users
			a.logger.Info("� Learning enabled")
			a.logger.Info("   💡 Tip: Use VectorMemory for best learning performance")
			a.logger.Info("   Example: agent.WithMemory(vectorMemory)")
		}
	}
}

// AddTool registers a tool with the agent
func (a *Agent) AddTool(t tools.Tool) error {
	return a.tools.Register(t)
}

// RemoveTool unregisters a tool
func (a *Agent) RemoveTool(name string) {
	a.tools.Unregister(name)
}

// ToolCount returns the number of registered tools
func (a *Agent) ToolCount() int {
	return a.tools.Count()
}

// AgentStatus contains comprehensive information about agent state
type AgentStatus struct {
	// Configuration
	Configuration struct {
		SystemPrompt     string  `json:"system_prompt"`
		Temperature      float64 `json:"temperature"`
		MaxTokens        int     `json:"max_tokens"`
		MaxIterations    int     `json:"max_iterations"`
		MinConfidence    float64 `json:"min_confidence"`
		EnableReflection bool    `json:"enable_reflection"`
		EnableLearning   bool    `json:"enable_learning"`
	} `json:"configuration"`

	// Reasoning Capabilities
	Reasoning struct {
		AutoReasoningEnabled bool `json:"auto_reasoning_enabled"`
		CoTAvailable         bool `json:"cot_available"`
		ReActAvailable       bool `json:"react_available"`
		ReflectionAvailable  bool `json:"reflection_available"`
	} `json:"reasoning"`

	// Tools
	Tools struct {
		TotalCount int      `json:"total_count"`
		ToolNames  []string `json:"tool_names"`
	} `json:"tools"`

	// Memory
	Memory struct {
		Type            string `json:"type"`
		MessageCount    int    `json:"message_count,omitempty"`
		SupportsSearch  bool   `json:"supports_search"`
		SupportsVectors bool   `json:"supports_vectors"`
	} `json:"memory"`

	// Learning
	Learning struct {
		Enabled              bool   `json:"enabled"`
		ExperienceStoreReady bool   `json:"experience_store_ready"`
		ToolSelectorReady    bool   `json:"tool_selector_ready"`
		ConversationID       string `json:"conversation_id"`
	} `json:"learning"`

	// Provider
	Provider struct {
		Type string `json:"type"`
	} `json:"provider"`
}

// Status returns comprehensive agent configuration and runtime state
func (a *Agent) Status() *AgentStatus {
	status := &AgentStatus{}

	// Configuration
	status.Configuration.SystemPrompt = a.options.SystemPrompt
	status.Configuration.Temperature = a.options.Temperature
	status.Configuration.MaxTokens = a.options.MaxTokens
	status.Configuration.MaxIterations = a.options.MaxIterations
	status.Configuration.MinConfidence = a.options.MinConfidence
	status.Configuration.EnableReflection = a.options.EnableReflection
	status.Configuration.EnableLearning = a.options.EnableLearning

	// Reasoning capabilities
	status.Reasoning.AutoReasoningEnabled = a.enableAutoReasoning
	status.Reasoning.CoTAvailable = (a.cotAgent != nil)
	status.Reasoning.ReActAvailable = (a.reactAgent != nil)
	status.Reasoning.ReflectionAvailable = (a.reflector != nil)

	// Tools
	status.Tools.TotalCount = a.tools.Count()
	status.Tools.ToolNames = make([]string, 0)
	for _, tool := range a.tools.All() {
		status.Tools.ToolNames = append(status.Tools.ToolNames, tool.Name())
	}

	// Memory
	status.Memory.Type = a.getMemoryType()
	if buffMem, ok := a.memory.(*memory.BufferMemory); ok {
		messages := buffMem.GetAll()
		status.Memory.MessageCount = len(messages)
	}

	// Check if memory supports advanced features
	if advMem, ok := a.memory.(types.AdvancedMemory); ok {
		status.Memory.SupportsSearch = true
		status.Memory.SupportsVectors = true
		// Try to get stats if available (best effort, ignore errors)
		if stats, _ := advMem.GetStats(context.Background()); stats != nil {
			status.Memory.MessageCount = stats.TotalMessages
		}
	}

	// Learning
	status.Learning.Enabled = a.options.EnableLearning
	status.Learning.ExperienceStoreReady = (a.experienceStore != nil)
	status.Learning.ToolSelectorReady = (a.toolSelector != nil)
	status.Learning.ConversationID = a.conversationID

	// Provider type (detect based on type assertion)
	status.Provider.Type = a.getProviderType()

	return status
}

// initExperienceStore initializes the experience store if learning is enabled
func (a *Agent) initExperienceStore() {
	if a.experienceStore != nil {
		return // Already initialized
	}

	// Check if memory supports advanced features (needed for semantic search)
	if advMem, ok := a.memory.(types.AdvancedMemory); ok {
		a.experienceStore = learning.NewExperienceStore(advMem)
		a.logger.Info("✅ Experience store ready (using VectorMemory)")

		// Also initialize tool selector
		a.initToolSelector()
	} else {
		// Running in limited mode with BufferMemory
		a.logger.Info("ℹ️  Learning running in limited mode (BufferMemory)")
		a.logger.Info("   For full learning with semantic search:")
		a.logger.Info("   docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant")
		// Don't fail - agent still works, just without advanced learning features
	}
}

// initToolSelector initializes the tool selector for learned tool recommendations
func (a *Agent) initToolSelector() {
	if a.toolSelector != nil {
		return // Already initialized
	}

	if a.experienceStore == nil {
		return // Need experience store first
	}

	a.toolSelector = learning.NewToolSelector(a.experienceStore, a.tools, a.logger)
	a.logger.Info("✅ Tool selector ready (ε-greedy learning active)")
}

// recordExperience records an experience for learning
func (a *Agent) recordExperience(ctx context.Context, query string, response string, err error, metadata map[string]interface{}) {
	if !a.options.EnableLearning || a.experienceStore == nil {
		return // Learning disabled or not initialized
	}

	// Create experience record
	exp := learning.Experience{
		ID:             uuid.New().String(),
		Timestamp:      time.Now(),
		Query:          query,
		Response:       response,
		Success:        (err == nil),
		ConversationID: a.conversationID,
		Metadata:       metadata,
	}

	// Add error details if failed
	if err != nil {
		exp.Error = err.Error()
	}

	// Extract intent and reasoning mode from metadata
	if intent, ok := metadata["intent"].(string); ok {
		exp.Intent = intent
	}
	if reasoning, ok := metadata["reasoning"].(string); ok {
		exp.ReasoningMode = reasoning
	}
	if tool, ok := metadata["tool"].(string); ok {
		exp.ToolCalled = tool
	}
	if args, ok := metadata["arguments"].(map[string]interface{}); ok {
		exp.Arguments = args
	}
	if latency, ok := metadata["latency_ms"].(int64); ok {
		exp.LatencyMs = latency
	}
	if confidence, ok := metadata["confidence"].(float64); ok {
		exp.Confidence = confidence
	}
	if reflected, ok := metadata["was_reflected"].(bool); ok {
		exp.WasReflected = reflected
	}
	if corrected, ok := metadata["was_corrected"].(bool); ok {
		exp.WasCorrected = corrected
	}

	// Record experience (async, don't block)
	go func() {
		if err := a.experienceStore.Record(context.Background(), exp); err != nil {
			a.logger.Debug("Failed to record experience: %v", err)
		} else {
			a.logger.Debug("📚 Recorded experience: %s (success=%v)", exp.ID, exp.Success)
		}
	}()
}

// getMemoryType returns a human-readable memory type
func (a *Agent) getMemoryType() string {
	switch a.memory.(type) {
	case *memory.BufferMemory:
		return "buffer"
	case *memory.VectorMemory:
		return "vector"
	default:
		if _, ok := a.memory.(types.AdvancedMemory); ok {
			return "advanced"
		}
		return "custom"
	}
}

// getProviderType returns a human-readable provider type
func (a *Agent) getProviderType() string {
	providerType := fmt.Sprintf("%T", a.provider)

	// Extract simple name from full package path
	// e.g., "*ollama.Provider" -> "ollama"
	parts := strings.Split(providerType, ".")
	if len(parts) > 1 {
		name := parts[len(parts)-2] // Get package name
		// Remove any leading asterisk or path
		name = strings.TrimPrefix(name, "*")
		name = strings.TrimPrefix(name, "provider/")
		return name
	}

	return providerType
}

// Chat sends a message and returns the response
func (a *Agent) Chat(ctx context.Context, message string) (string, error) {
	// Initialize experience store if learning is enabled
	if a.options.EnableLearning && a.experienceStore == nil {
		a.initExperienceStore()
	}

	// Track start time for latency
	startTime := time.Now()

	// Log user message
	logger.LogUserMessage(a.logger, message)

	// Metadata for experience recording
	metadata := make(map[string]interface{})
	var response string
	var err error

	// Check if auto-reasoning is enabled
	if a.enableAutoReasoning {
		approach := a.analyzeQuery(message)
		a.logger.Debug("🧠 Query analysis: %s approach selected", approach)
		metadata["reasoning"] = approach
		metadata["intent"] = a.detectIntent(message)

		switch approach {
		case "cot":
			response, err = a.chatWithCoT(ctx, message)
		case "react":
			response, err = a.chatWithReAct(ctx, message)
		default:
			// Fall through to simple chat if "simple"
			response, err = a.chatSimple(ctx, message)
		}
	} else {
		// Simple chat without reasoning
		metadata["reasoning"] = "simple"
		metadata["intent"] = a.detectIntent(message)
		response, err = a.chatSimple(ctx, message)
	}

	// Calculate latency
	metadata["latency_ms"] = time.Since(startTime).Milliseconds()

	// Extract tool usage from conversation history (if any)
	if a.memory != nil {
		if toolInfo := a.extractLastToolUsage(); toolInfo != nil {
			metadata["tool"] = toolInfo["name"]
			metadata["arguments"] = toolInfo["arguments"]
		}
	}

	// Record experience for learning
	a.recordExperience(ctx, message, response, err, metadata)

	return response, err
}

// extractLastToolUsage extracts the most recent tool call from memory
func (a *Agent) extractLastToolUsage() map[string]interface{} {
	if a.memory == nil {
		return nil
	}

	// Get recent history
	history, err := a.memory.GetHistory(10) // Last 10 messages
	if err != nil {
		return nil
	}

	// Search backwards for assistant message with tool calls
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role == types.RoleAssistant && len(msg.ToolCalls) > 0 {
			// Return first tool call info
			firstCall := msg.ToolCalls[0]
			return map[string]interface{}{
				"name":      firstCall.Function.Name,
				"arguments": firstCall.Function.Arguments,
			}
		}
	}

	return nil
}

// chatSimple performs simple LLM chat with tool calling (original behavior)
func (a *Agent) chatSimple(ctx context.Context, message string) (string, error) {
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
	if a.tools.Count() > 0 {
		chatOpts.Tools = a.tools.ToToolDefinitions()
	}

	// Run agent loop with tool calling
	response, err := a.runLoop(ctx, messages, chatOpts)
	if err != nil {
		a.logger.Error("Agent execution failed: %v", err)
		return "", err
	}

	// Log final response
	logger.LogResponse(a.logger, response)

	// Note: runLoop already saves the final response to memory
	return response, nil
}

// runLoop executes the agent loop with tool calling
func (a *Agent) runLoop(ctx context.Context, messages []types.Message, opts *types.ChatOptions) (string, error) {
	currentMessages := make([]types.Message, len(messages))
	copy(currentMessages, messages)

	for iteration := 0; iteration < a.options.MaxIterations; iteration++ {
		logger.LogIteration(a.logger, iteration, a.options.MaxIterations)

		// Log thinking
		logger.LogThinking(a.logger)

		// Call LLM
		response, err := a.provider.Chat(ctx, currentMessages, opts)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		// If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			a.logger.Debug("No tool calls, returning response")

			answer := response.Content

			// Apply reflection if enabled
			if a.options.EnableReflection && len(currentMessages) > 0 {
				// Get the original question (first user message)
				var question string
				for _, msg := range currentMessages {
					if msg.Role == types.RoleUser {
						question = msg.Content
						break
					}
				}
				if question != "" {
					answer = a.applyReflection(ctx, question, answer)
				}
			}

			// Save final assistant response to memory
			if a.memory != nil {
				finalMsg := types.Message{
					Role:    types.RoleAssistant,
					Content: answer,
				}
				if err := a.memory.Add(finalMsg); err != nil {
					return "", fmt.Errorf("failed to add final response to memory: %w", err)
				}
				a.logger.Debug("💾 Saved assistant response to memory")
			}
			return answer, nil
		}

		// Log tool calls
		a.logger.Info("🔧 Agent wants to call %d tool(s): %s", len(response.ToolCalls), logger.FormatToolCalls(response.ToolCalls))

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
			logger.LogToolCall(a.logger, toolCall.Function.Name, toolCall.Function.Arguments)

			result, err := a.tools.Execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				// Return error as tool result
				result = map[string]interface{}{
					"error": err.Error(),
				}
				logger.LogToolResult(a.logger, toolCall.Function.Name, false, err)
			} else {
				logger.LogToolResult(a.logger, toolCall.Function.Name, true, result)
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
	if a.tools.Count() > 0 {
		chatOpts.Tools = a.tools.ToToolDefinitions()
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

// analyzeQuery determines which reasoning approach to use
func (a *Agent) analyzeQuery(query string) string {
	queryLower := strings.ToLower(query)

	// Priority 1: Explicit tool usage requests (highest priority)
	explicitToolKeywords := []string{
		"use tool", "using tool", "call tool",
		"use calculator", "use the calculator",
		"search the web", "search web", "web search",
		"fetch from", "scrape from",
	}
	for _, keyword := range explicitToolKeywords {
		if strings.Contains(queryLower, keyword) {
			return "react"
		}
	}

	// Priority 2: Check for Chain-of-Thought indicators
	if needsCoT(queryLower) {
		return "cot"
	}

	// Priority 3: Check for ReAct/tool usage indicators
	if a.tools.Count() > 0 && needsTools(queryLower) {
		return "react"
	}

	// Default to simple chat
	return "simple"
}

// detectIntent identifies the user's intent from the query
func (a *Agent) detectIntent(query string) string {
	queryLower := strings.ToLower(query)

	// Math/calculation intent
	if strings.Contains(queryLower, "calculate") ||
		strings.Contains(queryLower, "compute") ||
		strings.Contains(queryLower, "solve") ||
		regexp.MustCompile(`\d+\s*[\+\-\*\/]\s*\d+`).MatchString(query) {
		return "calculation"
	}

	// Web search intent
	if strings.Contains(queryLower, "search") ||
		strings.Contains(queryLower, "find") ||
		strings.Contains(queryLower, "look up") ||
		strings.Contains(queryLower, "what is") {
		return "information_retrieval"
	}

	// File operations
	if strings.Contains(queryLower, "read file") ||
		strings.Contains(queryLower, "write file") ||
		strings.Contains(queryLower, "save to") ||
		strings.Contains(queryLower, "open file") {
		return "file_operation"
	}

	// Code/technical
	if strings.Contains(queryLower, "code") ||
		strings.Contains(queryLower, "program") ||
		strings.Contains(queryLower, "function") ||
		strings.Contains(queryLower, "debug") {
		return "coding"
	}

	// General conversation
	return "conversation"
}

// needsCoT detects if query requires step-by-step reasoning
func needsCoT(query string) bool {
	// Mathematical problem indicators
	mathIndicators := []string{
		"calculate", "compute", "solve", "what is",
		"how many", "how much", "if.*then",
	}

	// Multi-step reasoning indicators
	reasoningIndicators := []string{
		"step by step", "explain how", "why",
		"prove", "show that", "derive",
	}

	// Check indicators
	for _, indicator := range mathIndicators {
		if strings.Contains(query, indicator) {
			return true
		}
	}

	for _, indicator := range reasoningIndicators {
		if strings.Contains(query, indicator) {
			return true
		}
	}

	// Multiple numbers suggest calculation
	numberPattern := regexp.MustCompile(`\d+(\.\d+)?`)
	numbers := numberPattern.FindAllString(query, -1)
	if len(numbers) >= 2 {
		return true
	}

	return false
}

// needsTools detects if query requires tool usage
func needsTools(query string) bool {
	// Tool usage indicators
	toolIndicators := []string{
		"using", "with", "tool", "calculator",
		"search", "find", "look up", "get",
	}

	// Action verbs suggesting tool usage
	actionVerbs := []string{
		"calculate", "compute", "search", "find",
		"fetch", "retrieve", "get", "check",
	}

	for _, indicator := range toolIndicators {
		if strings.Contains(query, indicator) {
			return true
		}
	}

	for _, verb := range actionVerbs {
		if strings.Contains(query, verb) {
			return true
		}
	}

	return false
}

// chatWithCoT uses Chain-of-Thought reasoning
func (a *Agent) chatWithCoT(ctx context.Context, message string) (string, error) {
	a.logger.Info("🧠 Using Chain-of-Thought reasoning")

	// Lazy initialize CoT agent
	if a.cotAgent == nil {
		a.cotAgent = reasoning.NewCoTAgent(a.provider, a.memory, 10)
		a.cotAgent.WithLogger(a.logger)
	}

	// Think through the problem
	answer, err := a.cotAgent.Think(ctx, message)
	if err != nil {
		a.logger.Warn("⚠️  CoT reasoning failed, falling back to simple chat: %v", err)
		return a.chatSimple(ctx, message)
	}

	// Save to memory
	if a.memory != nil {
		a.cotAgent.SaveToMemory(ctx)
	}

	// Apply reflection if enabled
	if a.options.EnableReflection {
		answer = a.applyReflection(ctx, message, answer)
	}

	return answer, nil
}

// chatWithReAct uses ReAct pattern with tools
func (a *Agent) chatWithReAct(ctx context.Context, message string) (string, error) {
	a.logger.Info("🔧 Using ReAct reasoning with tools")

	// Lazy initialize ReAct agent
	if a.reactAgent == nil {
		allTools := a.tools.All()
		a.reactAgent = reasoning.NewReActAgent(a.provider, a.memory, a.options.MaxIterations)
		a.reactAgent.WithLogger(a.logger)
		a.reactAgent.WithTools(allTools...)
	}

	// Run ReAct loop
	var finalAnswer string
	for i := 0; i < a.options.MaxIterations; i++ {
		step, err := a.reactAgent.Think(ctx, message)
		if err != nil {
			a.logger.Warn("⚠️  ReAct iteration %d failed: %v", i+1, err)
			return "", fmt.Errorf("ReAct reasoning failed: %w", err)
		}

		// Check if we have final answer
		if step.Action == "Answer" {
			finalAnswer = step.Observation
			a.logger.Info("✅ ReAct completed in %d iterations", i+1)
			break
		}

		// Log iteration
		a.logger.Debug("   Iteration %d: %s → %s", i+1, step.Action, step.Observation)
	}

	if finalAnswer == "" {
		a.logger.Warn("⚠️  ReAct max iterations reached, falling back to simple chat")
		return a.chatSimple(ctx, message)
	}

	// Save to memory
	if a.memory != nil {
		userMsg := types.Message{Role: types.RoleUser, Content: message}
		assistantMsg := types.Message{Role: types.RoleAssistant, Content: finalAnswer}
		a.memory.Add(userMsg)
		a.memory.Add(assistantMsg)
	}

	return finalAnswer, nil
}

// applyReflection performs self-reflection on an answer and returns the final (possibly corrected) answer
func (a *Agent) applyReflection(ctx context.Context, question string, initialAnswer string) string {
	// Lazy initialize reflector
	if a.reflector == nil {
		a.reflector = reasoning.NewReflector(a.provider, a.memory)
		a.reflector.WithLogger(a.logger)
		// Add tools for verification
		allTools := a.tools.All()
		a.reflector.WithTools(allTools...)
	}

	// Perform reflection
	reflection, err := a.reflector.Reflect(ctx, question, initialAnswer)
	if err != nil {
		a.logger.Warn("⚠️  Reflection failed: %v, using initial answer", err)
		return initialAnswer
	}

	// Check confidence threshold
	if reflection.Confidence < a.options.MinConfidence {
		a.logger.Warn("⚠️  Low confidence (%.2f < %.2f)", reflection.Confidence, a.options.MinConfidence)

		// If answer was corrected, use the corrected version
		if reflection.WasCorrected {
			a.logger.Info("🔧 Using corrected answer")

			// Save correction note to memory
			if a.memory != nil {
				correctionNote := types.Message{
					Role:    types.RoleAssistant,
					Content: fmt.Sprintf("[CORRECTED via reflection] %s", reflection.FinalAnswer),
				}
				a.memory.Add(correctionNote)
			}

			return reflection.FinalAnswer
		}
	} else {
		a.logger.Info("✅ High confidence (%.2f)", reflection.Confidence)
	}

	return reflection.FinalAnswer
}

// ChatWithReflection performs chat with self-reflection and verification
// Returns the reflection check for analysis
// NOTE: For normal usage, just use Chat() with EnableReflection=true
func (a *Agent) ChatWithReflection(ctx context.Context, message string, minConfidence float64) (*types.ReflectionCheck, error) {
	a.logger.Info("💭 Chat with self-reflection enabled (min confidence: %.2f)", minConfidence)

	// Step 1: Get initial answer using normal chat
	initialAnswer, err := a.Chat(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial answer: %w", err)
	}

	// Step 2: Lazy initialize reflector
	if a.reflector == nil {
		a.reflector = reasoning.NewReflector(a.provider, a.memory)
		a.reflector.WithLogger(a.logger)
		// Add tools for verification (web_search, math_calculate, etc.)
		allTools := a.tools.All()
		a.reflector.WithTools(allTools...)
	}

	// Step 3: Perform reflection
	reflection, err := a.reflector.Reflect(ctx, message, initialAnswer)
	if err != nil {
		a.logger.Warn("⚠️  Reflection failed: %v, returning initial answer", err)
		// Return reflection with initial answer even if reflection failed
		return &types.ReflectionCheck{
			Question:      message,
			InitialAnswer: initialAnswer,
			FinalAnswer:   initialAnswer,
			Confidence:    0.5,
		}, nil
	}

	// Step 4: Check if confidence meets threshold
	if reflection.Confidence < minConfidence {
		a.logger.Warn("⚠️  Confidence (%.2f) below threshold (%.2f)", reflection.Confidence, minConfidence)
		if !reflection.WasCorrected {
			a.logger.Warn("   No correction was made - consider this answer uncertain")
		}
	} else {
		a.logger.Info("✅ Confidence (%.2f) meets threshold", reflection.Confidence)
	}

	// Step 5: Update memory with final answer if it was corrected
	if reflection.WasCorrected && a.memory != nil {
		// The initial answer was already saved by Chat()
		// Now we need to update or add a correction note
		correctionNote := types.Message{
			Role:    types.RoleAssistant,
			Content: fmt.Sprintf("[REFLECTION CORRECTION] %s", reflection.FinalAnswer),
		}
		if err := a.memory.Add(correctionNote); err != nil {
			a.logger.Warn("⚠️  Failed to save correction to memory: %v", err)
		}
	}

	return reflection, nil
}

// WithReflection enables or disables automatic reflection for all chats
func (a *Agent) WithReflection(enable bool) *Agent {
	a.options.EnableReflection = enable
	return a
}

// WithMinConfidence sets the minimum confidence threshold for reflection
func (a *Agent) WithMinConfidence(minConfidence float64) *Agent {
	a.options.MinConfidence = minConfidence
	return a
}

// Plan creates a task decomposition plan for a complex goal
func (a *Agent) Plan(ctx context.Context, goal string) (*types.Plan, error) {
	a.logger.Info("📋 Creating plan for goal: %s", goal)

	// Lazy initialize planner
	if a.planner == nil {
		a.planner = reasoning.NewPlanner(a.provider, a.memory, a.logger, true)
	}

	// Decompose goal into plan
	plan, err := a.planner.DecomposeGoal(ctx, goal)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	// Store plan in memory
	if err := a.planner.SaveToMemory(ctx, plan); err != nil {
		a.logger.Warn("⚠️  Failed to save plan to memory: %v", err)
	}

	return plan, nil
}

// ExecutePlan executes a plan by running each step through the agent
func (a *Agent) ExecutePlan(ctx context.Context, plan *types.Plan) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}

	// Lazy initialize planner
	if a.planner == nil {
		a.planner = reasoning.NewPlanner(a.provider, a.memory, a.logger, true)
	}

	// Execute plan with agent as executor
	executor := func(ctx context.Context, task string) (interface{}, error) {
		// Use agent.Chat to execute each step
		result, err := a.Chat(ctx, task)
		return result, err
	}

	return a.planner.ExecutePlan(ctx, plan, executor)
}

// GetPlanProgress returns the execution progress of a plan
func (a *Agent) GetPlanProgress(plan *types.Plan) *types.PlanProgress {
	if a.planner == nil {
		a.planner = reasoning.NewPlanner(a.provider, a.memory, a.logger, true)
	}
	return a.planner.GetProgress(plan)
}

// GetToolRecommendation returns a learned recommendation for which tool to use
func (a *Agent) GetToolRecommendation(ctx context.Context, query string, intent string) (*learning.ToolRecommendation, error) {
	// Ensure learning is initialized
	if !a.options.EnableLearning {
		return nil, fmt.Errorf("learning is not enabled")
	}

	if a.experienceStore == nil {
		a.initExperienceStore()
	}

	if a.toolSelector == nil {
		return nil, fmt.Errorf("tool selector not initialized")
	}

	return a.toolSelector.RecommendTool(ctx, query, intent)
}

// GetToolStats returns performance statistics for a specific tool
func (a *Agent) GetToolStats(ctx context.Context, toolName string, intent string) (*learning.ToolStats, error) {
	if !a.options.EnableLearning || a.toolSelector == nil {
		return nil, fmt.Errorf("learning is not enabled or tool selector not initialized")
	}

	return a.toolSelector.GetToolStats(ctx, toolName, intent)
}
