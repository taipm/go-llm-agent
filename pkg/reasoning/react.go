package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// ReActAgent implements the ReAct (Reasoning + Acting) pattern
// Paper: https://arxiv.org/abs/2210.03629
type ReActAgent struct {
	provider types.LLMProvider
	memory   types.Memory
	registry *tools.Registry
	logger   logger.Logger
	steps    []types.ReActStep
	maxSteps int
	verbose  bool
}

// NewReActAgent creates a new ReAct agent
func NewReActAgent(provider types.LLMProvider, memory types.Memory, maxSteps int) *ReActAgent {
	if maxSteps <= 0 {
		maxSteps = 10 // Default max iterations
	}
	return &ReActAgent{
		provider: provider,
		memory:   memory,
		registry: tools.NewRegistry(),
		logger:   logger.NewConsoleLogger(),
		steps:    make([]types.ReActStep, 0, maxSteps),
		maxSteps: maxSteps,
		verbose:  true,
	}
}

// WithTools adds tools to the agent
func (r *ReActAgent) WithTools(toolList ...tools.Tool) *ReActAgent {
	for _, tool := range toolList {
		r.registry.Register(tool)
	}
	return r
}

// WithLogger sets a custom logger
func (r *ReActAgent) WithLogger(log logger.Logger) *ReActAgent {
	r.logger = log
	return r
}

// SetVerbose controls whether ReAct steps are printed to stdout
func (r *ReActAgent) SetVerbose(verbose bool) {
	r.verbose = verbose
}

// GetSteps returns all ReAct steps from the current reasoning process
func (r *ReActAgent) GetSteps() []types.ReActStep {
	return r.steps
}

// ClearSteps resets the reasoning history
func (r *ReActAgent) ClearSteps() {
	r.steps = make([]types.ReActStep, 0, r.maxSteps)
}

// buildReActPrompt creates a prompt that guides the LLM to think explicitly
func (r *ReActAgent) buildReActPrompt(query string, previousSteps []types.ReActStep) string {
	var prompt strings.Builder

	prompt.WriteString("You are a helpful AI assistant using the ReAct (Reasoning + Acting) pattern.\n\n")

	// List available tools first
	if r.registry != nil && len(r.registry.All()) > 0 {
		prompt.WriteString("You have access to these tools:\n")
		for _, tool := range r.registry.All() {
			prompt.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
		}
		prompt.WriteString("\nIMPORTANT:\n")
		prompt.WriteString("- Use tools ONE AT A TIME for each calculation step\n")
		prompt.WriteString("- After each tool result, reflect on whether you need more steps\n")
		prompt.WriteString("- When you have enough information, provide your final answer\n\n")
	}

	// Show previous steps if any
	if len(previousSteps) > 0 {
		prompt.WriteString("Previous steps:\n")
		for _, step := range previousSteps {
			if step.Action != "Answer" {
				prompt.WriteString(fmt.Sprintf("- Used %s → got result: %s\n", step.Action, step.Observation))
			}
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString(fmt.Sprintf("Question: %s\n\n", query))

	if len(previousSteps) > 0 {
		prompt.WriteString("Based on the previous results, what should you do next?\n")
		prompt.WriteString("- If you need another calculation, call a tool\n")
		prompt.WriteString("- If you have enough information, provide your final answer\n")
	} else {
		prompt.WriteString("Think step-by-step:\n")
		prompt.WriteString("1. What calculations do you need?\n")
		prompt.WriteString("2. Call the appropriate tool for the FIRST step\n")
		prompt.WriteString("3. After seeing the result, decide the next action\n")
	}

	return prompt.String()
}

// buildToolDefinitions converts tools to LLM function definitions
func (r *ReActAgent) buildToolDefinitions() []types.ToolDefinition {
	var defs []types.ToolDefinition

	for _, tool := range r.registry.All() {
		def := types.ToolDefinition{
			Type: "function",
			Function: types.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}
		defs = append(defs, def)
	}

	return defs
}

// parseReActResponse extracts Thought, Action, Observation, Reflection from LLM response
func (r *ReActAgent) parseReActResponse(response string) (thought, action, observation, reflection string) {
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(strings.ToLower(line), "thought:") {
			thought = strings.TrimSpace(strings.TrimPrefix(line, "Thought:"))
			thought = strings.TrimSpace(strings.TrimPrefix(thought, "thought:"))
		} else if strings.HasPrefix(strings.ToLower(line), "action:") {
			action = strings.TrimSpace(strings.TrimPrefix(line, "Action:"))
			action = strings.TrimSpace(strings.TrimPrefix(action, "action:"))
		} else if strings.HasPrefix(strings.ToLower(line), "observation:") {
			observation = strings.TrimSpace(strings.TrimPrefix(line, "Observation:"))
			observation = strings.TrimSpace(strings.TrimPrefix(observation, "observation:"))
		} else if strings.HasPrefix(strings.ToLower(line), "reflection:") {
			reflection = strings.TrimSpace(strings.TrimPrefix(line, "Reflection:"))
			reflection = strings.TrimSpace(strings.TrimPrefix(reflection, "reflection:"))
		}
	}

	// If no explicit structure, treat entire response as thought
	if thought == "" && action == "" && observation == "" && reflection == "" {
		thought = response
		action = "Answer"
	}

	return
}

// executeAction parses and executes a tool call from the action string
// Expected formats:
//   - "tool_name(param1, param2)"
//   - "tool_name({'key': 'value'})"
//   - "tool_name"
func (r *ReActAgent) executeAction(ctx context.Context, action string) (string, error) {
	// Simple parsing: extract tool name and parameters
	action = strings.TrimSpace(action)

	// Try to find tool name (before parenthesis or the whole string)
	toolName := action
	paramsStr := ""

	if idx := strings.Index(action, "("); idx != -1 {
		toolName = strings.TrimSpace(action[:idx])
		if endIdx := strings.LastIndex(action, ")"); endIdx > idx {
			paramsStr = strings.TrimSpace(action[idx+1 : endIdx])
		}
	}

	// Get tool from registry
	tool := r.registry.Get(toolName)
	if tool == nil {
		return "", fmt.Errorf("tool '%s' not found in registry", toolName)
	}

	r.logger.Info("🔧 Executing tool: %s", toolName)
	if paramsStr != "" {
		r.logger.Debug("   Parameters: %s", paramsStr)
	}

	// Parse parameters (simple JSON or map[string]interface{})
	params := make(map[string]interface{})
	if paramsStr != "" {
		// Try to parse as JSON first
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			// If not JSON, treat as simple key-value
			r.logger.Debug("   Using raw parameter string")
			params["query"] = paramsStr
		}
	}

	// Execute tool
	result, err := tool.Execute(ctx, params)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	// Convert result to string
	resultStr := fmt.Sprintf("%v", result)
	r.logger.Debug("   Result: %s", resultStr)

	return resultStr, nil
}

