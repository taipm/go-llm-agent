package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
	"github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
	ctx := context.Background()

	// Create Ollama provider
	provider := ollama.New("http://localhost:11434", "llama3.2")

	// Create agent
	ag := agent.New(provider)

	fmt.Println("🌊 Streaming Chat Example")
	fmt.Println("=========================")
	fmt.Println()

	// Example questions for streaming
	questions := []string{
		"Write a short poem about Go programming language",
		"Explain what makes a good software engineer in 3 sentences",
		"Tell me a short story about an AI learning to code",
	}

	for i, question := range questions {
		fmt.Printf("\n--- Question %d ---\n", i+1)
		fmt.Printf("Q: %s\n", question)
		fmt.Print("A: ")

		// Stream the response
		err := ag.ChatStream(ctx, question, func(chunk types.StreamChunk) error {
			// Print content as it arrives
			fmt.Print(chunk.Content)
			
			// Print metadata when done
			if chunk.Done {
				fmt.Println()
				if chunk.Metadata != nil {
					fmt.Printf("\n[Tokens: %d prompt + %d completion = %d total]\n",
						chunk.Metadata.PromptTokens,
						chunk.Metadata.CompletionTokens,
						chunk.Metadata.TotalTokens)
				}
			}
			
			return nil
		})

		if err != nil {
			log.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
	}

	fmt.Println("\n✅ Streaming example completed!")
}
