package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Create LLM provider
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create agent
	ag := agent.New(llm,
		agent.WithSystemPrompt("You are a helpful planning assistant."),
	)

	fmt.Println("🎯 Task Planning Agent Demo")
	fmt.Println(string(make([]rune, 80)))
	fmt.Println()
	fmt.Println("This agent can break down complex goals into executable sub-tasks.")
	fmt.Println()

	ctx := context.Background()

	// Example 1: Simple multi-step goal
	fmt.Println("📋 Example 1: Research and Report")
	fmt.Println(string(make([]rune, 80)))
	runPlanningExample(ctx, ag, "Create a comprehensive report on the benefits of Go programming language in 2025")

	// Example 2: Technical task
	fmt.Println("\n\n📋 Example 2: Project Setup")
	fmt.Println(string(make([]rune, 80)))
	runPlanningExample(ctx, ag, "Set up a new Go web service with database, authentication, and API documentation")

	// Example 3: Learning plan
	fmt.Println("\n\n📋 Example 3: Learning Plan")
	fmt.Println(string(make([]rune, 80)))
	runPlanningExample(ctx, ag, "Learn how to build microservices with Go from beginner to advanced level")

	fmt.Println("\n" + string(make([]rune, 80)))
	fmt.Println("\n✅ All planning examples completed!")
	fmt.Println()
	fmt.Println("💡 Key Features Demonstrated:")
	fmt.Println("  • Goal decomposition into actionable steps")
	fmt.Println("  • Dependency tracking between steps")
	fmt.Println("  • Sequential execution with progress monitoring")
	fmt.Println("  • Error handling and failure recovery")
}

func runPlanningExample(ctx context.Context, ag *agent.Agent, goal string) {
	fmt.Printf("Goal: %s\n\n", goal)

	// Step 1: Create plan
	fmt.Println("🔍 Step 1: Decomposing goal into tasks...")
	plan, err := ag.Plan(ctx, goal)
	if err != nil {
		log.Printf("❌ Planning failed: %v\n", err)
		return
	}

	fmt.Printf("\n📝 Created plan with %d steps:\n\n", len(plan.Steps))
	for i, step := range plan.Steps {
		deps := "none"
		if len(step.Dependencies) > 0 {
			deps = fmt.Sprintf("%v", step.Dependencies)
		}
		fmt.Printf("   %d. [%s] %s\n", i+1, step.ID, step.Description)
		fmt.Printf("      Dependencies: %s\n", deps)
	}

	// Show plan as JSON
	fmt.Println("\n📄 Plan JSON:")
	jsonData, _ := json.MarshalIndent(plan, "   ", "  ")
	fmt.Printf("   %s\n", string(jsonData))

	// Step 2: Execute plan (just show progress, don't actually execute to save time)
	fmt.Println("\n⚙️  Step 2: Execution simulation...")
	fmt.Println("   (In production, use agent.ExecutePlan(ctx, plan) to execute)")
	fmt.Println()

	// Simulate progress
	progress := ag.GetPlanProgress(plan)
	fmt.Printf("   Progress: %d/%d steps (%.0f%%)\n",
		progress.CompletedSteps,
		progress.TotalSteps,
		progress.Progress*100)

	if progress.CurrentStep != nil {
		fmt.Printf("   Current: %s\n", progress.CurrentStep.Description)
	}

	fmt.Println("\n✨ Planning complete!")
}
