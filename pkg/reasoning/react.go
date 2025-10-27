package reasoning

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// ReActAgent implements the ReAct (Reasoning + Acting) pattern
// Paper: https://arxiv.org/abs/2210.03629
type ReActAgent struct {
	provider  types.LLMProvider
	memory    types.Memory
	steps     []types.ReActStep
	maxSteps  int
	verbose   bool
}

// NewReActAgent creates a new ReAct agent
func NewReActAgent(provider types.LLMProvider, memory types.Memory, maxSteps int) *ReActAgent {
	if maxSteps <= 0 {
		maxSteps = 10 // Default max iterations
	}
	return &ReActAgent{
		provider: provider,
		memory:   memory,
		steps:    make([]types.ReActStep, 0, maxSteps),
		maxSteps: maxSteps,
		verbose:  true,
	}
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
	
	prompt.WriteString("You are a helpful AI assistant that thinks step-by-step using the ReAct pattern.\n\n")
	prompt.WriteString("ReAct means: Reasoning (think) + Acting (do)\n\n")
	prompt.WriteString("For each step, you must explicitly provide:\n")
	prompt.WriteString("1. Thought: Your reasoning about what to do next\n")
	prompt.WriteString("2. Action: The tool/function to call (or 'Answer' if you're ready to respond)\n")
	prompt.WriteString("3. Observation: What you learned from the action\n")
	prompt.WriteString("4. Reflection: What this means for solving the problem\n\n")
	
	prompt.WriteString("Example format:\n")
	prompt.WriteString("Thought: I need to check the weather to answer this question\n")
	prompt.WriteString("Action: call_tool('get_weather', {location: 'Paris'})\n")
	prompt.WriteString("Observation: Temperature is 20°C, sunny\n")
	prompt.WriteString("Reflection: Now I have the information needed to answer\n\n")
	
	// Show previous steps if any
	if len(previousSteps) > 0 {
		prompt.WriteString("Previous reasoning steps:\n")
		for _, step := range previousSteps {
			prompt.WriteString(fmt.Sprintf("\n--- Iteration %d ---\n", step.Iteration))
			prompt.WriteString(fmt.Sprintf("Thought: %s\n", step.Thought))
			prompt.WriteString(fmt.Sprintf("Action: %s\n", step.Action))
			prompt.WriteString(fmt.Sprintf("Observation: %s\n", step.Observation))
			prompt.WriteString(fmt.Sprintf("Reflection: %s\n", step.Reflection))
		}
		prompt.WriteString("\n")
	}
	
	prompt.WriteString(fmt.Sprintf("User Query: %s\n\n", query))
	prompt.WriteString("Now provide your next reasoning step:\n")
	
	return prompt.String()
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

// Think performs one iteration of ReAct reasoning
func (r *ReActAgent) Think(ctx context.Context, query string) (*types.ReActStep, error) {
	iteration := len(r.steps) + 1
	
	// Build prompt with previous steps
	prompt := r.buildReActPrompt(query, r.steps)
	
	// Call LLM
	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: prompt,
		},
	}
	
	response, err := r.provider.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}
	
	// Parse response
	thought, action, observation, reflection := r.parseReActResponse(response.Content)
	
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
