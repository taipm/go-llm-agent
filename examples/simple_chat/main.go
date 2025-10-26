package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
	ctx := context.Background()

	// Create Ollama provider
	provider := ollama.New("http://localhost:11434", "qwen3:1.7b")

	// Create agent
	ag := agent.New(provider)

	fmt.Println("🤖 Simple Chat Example")
	fmt.Println("=====================")
	fmt.Println()

	// Example questions
	questions := []string{
		"What is the capital of France?",
		"Explain what is Go programming language in one sentence.",
		"What is 15 + 27?",
	}

	for i, question := range questions {
		fmt.Printf("Question %d: %s\n", i+1, question)

		response, err := ag.Chat(ctx, question)
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}

		fmt.Printf("Answer: %s\n", response)
		fmt.Println()
	}

	fmt.Println("✅ Simple chat example completed!")
}
