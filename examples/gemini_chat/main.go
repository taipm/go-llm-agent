package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Option 1: Use FromEnv() for automatic configuration
	llm, err := provider.FromEnv()
	if err != nil {
		// Option 2: Fall back to manual Gemini configuration
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable is required")
		}
		llm, err = provider.New(provider.Config{
			Type:   provider.ProviderGemini,
			APIKey: apiKey,
			Model:  "gemini-2.5-flash",
		})
		if err != nil {
			log.Fatalf("Failed to create provider: %v", err)
		}
	}

	fmt.Println("=== Example 1: Simple Chat ===")
	simpleChat(llm)

	fmt.Println()
	fmt.Println("=== Example 2: Streaming Chat ===")
	streamingChat(llm)

	fmt.Println()
	fmt.Println("=== Example 3: Tool Calling ===")
	toolCallingExample(llm)
}

func simpleChat(llm types.LLMProvider) {
	ctx := context.Background()

	messages := []types.Message{
		{Role: types.RoleUser, Content: "What is the capital of France?"},
	}

	response, err := llm.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	})
	if err != nil {
		log.Fatalf("Chat error: %v", err)
	}

	fmt.Printf("Response: %s\n", response.Content)
	if response.Metadata != nil {
		fmt.Printf("Tokens: %d (prompt: %d, completion: %d)\n",
			response.Metadata.TotalTokens,
			response.Metadata.PromptTokens,
			response.Metadata.CompletionTokens)
	}
}

func streamingChat(llm types.LLMProvider) {
	ctx := context.Background()

	messages := []types.Message{
		{Role: types.RoleUser, Content: "Count from 1 to 5."},
	}

	fmt.Print("Streaming: ")
	err := llm.Stream(ctx, messages, &types.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	}, func(chunk types.StreamChunk) error {
		if chunk.Content != "" {
			fmt.Print(chunk.Content)
		}
		if chunk.Done {
			fmt.Println("\n[Done]")
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Stream error: %v", err)
	}
}

func toolCallingExample(llm types.LLMProvider) {
	ctx := context.Background()

	weatherTool := types.ToolDefinition{
		Type: "function",
		Function: types.FunctionDefinition{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
			Parameters: &types.JSONSchema{
				Type: "object",
				Properties: map[string]*types.JSONSchema{
					"location": {
						Type:        "string",
						Description: "The city and country, e.g. 'Paris, France'",
					},
					"unit": {
						Type:        "string",
						Enum:        []interface{}{"celsius", "fahrenheit"},
						Description: "The temperature unit",
					},
				},
				Required: []string{"location"},
			},
		},
	}

	messages := []types.Message{
		{Role: types.RoleUser, Content: "What's the weather like in Tokyo?"},
	}

	response, err := llm.Chat(ctx, messages, &types.ChatOptions{
		Tools: []types.ToolDefinition{weatherTool},
	})
	if err != nil {
		log.Fatalf("Chat error: %v", err)
	}

	if len(response.ToolCalls) > 0 {
		fmt.Println("Model wants to call tools:")
		for _, tc := range response.ToolCalls {
			fmt.Printf("  - %s: %s(%v)\n", tc.ID, tc.Function.Name, tc.Function.Arguments)
		}

		// Simulate tool execution
		toolResults := types.Message{
			Role:    types.RoleTool,
			Content: `{"temperature": 22, "condition": "sunny", "unit": "celsius"}`,
			ToolID:  "get_weather",
		}

		messages = append(messages, types.Message{
			Role:      types.RoleAssistant,
			Content:   response.Content,
			ToolCalls: response.ToolCalls,
		}, toolResults)

		finalResponse, err := llm.Chat(ctx, messages, nil)
		if err != nil {
			log.Fatalf("Chat error: %v", err)
		}

		fmt.Printf("\nFinal response: %s\n", finalResponse.Content)
	} else {
		fmt.Printf("Response: %s\n", response.Content)
	}
}
