package main

import (
	"context"
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

	// Create agent with reflection enabled (default)
	// Reflection is automatically applied to all Chat() calls
	ag := agent.New(llm,
		agent.WithMinConfidence(0.7), // Require 70% confidence (optional, this is default)
	)

	fmt.Println("🤖 Self-Reflection Agent Demo")
	fmt.Println(string(make([]rune, 80)))
	fmt.Println()
	fmt.Println("Reflection is ENABLED by default for all conversations.")
	fmt.Println("The agent will automatically:")
	fmt.Println("  - Identify concerns about its answers")
	fmt.Println("  - Verify facts, calculations, and consistency")
	fmt.Println("  - Correct mistakes when confidence is low")
	fmt.Println()

	ctx := context.Background()

	// Test 1: Factual question (common mistake)
	runTest(ctx, ag, 1, "Factual Question",
		"What is the capital of Australia?",
		"Common mistake: confusing Sydney (largest city) with Canberra (capital)")

	// Test 2: Mathematical calculation
	runTest(ctx, ag, 2, "Calculation",
		"What is 156 * 73 + 48?",
		"Should use math_calculate tool for accurate result")

	fmt.Println("\n" + string(make([]rune, 80)))
	fmt.Println("\n✅ All tests completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✓ Automatic reflection on all answers")
	fmt.Println("  ✓ Fact verification using tools/LLM knowledge")
	fmt.Println("  ✓ Calculation verification with math tools")
	fmt.Println("  ✓ Consistency checking against conversation history")
	fmt.Println("  ✓ Confidence scoring (0.0 - 1.0)")
	fmt.Println("  ✓ Automatic correction when needed")
	fmt.Println()
	fmt.Println("💡 Tip: Disable reflection with: agent.WithReflection(false)")
	fmt.Println("💡 Tip: Adjust threshold with: agent.WithMinConfidence(0.9)")
}

func runTest(ctx context.Context, ag *agent.Agent, num int, category, question, note string) {
	fmt.Printf("\n� Test %d: %s\n", num, category)
	fmt.Printf("Question: %s\n", question)
	fmt.Printf("Note: %s\n", note)
	fmt.Println(string(make([]rune, 80)))

	answer, err := ag.Chat(ctx, question)
	if err != nil {
		log.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Final Answer: %s\n", answer)
	fmt.Println()
}
