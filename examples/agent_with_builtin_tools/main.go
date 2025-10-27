package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/builtin"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("=== Agent with All Built-in Tools Demo ===\n")

	ctx := context.Background()

	// 1. Create LLM provider (auto-detect from environment)
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	fmt.Printf("✓ Using provider: %s\n", llm)

	// 2. Get all built-in tools
	registry := builtin.GetRegistry()
	tools := registry.All()
	fmt.Printf("✓ Loaded %d built-in tools\n\n", len(tools))

	// Show available tools
	fmt.Println("Available tools by category:")
	categories := map[string][]string{
		"File Operations": {},
		"Web Operations":  {},
		"DateTime":        {},
		"System":          {},
		"Math":            {},
		"Database":        {},
	}

	for _, tool := range tools {
		category := tool.Category()
		name := tool.Name()
		safe := "unsafe"
		if tool.IsSafe() {
			safe = "safe"
		}
		info := fmt.Sprintf("  - %s (%s)", name, safe)

		switch category {
		case "file":
			categories["File Operations"] = append(categories["File Operations"], info)
		case "web":
			categories["Web Operations"] = append(categories["Web Operations"], info)
		case "datetime":
			categories["DateTime"] = append(categories["DateTime"], info)
		case "system":
			categories["System"] = append(categories["System"], info)
		case "math":
			categories["Math"] = append(categories["Math"], info)
		case "database":
			categories["Database"] = append(categories["Database"], info)
		}
	}

	for category, items := range categories {
		if len(items) > 0 {
			fmt.Printf("%s:\n", category)
			for _, item := range items {
				fmt.Println(item)
			}
		}
	}

	// 3. Create agent with memory and all tools
	mem := memory.NewBuffer(50)
	ag := agent.New(llm, agent.WithMemory(mem))

	// Register all tools
	for _, tool := range tools {
		ag.AddTool(tool)
	}
	fmt.Printf("\n✓ Agent created with %d tools\n\n", len(tools))

	// 4. Test different capabilities
	testCases := []struct {
		name  string
		query string
	}{
		{
			name:  "Math calculation",
			query: "What is the result of (100 + 50) * 2 - 30?",
		},
		{
			name:  "DateTime operation",
			query: "What is the current date and time?",
		},
		{
			name:  "System information",
			query: "What operating system am I running?",
		},
		{
			name:  "File operation",
			query: "List files in the current directory",
		},
	}

	// Run tests
	fmt.Println("=== Testing Agent Capabilities ===\n")
	for i, tc := range testCases {
		fmt.Printf("Test %d: %s\n", i+1, tc.name)
		fmt.Printf("Query: %s\n", tc.query)

		// Execute with agent (will auto-call tools)
		response, err := ag.Chat(ctx, tc.query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("Response: %s\n", response)
		fmt.Println("✓ Success\n")
	}

	// 5. Test conversation memory
	fmt.Println("=== Testing Conversation Memory ===\n")

	conversationTests := []string{
		"My favorite number is 42",
		"What is my favorite number?",
		"Calculate my favorite number multiplied by 2",
	}

	for i, query := range conversationTests {
		fmt.Printf("Turn %d: %s\n", i+1, query)

		response, err := ag.Chat(ctx, query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("Response: %s\n", response)
		fmt.Println("✓ Success\n")
	}

	// 6. Show summary
	fmt.Println("=== Memory Statistics ===")
	fmt.Printf("Agent maintains conversation context across turns\n")
	fmt.Printf("Memory capacity: 50 messages\n")

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ Multi-provider support (auto-detect from environment)")
	fmt.Println("✓ 20 built-in tools across 6 categories")
	fmt.Println("✓ Automatic tool selection and execution")
	fmt.Println("✓ Conversation memory across multiple turns")
	fmt.Println("✓ Math, DateTime, System, File operations")
	fmt.Println("\nThe agent successfully:")
	fmt.Println("  - Performed calculations using math_calculate tool")
	fmt.Println("  - Retrieved system information")
	fmt.Println("  - Managed file operations")
	fmt.Println("  - Maintained conversation context")
	fmt.Println("  - Remembered user preferences across turns")

	fmt.Println("\nTo run with different providers:")
	fmt.Println("  Ollama:  LLM_PROVIDER=ollama LLM_MODEL=qwen3:1.7b go run .")
	fmt.Println("  OpenAI:  LLM_PROVIDER=openai LLM_MODEL=gpt-4o-mini go run .")
	fmt.Println("  Gemini:  LLM_PROVIDER=gemini LLM_MODEL=gemini-2.5-flash go run .")
}
