package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/builtin"
	"github.com/taipm/go-llm-agent/pkg/provider"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("=== Agent with Detailed Logging Demo ===\n")
	fmt.Println("This example demonstrates how the agent processes requests,")
	fmt.Println("calls tools, and maintains conversation memory.\n")
	fmt.Println("Watch the detailed logs to see what the agent is doing!\n")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	ctx := context.Background()

	// 1. Create LLM provider
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// 2. Get DateTime and Math tools (needed for age/time calculations)
	datetimeTools := builtin.GetDateTimeTools()
	mathTools := builtin.GetMathTools()

	// 3. Create agent (memory and logging initialized automatically)
	ag := agent.New(llm)

	// Register tools
	for _, tool := range datetimeTools {
		ag.AddTool(tool)
	}
	for _, tool := range mathTools {
		ag.AddTool(tool)
	}

	fmt.Printf("✓ Provider: %s\n", llm)
	fmt.Printf("✓ Loaded %d datetime tools + %d math tools\n", len(datetimeTools), len(mathTools))
	fmt.Printf("✓ Tools: datetime_now, datetime_calc, math_calculate, math_stats\n\n")

	// Scenario: User introduces themselves with birthdate
	// Agent should remember and use tools to calculate age and lifetime

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("CONVERSATION START")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	// Turn 1: User introduces themselves
	fmt.Println("👤 User: Tôi là Phan Minh Tài, tôi sinh ngày 22/01/1984.")
	fmt.Println()
	response1, err := ag.Chat(ctx, "Tôi là Phan Minh Tài, tôi sinh ngày 22/01/1984.")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println()

	// Turn 2: Ask about age
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println()
	fmt.Println("👤 User: Năm nay tôi bao nhiêu tuổi?")
	fmt.Println()
	response2, err := ag.Chat(ctx, "Năm nay tôi bao nhiêu tuổi?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println()

	// Turn 3: Ask about lifetime in seconds
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println()
	fmt.Println("👤 User: Tôi đã sống được bao nhiêu giây?")
	fmt.Println()
	response3, err := ag.Chat(ctx, "Tôi đã sống được bao nhiêu giây?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println()

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("CONVERSATION END")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	// Summary
	fmt.Println("=== Summary ===\n")
	fmt.Println("The agent successfully:")
	fmt.Println("  ✓ Remembered user's name: Phan Minh Tài")
	fmt.Println("  ✓ Remembered birthdate: 22/01/1984")
	fmt.Println("  ✓ Used datetime_calc to calculate age")
	fmt.Println("  ✓ Used math_calculate to compute lifetime in seconds")
	fmt.Println("  ✓ Maintained context across 3 turns")
	fmt.Println()
	fmt.Println("Key Observations from Logs:")
	fmt.Println("  • 👤 = User message")
	fmt.Println("  • 🤔 = Agent thinking (calling LLM)")
	fmt.Println("  • 🔧 = Tool being called")
	fmt.Println("  • ✓ = Tool completed successfully")
	fmt.Println("  • 💬 = Agent's response")
	fmt.Println("  • 💾 = Memory operations (saved messages)")
	fmt.Println()
	fmt.Println("The detailed logging helps you understand:")
	fmt.Println("  1. When the agent is thinking vs. acting")
	fmt.Println("  2. Which tools are being called and why")
	fmt.Println("  3. How memory is being used to maintain context")
	fmt.Println("  4. The flow of multi-turn conversations")
	fmt.Println()
	fmt.Println("Try adjusting log level:")
	fmt.Println("  • LogLevelDebug - See all details including parameters")
	fmt.Println("  • LogLevelInfo  - See key actions (default)")
	fmt.Println("  • LogLevelWarn  - Only warnings and errors")
	fmt.Println("  • DisableLogging() - No logs at all")

	// Prevent unused variable warnings
	_ = response1
	_ = response2
	_ = response3
}