// Think performs one iteration of ReAct reasoning
func (r *ReActAgent) Think(ctx context.Context, query string) (*types.ReActStep, error) {
	iteration := len(r.steps) + 1

	// Build prompt with previous steps
	prompt := r.buildReActPrompt(query, r.steps)

	// Prepare tools for LLM
	var toolDefs []types.ToolDefinition
	if r.registry != nil && len(r.registry.All()) > 0 {
		toolDefs = r.buildToolDefinitions()
	}

	// Call LLM with function calling
	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: prompt,
		},
	}

	response, err := r.provider.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   1000,
		Tools:       toolDefs,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	thought := ""
	action := "Answer"
	observation := ""
	reflection := ""
	finalAnswer := ""

	// Check if LLM wants to call a tool
	if len(response.ToolCalls) > 0 {
		// Execute the first tool call
		toolCall := response.ToolCalls[0]
		action = toolCall.Function.Name

		r.logger.Info("🔧 LLM requested tool: %s", action)
		r.logger.Debug("   Parameters: %v", toolCall.Function.Arguments)

		// Execute tool
		tool := r.registry.Get(action)
		if tool != nil {
			result, err := tool.Execute(ctx, toolCall.Function.Arguments)
			if err != nil {
				r.logger.Warn("⚠️  Tool execution failed: %v", err)
				observation = fmt.Sprintf("Tool execution failed: %v", err)
			} else {
				observation = fmt.Sprintf("%v", result)
				r.logger.Info("✅ Tool executed: %s = %s", action, observation)
			}
		} else {
			observation = fmt.Sprintf("Tool '%s' not found", action)
			r.logger.Warn("⚠️  %s", observation)
		}

		// Use LLM's thinking as thought
		thought = response.Content
		reflection = fmt.Sprintf("Tool %s returned: %s", action, observation)
	} else {
		// No tool call - this is the final answer
		thought = response.Content
		action = "Answer"
		finalAnswer = response.Content
		observation = finalAnswer
		reflection = "Ready to provide answer"
	}

	// Create ReAct step
	step := types.ReActStep{
		Iteration:   iteration,
		Thought:     thought,
		Action:      action,
		Observation: observation,
		Reflection:  reflection,
		Timestamp:   time.Now(),
	}

	// Store step
	r.steps = append(r.steps, step)

	// Print if verbose
	if r.verbose {
		r.printStep(step)
	}

	// Store in memory if available
	if r.memory != nil {
		metadata := map[string]interface{}{
			"react_step": step,
			"category":   types.CategoryReasoning,
		}
		msg := types.Message{
			Role:     types.RoleAssistant,
			Content:  fmt.Sprintf("ReAct Step %d: %s", iteration, thought),
			Metadata: metadata,
		}
		r.memory.Add(msg)
	}

	return &step, nil
}

