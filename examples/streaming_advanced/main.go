package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
	"github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
	ctx := context.Background()

	// Create Ollama provider
	provider := ollama.New("http://localhost:11434", "llama3.2")

	// Create agent with memory
	mem := memory.NewBuffer(50)
	ag := agent.New(provider, agent.WithMemory(mem))

	fmt.Println("🌊 Advanced Streaming Example")
	fmt.Println("==============================")
	fmt.Println()

	// Simulate typing effect with streaming
	questions := []struct {
		query string
		delay time.Duration // Delay to see streaming effect
	}{
		{"What is your purpose?", 50 * time.Millisecond},
		{"My name is Alice", 0},
		{"What's my name?", 50 * time.Millisecond},
	}

	for i, q := range questions {
		fmt.Printf("\n--- Turn %d ---\n", i+1)
		fmt.Printf("User: %s\n", q.query)
		fmt.Print("Assistant: ")

		var fullResponse string
		startTime := time.Now()

		err := ag.ChatStream(ctx, q.query, func(chunk types.StreamChunk) error {
			// Accumulate full response
			fullResponse += chunk.Content

			// Print with optional delay for visual effect
			if q.delay > 0 && chunk.Content != "" {
				time.Sleep(q.delay)
			}
			fmt.Print(chunk.Content)

			// Show stats when done
			if chunk.Done {
				duration := time.Since(startTime)
				fmt.Printf("\n[⏱️  %.2fs", duration.Seconds())
				
				if chunk.Metadata != nil {
					tokensPerSec := float64(chunk.Metadata.CompletionTokens) / duration.Seconds()
					fmt.Printf(" | 🎯 %d tokens | ⚡ %.1f tokens/sec",
						chunk.Metadata.CompletionTokens,
						tokensPerSec)
				}
				fmt.Println("]")
			}

			return nil
		})

		if err != nil {
			log.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Show conversation history
	fmt.Println("\n📜 Conversation History")
	fmt.Println("======================")
	history, _ := ag.GetHistory()
	fmt.Printf("Total messages in memory: %d\n", len(history))

	fmt.Println("\n✅ Advanced streaming example completed!")
}
