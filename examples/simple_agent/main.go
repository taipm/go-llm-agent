package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider"
)

func main() {
	// Load .env from parent directory
	godotenv.Load("../.env")

	fmt.Println("=== Simple Agent with Auto-Reasoning ===\n")

	// Get LLM configuration from environment
	llmProvider := os.Getenv("LLM_PROVIDER")
	llmModel := os.Getenv("LLM_MODEL")
	baseURL := os.Getenv("OLLAMA_BASE_URL")

	if llmProvider == "" {
		llmProvider = "ollama"
	}
	if llmModel == "" {
		llmModel = "qwen3:1.7b"
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	fmt.Printf("📡 Using %s with model %s\n", llmProvider, llmModel)

	// Create LLM provider
	llm, err := provider.New(provider.Config{
		Type:    provider.ProviderType(llmProvider),
		Model:   llmModel,
		BaseURL: baseURL,
	})
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}
	fmt.Println("✅ LLM Provider ready\n")

	// Create agent - everything auto-configured!
	// ✅ Auto-reasoning enabled (CoT, ReAct, Simple)
	// ✅ 25+ builtin tools loaded
	// ✅ DEBUG logging for detailed reasoning steps
	// ✅ Memory with 100 messages
	fmt.Println("🤖 Creating Agent...")
	ag := agent.New(llm)

	fmt.Printf("✅ Agent ready with %d builtin tools\n", ag.ToolCount())
	fmt.Println("✅ Auto-reasoning: ENABLED\n")

	// Test questions demonstrating different reasoning modes
	questions := []struct {
		query    string
		expected string
	}{
		{
			query:    "What is 15 * 23 + 47?",
			expected: "CoT (math calculation)",
		},
		{
			query:    "Hello, how are you?",
			expected: "Simple (greeting)",
		},
		{
			query:    "Calculate the compound interest on $1000 at 5% for 3 years",
			expected: "CoT (multi-step math)",
		},
		{
			query:    "Use calculator to compute 156 * 73",
			expected: "ReAct (tool: calculator)",
		},
		{
			query:    "Search the web for latest Go programming news",
			expected: "ReAct (tool: web_search)",
		},
	}

	ctx := context.Background()

	for i, test := range questions {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Question %d: %s\n", i+1, test.query)
		fmt.Printf("Expected Mode: %s\n", test.expected)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		answer, err := ag.Chat(ctx, test.query)
		if err != nil {
			log.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\n✅ Answer: %s\n\n", answer)
	}

	fmt.Println("\n=== Demo Complete ===")
}
