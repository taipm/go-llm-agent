package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/tools/datetime"
	"github.com/taipm/go-llm-agent/pkg/tools/file"
	"github.com/taipm/go-llm-agent/pkg/tools/web"
	"github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
	fmt.Println("=== go-llm-agent Built-in Tools Demo ===\n")

	// Initialize provider
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize provider: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Register file tools (read, list, write, delete)
	fileConfig := file.Config{
		AllowedPaths:  []string{".", "/tmp", os.TempDir()},
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		AllowSymlinks: false,
	}
	registry.Register(file.NewReadTool(fileConfig))
	registry.Register(file.NewListTool(fileConfig))

	// Register write tool with backup enabled
	writeConfig := file.WriteConfig{
		Config:       fileConfig,
		CreateDirs:   true,
		Backup:       true,
		BackupSuffix: ".bak",
	}
	registry.Register(file.NewWriteTool(writeConfig))

	// Register delete tool with safety restrictions
	deleteConfig := file.DeleteConfig{
		Config:              fileConfig,
		ProtectedPaths:      file.DefaultDeleteConfig.ProtectedPaths,
		AllowRecursive:      true,
		RequireConfirmation: true,
	}
	registry.Register(file.NewDeleteTool(deleteConfig))

	// Register datetime tools
	registry.Register(datetime.NewNowTool())
	registry.Register(datetime.NewFormatTool())
	registry.Register(datetime.NewCalcTool())

	// Register web tools
	webConfig := web.Config{
		Timeout:         30 * time.Second,
		MaxResponseSize: 1024 * 1024, // 1MB
		UserAgent:       "GoLLMAgent-Demo/1.0",
		AllowPrivateIPs: false,
	}
	registry.Register(web.NewFetchTool(webConfig))

	postConfig := web.PostConfig{
		Timeout:         30 * time.Second,
		MaxResponseSize: 1024 * 1024,
		UserAgent:       "GoLLMAgent-Demo/1.0",
		AllowPrivateIPs: false,
	}
	registry.Register(web.NewPostTool(postConfig))

	scrapeConfig := web.ScrapeConfig{
		Timeout:         30 * time.Second,
		MaxResponseSize: 5 * 1024 * 1024, // 5MB for HTML
		UserAgent:       "GoLLMAgent-Demo/1.0",
		AllowPrivateIPs: false,
		RateLimit:       1 * time.Second, // 1 second between requests
	}
	registry.Register(web.NewScrapeTool(scrapeConfig))

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
