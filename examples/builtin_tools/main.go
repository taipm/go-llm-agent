package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/taipm/go-llm-agent/pkg/builtin"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
	fmt.Println("=== go-llm-agent Built-in Tools Demo ===")

	// Initialize provider
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize provider: %v", err)
	}

	// Create tool registry with all built-in tools (super simple!)
	// This automatically registers all 13 built-in tools with sensible defaults
	registry := builtin.GetRegistry()

	// Alternative: Use custom configuration
	// config := builtin.DefaultConfig()
	// config.File.Base.AllowedPaths = []string{"/custom/path"}
	// config.NoWeb = true // Skip web tools
	// registry := builtin.GetRegistryWithConfig(config)

	fmt.Printf("Registered %d tools:\n", registry.Count())
	for _, name := range registry.Names() {
		tool := registry.Get(name)
		safetyStr := "safe"
		if !tool.IsSafe() {
			safetyStr = "UNSAFE"
		}
		fmt.Printf("  - %s (%s, %s): %s\n", name, tool.Category(), safetyStr, tool.Description())
	}
	fmt.Println()

	// Example 1: List current directory
	fmt.Println("Example 1: List current directory")
	demoListDirectory(registry)

	// Example 2: Get current time
	fmt.Println("\nExample 2: Get current time")
	demoCurrentTime(registry)

	// Example 3: Use with LLM provider
	fmt.Println("\nExample 3: Use with LLM provider")
	demoWithLLM(llm, registry)
}

func demoListDirectory(registry *tools.Registry) {
	ctx := context.Background()

	result, err := registry.Execute(ctx, "file_list", map[string]interface{}{
		"path":      ".",
		"recursive": false,
		"pattern":   "*.go",
	})

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Pretty print result
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
}

func demoCurrentTime(registry *tools.Registry) {
	ctx := context.Background()

	result, err := registry.Execute(ctx, "datetime_now", map[string]interface{}{
		"format":   "RFC3339",
		"timezone": "Asia/Tokyo",
	})

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
}

func demoWithLLM(llm types.LLMProvider, registry *tools.Registry) {
	ctx := context.Background()

	// Convert all tools to tool definitions
	toolDefs := registry.ToToolDefinitions()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "What time is it in Tokyo right now? Use the datetime tool.",
		},
	}

	// Chat with tools
	response, err := llm.Chat(ctx, messages, &types.ChatOptions{
		Tools: toolDefs,
	})

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Check if LLM wants to call tools
	if len(response.ToolCalls) > 0 {
		fmt.Println("LLM requested tool calls:")
		for _, tc := range response.ToolCalls {
			fmt.Printf("  Tool: %s\n", tc.Function.Name)
			fmt.Printf("  Arguments: %v\n", tc.Function.Arguments)

			// Execute the tool
			result, err := registry.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				log.Printf("Tool execution error: %v", err)
				continue
			}

			fmt.Println("  Result:")
			jsonData, _ := json.MarshalIndent(result, "    ", "  ")
			fmt.Printf("    %s\n", string(jsonData))

			// Add tool result to messages
			messages = append(messages, types.Message{
				Role:      types.RoleAssistant,
				ToolCalls: response.ToolCalls,
			})

			resultJSON, _ := json.Marshal(result)
			messages = append(messages, types.Message{
				Role:    types.RoleTool,
				Content: string(resultJSON),
				ToolID:  tc.ID,
			})
		}

		// Get final response from LLM
		finalResponse, err := llm.Chat(ctx, messages, nil)
		if err != nil {
			log.Printf("Error getting final response: %v", err)
			return
		}

		fmt.Println("\nFinal answer:")
		fmt.Println(finalResponse.Content)
	} else {
		fmt.Println("LLM response (no tool calls):")
		fmt.Println(response.Content)
	}
}

func init() {
	// Ensure we're in the example directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		if err := os.Chdir("examples/builtin_tools"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not change to examples directory: %v\n", err)
		}
	}
}
