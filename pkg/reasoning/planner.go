package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// Planner decomposes complex goals into executable sub-tasks
type Planner struct {
	provider types.LLMProvider
	memory   types.Memory
	logger   logger.Logger
	verbose  bool
}

// NewPlanner creates a new task planner
func NewPlanner(provider types.LLMProvider, memory types.Memory, log logger.Logger, verbose bool) *Planner {
	return &Planner{
		provider: provider,
		memory:   memory,
		logger:   log,
		verbose:  verbose,
	}
}

// DecomposeGoal breaks a complex goal into sequential sub-tasks
func (p *Planner) DecomposeGoal(ctx context.Context, goal string) (*types.Plan, error) {
	p.logger.Info("🎯 Decomposing goal into tasks...")
	p.logger.Debug("📝 Goal: %s", goal)

	// Construct prompt for LLM to decompose goal
	prompt := fmt.Sprintf(`You are a task planning expert. Break down this complex goal into clear, sequential steps.

Goal: %s

Requirements:
1. Create 3-7 concrete, actionable steps
2. Each step should be specific and measurable
3. Steps should be in logical execution order
4. Identify dependencies between steps (which steps must complete before others)
5. Make steps atomic (each should accomplish one clear thing)

Respond in JSON format:
{
  "steps": [
    {
      "id": "step-1",
      "description": "Clear description of what to do",
      "dependencies": []
    },
    {
      "id": "step-2", 
      "description": "Another step",
      "dependencies": ["step-1"]
    }
  ]
}

Only return valid JSON, no additional text.`, goal)

	// Call LLM to generate plan
	messages := []types.Message{
		{Role: types.RoleUser, Content: prompt},
	}

	opts := &types.ChatOptions{
		Temperature: 0.3, // Low temperature for consistent planning
		MaxTokens:   1500,
	}

	response, err := p.provider.Chat(ctx, messages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose goal: %w", err)
	}

	// Parse JSON response
	var planData struct {
		Steps []struct {
			ID           string   `json:"id"`
			Description  string   `json:"description"`
			Dependencies []string `json:"dependencies"`
		} `json:"steps"`
	}

	// Clean response (remove markdown code blocks if present)
	content := strings.TrimSpace(response.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if err := json.Unmarshal([]byte(content), &planData); err != nil {
		p.logger.Warn("Failed to parse JSON, attempting alternative parsing...")
		// Try to extract JSON from response
		startIdx := strings.Index(content, "{")
		endIdx := strings.LastIndex(content, "}")
		if startIdx >= 0 && endIdx > startIdx {
			content = content[startIdx : endIdx+1]
			if err := json.Unmarshal([]byte(content), &planData); err != nil {
				return nil, fmt.Errorf("failed to parse plan JSON: %w\nResponse: %s", err, response.Content)
			}
		} else {
			return nil, fmt.Errorf("failed to find JSON in response: %s", response.Content)
		}
	}

	// Validate plan
	if len(planData.Steps) == 0 {
		return nil, fmt.Errorf("plan has no steps")
	}

	// Create Plan structure
	plan := &types.Plan{
		ID:        uuid.New().String(),
		Goal:      goal,
		Status:    types.PlanStatusPending,
		CreatedAt: time.Now(),
		Steps:     make([]types.PlanStep, 0, len(planData.Steps)),
	}

	// Convert to PlanStep
	for _, step := range planData.Steps {
		plan.Steps = append(plan.Steps, types.PlanStep{
			ID:           step.ID,
			Description:  step.Description,
			Dependencies: step.Dependencies,
			Status:       types.PlanStatusPending,
		})
	}

	// Log plan
	if p.verbose {
		p.logger.Info("📋 Created plan with %d steps:", len(plan.Steps))
		for i, step := range plan.Steps {
			deps := "none"
			if len(step.Dependencies) > 0 {
				deps = strings.Join(step.Dependencies, ", ")
			}
			p.logger.Info("   %d. %s (dependencies: %s)", i+1, step.Description, deps)
		}
	}

	// Store plan in memory
	if p.memory != nil {
		msg := types.Message{
			Role:    types.RoleAssistant,
			Content: fmt.Sprintf("Created plan for goal: %s\n\n%s", goal, p.formatPlan(plan)),
		}
		if err := p.memory.Add(msg); err != nil {
			p.logger.Warn("Failed to store plan in memory: %v", err)
		}
	}

	return plan, nil
}

// ExecutePlan executes a plan step-by-step with dependency tracking
func (p *Planner) ExecutePlan(ctx context.Context, plan *types.Plan, executor func(context.Context, string) (interface{}, error)) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}

	p.logger.Info("▶️  Starting plan execution: %s", plan.Goal)
	plan.Status = types.PlanStatusInProgress
	plan.StartedAt = time.Now()

	// Track completed steps
	completed := make(map[string]bool)

	// Execute steps in order, respecting dependencies
	for {
		// Find next executable step
		nextStep := p.findNextStep(plan, completed)
		if nextStep == nil {
			// No more executable steps
			break
		}

		// Execute step
		p.logger.Info("🔄 Executing step: %s", nextStep.Description)
		nextStep.Status = types.PlanStatusInProgress
		nextStep.StartedAt = time.Now()

		result, err := executor(ctx, nextStep.Description)

		nextStep.CompletedAt = time.Now()

		if err != nil {
			nextStep.Status = types.PlanStatusFailed
			nextStep.Error = err
			p.logger.Error("❌ Step failed: %v", err)

			// Mark plan as failed
			plan.Status = types.PlanStatusFailed
			return fmt.Errorf("step %s failed: %w", nextStep.ID, err)
		}

		nextStep.Status = types.PlanStatusCompleted
		nextStep.Result = result
		completed[nextStep.ID] = true

		p.logger.Info("✅ Step completed: %s", nextStep.Description)
	}

	// Check if all steps completed
	allCompleted := true
	for _, step := range plan.Steps {
		if step.Status != types.PlanStatusCompleted && step.Status != types.PlanStatusSkipped {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		plan.Status = types.PlanStatusCompleted
		plan.CompletedAt = time.Now()
		p.logger.Info("✨ Plan completed successfully!")

		// Store completion in memory
		if p.memory != nil {
			msg := types.Message{
				Role:    types.RoleAssistant,
				Content: fmt.Sprintf("Completed plan: %s\n\nDuration: %v", plan.Goal, plan.CompletedAt.Sub(plan.StartedAt)),
			}
			p.memory.Add(msg)
		}
	}

	return nil
}

