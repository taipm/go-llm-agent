package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
	ctx := context.Background()

	// Create Ollama provider
	provider := ollama.New("http://localhost:11434", "llama3.2")

	// Create memory to store conversation history
	mem := memory.NewBuffer(50) // Store up to 50 messages

	// Create agent with memory
	ag := agent.New(provider, agent.WithMemory(mem))

	fmt.Println("💬 Multi-turn Conversation Example")
	fmt.Println("==================================")
	fmt.Println()

	// Simulate a multi-turn conversation
	conversations := []string{
		"Hi! My name is Alice and I'm a software engineer.",
		"What's my name?",
		"What's my profession?",
		"I also love traveling to Japan.",
		"What country do I love to travel to?",
		"Can you summarize what you know about me?",
	}

	for i, message := range conversations {
		fmt.Printf("Turn %d - User: %s\n", i+1, message)

		response, err := ag.Chat(ctx, message)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Turn %d - Agent: %s\n", i+1, response)
		fmt.Println()
	}

	// Show conversation history
	history, err := ag.GetHistory()
	if err != nil {
		log.Fatalf("Failed to get history: %v", err)
	}

	fmt.Println("📜 Conversation History")
	fmt.Println("======================")
	fmt.Printf("Total messages in memory: %d\n", len(history))

	fmt.Println()
	fmt.Println("✅ Conversation example completed!")
}
