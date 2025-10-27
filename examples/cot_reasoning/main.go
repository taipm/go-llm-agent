package main
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/reasoning"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("⚠️  No .env file found, using environment variables")
	}

	fmt.Println("=== Chain-of-Thought Reasoning Example ===\n")

	// 1. Setup LLM Provider
	fmt.Println("📡 Connecting to LLM...")
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("❌ Failed to create provider: %v", err)
	}
	fmt.Println("✅ LLM Provider ready\n")

	// 2. Create CoT Agent
	fmt.Println("🧠 Creating Chain-of-Thought Agent...")
	agent := reasoning.NewCoTAgent(llm, nil, 10) // No memory needed for demo
	agent.SetVerbose(true)
	fmt.Println("✅ CoT Agent ready\n")

	// 3. Test Questions
	questions := []struct {
		title    string
		question string
	}{
		{
			title:    "Math Problem",
			question: "If a train travels 120 kilometers in 1.5 hours, what is its speed in miles per hour? (1 km = 0.621371 miles)",
		},
		{
			title:    "Logic Puzzle",
			question: "Sarah is 3 times as old as Tom. In 5 years, she will be twice as old as Tom. How old are they now?",
		},
		{
			title:    "Word Problem",
			question: "A store has a 25% off sale. If an item costs $80 after the discount, what was its original price?",
		},
		{
			title:    "Complex Calculation",
			question: "Calculate the compound interest earned on $1000 invested at 5% annual rate for 3 years, compounded annually.",
		},
	}

	ctx := context.Background()

	for i, q := range questions {
		fmt.Println("======================================================================")
		fmt.Printf("Question %d: %s\n", i+1, q.title)
		fmt.Println("======================================================================")
		fmt.Printf("\n📝 %s\n\n", q.question)

		// Think through the problem
		answer, err := agent.Think(ctx, q.question)
		if err != nil {
			log.Printf("❌ Error: %v\n\n", err)
			continue
		}

		// Show reasoning history
		fmt.Println("\n" + agent.GetReasoningHistory())

		// Validate reasoning
		valid, issues := agent.Validate()
		if valid {
			fmt.Println("✅ Reasoning is logically sound")
		} else {
			fmt.Printf("⚠️  Validation issues found:\n")
			for _, issue := range issues {
				fmt.Printf("   - %s\n", issue)
			}
		}

		fmt.Printf("\n📊 Final Answer: %s\n", answer)
		fmt.Println()
	}

	fmt.Println("======================================================================")
	fmt.Println("🎉 Chain-of-Thought Demo Completed!")
	fmt.Println("======================================================================")
	fmt.Println("\n💡 Key Benefits of CoT:")
	fmt.Println("   ✓ Transparent reasoning process")
	fmt.Println("   ✓ Fewer calculation errors")
	fmt.Println("   ✓ Better explanation quality")
	fmt.Println("   ✓ Easier to verify logic")
	fmt.Println("   ✓ Improved performance on complex problems")
}
