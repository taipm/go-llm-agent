package reasoning

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// CoTAgent implements Chain-of-Thought reasoning
// Reference: https://arxiv.org/abs/2201.11903
type CoTAgent struct {
	provider types.LLMProvider
	memory   types.Memory
	registry *tools.Registry
	maxSteps int
	verbose  bool
	logger   logger.Logger

	// Current reasoning chain
	chain *types.CoTChain
}

// NewCoTAgent creates a new Chain-of-Thought agent
func NewCoTAgent(provider types.LLMProvider, memory types.Memory, maxSteps int) *CoTAgent {
	if maxSteps <= 0 {
		maxSteps = 10 // Default max steps
	}
	return &CoTAgent{
		provider: provider,
		memory:   memory,
		registry: tools.NewRegistry(),
		maxSteps: maxSteps,
		verbose:  false,
		logger:   logger.NewConsoleLogger(),
	}
}

// WithTools adds tools to the agent
func (c *CoTAgent) WithTools(toolList ...tools.Tool) *CoTAgent {
	for _, tool := range toolList {
		c.registry.Register(tool)
	}
	return c
}

// WithLogger sets the logger
func (c *CoTAgent) WithLogger(log logger.Logger) *CoTAgent {
	c.logger = log
	return c
}

// SetVerbose enables/disables verbose output
func (c *CoTAgent) SetVerbose(verbose bool) {
	c.verbose = verbose
}

// Think performs Chain-of-Thought reasoning on a question
func (c *CoTAgent) Think(ctx context.Context, question string) (string, error) {
	// Initialize new chain
	c.chain = &types.CoTChain{
		Query:     question,
		Steps:     make([]types.CoTStep, 0),
		StartTime: time.Now(),
	}

	if c.logger != nil {
		c.logger.Debug("🧠 Starting Chain-of-Thought reasoning")
		c.logger.Debug("📝 Question: %s", question)
	}

	// Build CoT prompt
	prompt := c.buildCoTPrompt(question)

	// Get LLM response
	messages := []types.Message{
		{Role: "user", Content: prompt},
	}

	if c.logger != nil {
		c.logger.Debug("🤖 Calling LLM for CoT reasoning...")
	}

	response, err := c.provider.Chat(ctx, messages, nil)
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse CoT steps
	steps, finalAnswer, err := c.parseCoTResponse(response.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse CoT response: %w", err)
	}

	// Validate steps
	if len(steps) == 0 {
		return "", fmt.Errorf("no reasoning steps found in response")
	}

	if len(steps) > c.maxSteps {
		return "", fmt.Errorf("exceeded max steps (%d > %d)", len(steps), c.maxSteps)
	}

	// Log reasoning steps
	if c.logger != nil {
		c.logger.Info("💭 Chain-of-Thought Steps:")
		for i, step := range steps {
			c.logger.Info("   Step %d: %s", i+1, step.Description)
			if step.Reasoning != "" {
				c.logger.Debug("      Reasoning: %s", step.Reasoning)
			}
		}
		c.logger.Info("✅ Final Answer: %s", finalAnswer)
	}

	// Update chain
	c.chain.Steps = steps
	c.chain.Answer = finalAnswer
	c.chain.EndTime = time.Now()
	c.chain.CreatedAt = time.Now()

	// Save to memory if available
	if c.memory != nil {
		if err := c.SaveToMemory(ctx); err != nil && c.verbose {
			fmt.Printf("⚠️  Failed to save to memory: %v\n", err)
		}
	}

	return finalAnswer, nil
}

