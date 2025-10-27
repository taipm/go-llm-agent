package main

import (
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

	fmt.Println("🔍 Agent Status Inspection Demo")
	fmt.Println(string(make([]rune, 80)))
	fmt.Println()

	// Scenario 1: Default agent
	fmt.Println("📊 Scenario 1: Default Agent Configuration")
	fmt.Println(string(make([]rune, 80)))
	ag1 := agent.New(llm)
	printStatus(ag1)

	// Scenario 2: Customized agent
	fmt.Println("\n📊 Scenario 2: Customized Agent")
	fmt.Println(string(make([]rune, 80)))
	ag2 := agent.New(llm,
		agent.WithSystemPrompt("You are a math expert."),
		agent.WithTemperature(0.3),
		agent.WithMaxTokens(1000),
		agent.WithReflection(false), // Disable reflection
		agent.WithMinConfidence(0.9),
		agent.WithoutAutoReasoning(), // Disable auto-reasoning
	)
	printStatus(ag2)

	// Scenario 3: Agent with minimal tools
	fmt.Println("\n📊 Scenario 3: Minimal Tools Agent")
	fmt.Println(string(make([]rune, 80)))
	ag3 := agent.New(llm,
		agent.WithoutBuiltinTools(), // No builtin tools
	)
	printStatus(ag3)

	fmt.Println("\n" + string(make([]rune, 80)))
	fmt.Println("\n✅ Status inspection completed!")
	fmt.Println()
	fmt.Println("💡 Use agent.Status() to:")
	fmt.Println("  • Debug agent configuration")
	fmt.Println("  • Monitor runtime state")
	fmt.Println("  • Verify capabilities before execution")
	fmt.Println("  • Generate system reports")
}

func printStatus(ag *agent.Agent) {
	status := ag.Status()

	fmt.Println("\n🎛️  Configuration:")
	fmt.Printf("   System Prompt: %s\n", truncate(status.Configuration.SystemPrompt, 50))
	fmt.Printf("   Temperature: %.2f\n", status.Configuration.Temperature)
	fmt.Printf("   Max Tokens: %d\n", status.Configuration.MaxTokens)
	fmt.Printf("   Max Iterations: %d\n", status.Configuration.MaxIterations)
	fmt.Printf("   Min Confidence: %.2f\n", status.Configuration.MinConfidence)
	fmt.Printf("   Reflection Enabled: %v\n", status.Configuration.EnableReflection)

	fmt.Println("\n🧠 Reasoning Capabilities:")
	fmt.Printf("   Auto-Reasoning: %v\n", status.Reasoning.AutoReasoningEnabled)
	fmt.Printf("   CoT Available: %v\n", status.Reasoning.CoTAvailable)
	fmt.Printf("   ReAct Available: %v\n", status.Reasoning.ReActAvailable)
	fmt.Printf("   Reflection Available: %v\n", status.Reasoning.ReflectionAvailable)

	fmt.Println("\n🔧 Tools:")
	fmt.Printf("   Total Count: %d\n", status.Tools.TotalCount)
	if status.Tools.TotalCount > 0 {
		fmt.Printf("   Available: %s\n", formatToolList(status.Tools.ToolNames))
	}

	fmt.Println("\n💾 Memory:")
	fmt.Printf("   Type: %s\n", status.Memory.Type)
	fmt.Printf("   Message Count: %d\n", status.Memory.MessageCount)
	fmt.Printf("   Supports Search: %v\n", status.Memory.SupportsSearch)
	fmt.Printf("   Supports Vectors: %v\n", status.Memory.SupportsVectors)

	fmt.Println("\n🤖 Provider:")
	fmt.Printf("   Type: %s\n", status.Provider.Type)

	// JSON output (optional)
	fmt.Println("\n📄 JSON Format:")
	jsonData, _ := json.MarshalIndent(status, "   ", "  ")
	fmt.Printf("   %s\n", string(jsonData))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func formatToolList(tools []string) string {
	if len(tools) == 0 {
		return "none"
	}
	if len(tools) <= 5 {
		return fmt.Sprintf("%v", tools)
	}
	// Show first 5 and count
	return fmt.Sprintf("%v ... (+%d more)", tools[:5], len(tools)-5)
}
