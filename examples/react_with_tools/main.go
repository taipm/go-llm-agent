package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/reasoning"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// CalculatorTool performs basic math operations
type CalculatorTool struct {
	tools.BaseTool
}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{
		BaseTool: tools.NewBaseTool(
			"calculator",
			"Performs basic math calculations. Provide operation (add/subtract/multiply/divide) and two numbers a and b",
			tools.CategoryMath,
			false, // doesn't require auth
			true,  // is safe
		),
	}
}

func (c *CalculatorTool) Parameters() *types.JSONSchema {
	return nil // Simple tool, will parse params manually
}

func (c *CalculatorTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	// Get numbers (handle both float64 and string)
	var a, b float64
	var err error

	switch v := params["a"].(type) {
	case float64:
		a = v
	case string:
		a, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("parameter 'a' must be a number")
		}
	default:
		return nil, fmt.Errorf("parameter 'a' must be a number")
	}

	switch v := params["b"].(type) {
	case float64:
		b = v
	case string:
		b, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("parameter 'b' must be a number")
		}
	default:
		return nil, fmt.Errorf("parameter 'b' must be a number")
	}

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	return result, nil
}

func main() {
	// Load .env
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("⚠️  No .env file found")
	}

	fmt.Println("=== ReAct Agent with Tools Example ===\n")

	// Setup LLM
	fmt.Println("📡 Connecting to LLM...")
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("❌ Failed to create provider: %v", err)
	}
	fmt.Println("✅ LLM Provider ready\n")

	}
	fmt.Println("✅ LLM Provider ready\n")

	// Create logger
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LogLevelInfo)

	// Create ReAct agent with tools
	fmt.Println("🧠 Creating ReAct Agent with Tools...")
	reactAgent := reasoning.NewReActAgent(llm, nil, 5)
	reactAgent.WithLogger(log)
	reactAgent.WithTools(NewCalculatorTool())
	reactAgent.SetVerbose(true)
	fmt.Println("✅ ReAct Agent ready with calculator tool\n")	// Test questions that require tool usage
	questions := []string{
		"Calculate 15 * 23 + 47 using the calculator tool",
		"What is the result of 100 - 35 (25% of 100 is 25)?",
		"If I have 120 and divide by 1.5, then multiply by 0.621371, what do I get?",
	}

	ctx := context.Background()

	for i, q := range questions {
		fmt.Println("======================================================================")
		fmt.Printf("Question %d: %s\n", i+1, q)
		fmt.Println("======================================================================\n")

		answer, err := reactAgent.Solve(ctx, q)
		if err != nil {
			log.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Println("\n" + reactAgent.GetReasoningHistory())
		fmt.Printf("\n✅ Final Answer: %s\n\n", answer)
	}

	fmt.Println("======================================================================")
	fmt.Println("🎉 Demo Completed!")
	fmt.Println("======================================================================")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("   ✓ ReAct reasoning pattern (Thought → Action → Observation)")
	fmt.Println("   ✓ Tool integration and execution")
	fmt.Println("   ✓ Enhanced logging with reasoning traces")
	fmt.Println("   ✓ Iterative problem solving")
}