// findNextStep finds the next step that can be executed (all dependencies met)
func (p *Planner) findNextStep(plan *types.Plan, completed map[string]bool) *types.PlanStep {
	for i := range plan.Steps {
		step := &plan.Steps[i]

		// Skip if already completed or in progress
		if step.Status != types.PlanStatusPending {
			continue
		}

		// Check if all dependencies are met
		canExecute := true
		for _, depID := range step.Dependencies {
			if !completed[depID] {
				canExecute = false
				break
			}
		}

		if canExecute {
			return step
		}
	}

	return nil
}

// GetProgress returns current execution progress
func (p *Planner) GetProgress(plan *types.Plan) *types.PlanProgress {
	if plan == nil {
		return nil
	}

	progress := &types.PlanProgress{
		TotalSteps: len(plan.Steps),
	}

	for i := range plan.Steps {
		step := &plan.Steps[i]
		switch step.Status {
		case types.PlanStatusCompleted:
			progress.CompletedSteps++
		case types.PlanStatusFailed:
			progress.FailedSteps++
		case types.PlanStatusInProgress:
			progress.CurrentStep = step
		}
	}

	// Calculate progress percentage
	if progress.TotalSteps > 0 {
		progress.Progress = float64(progress.CompletedSteps) / float64(progress.TotalSteps)
	}

	return progress
}

// formatPlan returns a human-readable plan description
func (p *Planner) formatPlan(plan *types.Plan) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Plan: %s\n", plan.Goal))
	sb.WriteString(fmt.Sprintf("Steps:\n"))

	for i, step := range plan.Steps {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s", i+1, step.Status, step.Description))
		if len(step.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf(" (depends on: %s)", strings.Join(step.Dependencies, ", ")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// SaveToMemory stores the plan in memory with appropriate metadata
func (p *Planner) SaveToMemory(ctx context.Context, plan *types.Plan) error {
	if p.memory == nil {
		return fmt.Errorf("memory not configured")
	}

	content := p.formatPlan(plan)

	msg := types.Message{
		Role:    types.RoleAssistant,
		Content: content,
	}

	return p.memory.Add(msg)
}
