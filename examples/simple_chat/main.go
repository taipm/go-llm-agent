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
	// Load .env file if exists
	_ = godotenv.Load()

	ctx := context.Background()

	// Create provider from environment variables
	// This allows switching providers without code changes!
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Create agent
	ag := agent.New(llm)

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