// buildCoTPrompt creates a prompt that encourages step-by-step reasoning
func (c *CoTAgent) buildCoTPrompt(question string) string {
	var sb strings.Builder

	sb.WriteString("You are a helpful assistant that solves problems using step-by-step reasoning.\n\n")
	sb.WriteString("When answering questions, break down your thinking into clear steps.\n")
	sb.WriteString("Use the following format:\n\n")
	sb.WriteString("Step 1: [First reasoning step with explanation]\n")
	sb.WriteString("Step 2: [Second reasoning step with explanation]\n")
	sb.WriteString("...\n")
	sb.WriteString("Step N: [Final reasoning step]\n\n")
	sb.WriteString("Answer: [Your final answer based on the reasoning above]\n\n")

	// Add few-shot examples for better results
	sb.WriteString("Example:\n")
	sb.WriteString("Question: If a store has 15 apples and sells 40% of them, how many apples are left?\n\n")
	sb.WriteString("Step 1: Calculate 40% of 15 apples = 15 × 0.40 = 6 apples sold\n")
	sb.WriteString("Step 2: Subtract sold apples from total = 15 - 6 = 9 apples remaining\n\n")
	sb.WriteString("Answer: 9 apples are left in the store.\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("Now solve this question:\n\n")
	sb.WriteString(fmt.Sprintf("Question: %s\n\n", question))
	sb.WriteString("Think step by step:\n")

	return sb.String()
}

// parseCoTResponse extracts reasoning steps and final answer from LLM response
func (c *CoTAgent) parseCoTResponse(response string) ([]types.CoTStep, string, error) {
	steps := make([]types.CoTStep, 0)
	finalAnswer := ""

	// Split by lines
	lines := strings.Split(response, "\n")

	// Regex patterns
	stepPattern := regexp.MustCompile(`(?i)^Step\s+(\d+):\s*(.+)$`)
	answerPattern := regexp.MustCompile(`(?i)^Answer:\s*(.+)$`)

	stepNumber := 0
	var currentStepText strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for step
		if matches := stepPattern.FindStringSubmatch(line); matches != nil {
			// Save previous step if exists
			if stepNumber > 0 && currentStepText.Len() > 0 {
				steps = append(steps, types.CoTStep{
					StepNumber:  stepNumber,
					Description: strings.TrimSpace(currentStepText.String()),
					Timestamp:   time.Now(),
				})
				currentStepText.Reset()
			}

			// Start new step
			stepNumber++
			currentStepText.WriteString(matches[2])
			continue
		}

		// Check for answer
		if matches := answerPattern.FindStringSubmatch(line); matches != nil {
			// Save last step if exists
			if stepNumber > 0 && currentStepText.Len() > 0 {
				steps = append(steps, types.CoTStep{
					StepNumber:  stepNumber,
					Description: strings.TrimSpace(currentStepText.String()),
					Timestamp:   time.Now(),
				})
			}

			finalAnswer = strings.TrimSpace(matches[1])
			// Continue reading to capture multi-line answers
			continue
		}

		// If we're reading an answer, append to it
		if finalAnswer != "" {
			finalAnswer += " " + line
		} else if stepNumber > 0 {
			// Continue previous step
			currentStepText.WriteString(" ")
			currentStepText.WriteString(line)
		}
	}

	// Save last step if not saved yet
	if stepNumber > 0 && currentStepText.Len() > 0 && len(steps) < stepNumber {
		steps = append(steps, types.CoTStep{
			StepNumber:  stepNumber,
			Description: strings.TrimSpace(currentStepText.String()),
			Timestamp:   time.Now(),
		})
	}

	// If no structured steps found, try to extract from free-form text
	if len(steps) == 0 {
		// Look for implicit steps (numbered lists, etc.)
		numberedPattern := regexp.MustCompile(`(?i)^(\d+)\.\s*(.+)$`)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if matches := numberedPattern.FindStringSubmatch(line); matches != nil {
				stepNum := len(steps) + 1
				steps = append(steps, types.CoTStep{
					StepNumber:  stepNum,
					Description: matches[2],
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// If still no steps, use the whole response as a single step
	if len(steps) == 0 && response != "" {
		steps = append(steps, types.CoTStep{
			StepNumber:  1,
			Description: response,
			Timestamp:   time.Now(),
		})
		finalAnswer = response
	}

	return steps, strings.TrimSpace(finalAnswer), nil
}

// GetChain returns the current reasoning chain
func (c *CoTAgent) GetChain() *types.CoTChain {
	return c.chain
}

// GetReasoningHistory returns a formatted string of the reasoning process
func (c *CoTAgent) GetReasoningHistory() string {
	if c.chain == nil {
		return "No reasoning history available."
	}

	var sb strings.Builder

	sb.WriteString("=== Chain-of-Thought Reasoning ===\n\n")
	sb.WriteString(fmt.Sprintf("Question: %s\n\n", c.chain.Query))

	for _, step := range c.chain.Steps {
		sb.WriteString(fmt.Sprintf("Step %d: %s\n", step.StepNumber, step.Description))
	}

	sb.WriteString(fmt.Sprintf("\n✅ Final Answer: %s\n", c.chain.Answer))

	duration := c.chain.EndTime.Sub(c.chain.StartTime)
	sb.WriteString(fmt.Sprintf("\n⏱️  Time taken: %v\n", duration))

	return sb.String()
}

// SaveToMemory saves the CoT chain to memory
func (c *CoTAgent) SaveToMemory(ctx context.Context) error {
	if c.chain == nil {
		return fmt.Errorf("no chain to save")
	}

	// Save question
	questionMsg := types.Message{
		Role:    "user",
		Content: c.chain.Query,
	}
	if err := c.memory.Add(questionMsg); err != nil {
		return fmt.Errorf("failed to save question: %w", err)
	}

	// Save each step
	for _, step := range c.chain.Steps {
		stepMsg := types.Message{
			Role:    "assistant",
			Content: fmt.Sprintf("CoT Step %d: %s", step.StepNumber, step.Description),
			Metadata: map[string]interface{}{
				"category": types.CategoryReasoning,
				"cot_step": step,
			},
		}
		if err := c.memory.Add(stepMsg); err != nil {
			return fmt.Errorf("failed to save step %d: %w", step.StepNumber, err)
		}
	}

	// Save final answer
	answerMsg := types.Message{
		Role:    "assistant",
		Content: fmt.Sprintf("Solved: %s\nUsed %d reasoning steps\nAnswer: %s", c.chain.Query, len(c.chain.Steps), c.chain.Answer),
		Metadata: map[string]interface{}{
			"category":  types.CategoryReasoning,
			"cot_chain": c.chain,
		},
	}
	if err := c.memory.Add(answerMsg); err != nil {
		return fmt.Errorf("failed to save answer: %w", err)
	}

	return nil
}

// Validate checks if the reasoning chain is logically sound
func (c *CoTAgent) Validate() (bool, []string) {
	if c.chain == nil {
		return false, []string{"No reasoning chain available"}
	}

	issues := make([]string, 0)

	// Check if we have steps
	if len(c.chain.Steps) == 0 {
		issues = append(issues, "No reasoning steps found")
	}

	// Check step numbering
	for i, step := range c.chain.Steps {
		expectedNum := i + 1
		if step.StepNumber != expectedNum {
			issues = append(issues, fmt.Sprintf("Step numbering issue: expected %d, got %d", expectedNum, step.StepNumber))
		}

		// Check if step has content
		if strings.TrimSpace(step.Description) == "" {
			issues = append(issues, fmt.Sprintf("Step %d is empty", step.StepNumber))
		}
	}

	// Check if we have a final answer
	if strings.TrimSpace(c.chain.Answer) == "" {
		issues = append(issues, "No final answer provided")
	}

	return len(issues) == 0, issues
}