// printStep prints a ReAct step to stdout
func (r *ReActAgent) printStep(step types.ReActStep) {
	fmt.Printf("\n=== ReAct Iteration %d ===\n", step.Iteration)
	fmt.Printf("💭 Thought: %s\n", step.Thought)
	fmt.Printf("🎯 Action: %s\n", step.Action)
	if step.Observation != "" {
		fmt.Printf("👁️  Observation: %s\n", step.Observation)
	}
	if step.Reflection != "" {
		fmt.Printf("🤔 Reflection: %s\n", step.Reflection)
	}
	fmt.Println()
}

// Solve uses ReAct pattern to solve a problem iteratively
func (r *ReActAgent) Solve(ctx context.Context, query string) (string, error) {
	r.ClearSteps() // Start fresh

	for i := 0; i < r.maxSteps; i++ {
		step, err := r.Think(ctx, query)
		if err != nil {
			return "", fmt.Errorf("reasoning failed at step %d: %w", i+1, err)
		}

		// Check if agent wants to provide final answer
		if strings.Contains(strings.ToLower(step.Action), "answer") ||
			strings.Contains(strings.ToLower(step.Action), "respond") ||
			strings.Contains(strings.ToLower(step.Action), "final") {
			// Extract final answer
			finalAnswer := step.Thought
			if step.Reflection != "" {
				finalAnswer = step.Reflection
			}

			if r.verbose {
				fmt.Printf("✅ Final Answer: %s\n", finalAnswer)
			}

			return finalAnswer, nil
		}

		// In real implementation, this is where tool execution would happen
		// For now, we simulate observation from the action
		step.Observation = fmt.Sprintf("Executed: %s", step.Action)
	}

	return "", fmt.Errorf("max iterations (%d) reached without final answer", r.maxSteps)
}

// GetReasoningHistory returns a formatted string of all reasoning steps
func (r *ReActAgent) GetReasoningHistory() string {
	var history strings.Builder

	history.WriteString("=== ReAct Reasoning History ===\n\n")

	for _, step := range r.steps {
		history.WriteString(fmt.Sprintf("Iteration %d (%s):\n",
			step.Iteration, step.Timestamp.Format("15:04:05")))
		history.WriteString(fmt.Sprintf("  💭 Thought: %s\n", step.Thought))
		history.WriteString(fmt.Sprintf("  🎯 Action: %s\n", step.Action))
		if step.Observation != "" {
			history.WriteString(fmt.Sprintf("  👁️  Observation: %s\n", step.Observation))
		}
		if step.Reflection != "" {
			history.WriteString(fmt.Sprintf("  🤔 Reflection: %s\n", step.Reflection))
		}
		history.WriteString("\n")
	}

	return history.String()
}

// SaveToMemory stores all ReAct steps to memory with proper categorization
func (r *ReActAgent) SaveToMemory(query string, answer string) error {
	if r.memory == nil {
		return nil // No memory configured
	}

	// Save summary message
	summary := fmt.Sprintf("Solved: %s\nUsed %d reasoning steps\nAnswer: %s",
		query, len(r.steps), answer)

	msg := types.Message{
		Role:    types.RoleAssistant,
		Content: summary,
		Metadata: map[string]interface{}{
			"category":     types.CategoryReasoning,
			"react_steps":  r.steps,
			"step_count":   len(r.steps),
			"query":        query,
			"final_answer": answer,
		},
	}

	return r.memory.Add(msg)
}
