package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/examples/tools"
	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
	ctx := context.Background()

	// Create Ollama provider
	provider := ollama.New("http://localhost:11434", "llama3.2")

	// Create agent
	ag := agent.New(provider)

	// Add tools
	calculator := &tools.CalculatorTool{}
	weather := &tools.WeatherTool{}

	if err := ag.AddTool(calculator); err != nil {
		log.Fatalf("Failed to add calculator tool: %v", err)
	}

	if err := ag.AddTool(weather); err != nil {
		log.Fatalf("Failed to add weather tool: %v", err)
	}

	fmt.Println("🔧 Tool Usage Example")
	fmt.Println("====================")
	fmt.Println()

	// Example questions that require tools
	questions := []string{
		"What is the square root of 144?",
		"What's the weather like in Tokyo?",
		"Calculate 25 multiplied by 4, then tell me the result",
		"What's the temperature in London?",
	}

	for i, question := range questions {
		fmt.Printf("Question %d: %s\n", i+1, question)

		response, err := ag.Chat(ctx, question)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Answer: %s\n", response)
		fmt.Println()
	}

	fmt.Println("✅ Tool usage example completed!")
}
