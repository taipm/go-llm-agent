package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/reasoning"
	"github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  No .env file found, using environment variables")
	}
	fmt.Println("=== ReAct Agent with Vector Memory Example ===\n")

	// 1. Setup LLM Provider (Ollama)
	fmt.Println("📡 Connecting to Ollama...")
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatal("Failed to create provider:", err)
	}
	fmt.Println("✅ LLM Provider ready\n")

	// 2. Setup Vector Memory with Qdrant (optional)
	var mem types.Memory
	useVectorMemory := os.Getenv("USE_VECTOR_MEMORY") == "true"

	if useVectorMemory {
		fmt.Println("🧠 Setting up Vector Memory with Qdrant...")
		ctx := context.Background()

		// Get Qdrant configuration from environment
		qdrantURL := os.Getenv("QDRANT_URL")
		if qdrantURL == "" {
			qdrantURL = "localhost:6334"
		}
		collectionName := os.Getenv("QDRANT_COLLECTION")
		if collectionName == "" {
			collectionName = "react_agent_demo"
		}
		embeddingModel := os.Getenv("EMBEDDING_MODEL")
		if embeddingModel == "" {
			embeddingModel = "nomic-embed-text:latest"
		}

		vectorMem, err := memory.NewVectorMemory(ctx, memory.VectorMemoryConfig{
			QdrantURL:      qdrantURL,
			CollectionName: collectionName,
			Embedder:       memory.NewOllamaEmbedder("", embeddingModel),
			CacheSize:      50,
		})
		if err != nil {
			log.Printf("⚠️  Failed to setup Vector Memory (falling back to Buffer): %v\n", err)
			mem = memory.NewBuffer(100)
		} else {
			mem = vectorMem
			defer vectorMem.Close()
			fmt.Println("✅ Vector Memory ready (semantic search enabled)\n")
		}
	} else {
		fmt.Println("🧠 Using simple Buffer Memory")
		mem = memory.NewBuffer(100)
		fmt.Println("💡 Tip: Set USE_VECTOR_MEMORY=true to enable semantic search\n")
	}

	// 3. Create ReAct Agent
	fmt.Println("🤖 Creating ReAct Agent...")
	agent := reasoning.NewReActAgent(llm, mem, 5)
	agent.SetVerbose(true)
	fmt.Println("✅ ReAct Agent ready\n")

	// 4. Test Questions
	questions := []string{
		"What is 15 * 23 + 47?",
		"Explain the Pythagorean theorem in simple terms",
		"If I have 100 dollars and spend 35% on food, how much do I have left?",
	}

	for i, question := range questions {
		fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
		fmt.Printf("Question %d: %s\n", i+1, question)
		fmt.Printf(strings.Repeat("=", 70) + "\n\n")

		ctx := context.Background()

		// Use ReAct to solve
		answer, err := agent.Solve(ctx, question)
		if err != nil {
			log.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("\n" + strings.Repeat("-", 70) + "\n")
		fmt.Printf("✅ Final Answer: %s\n", answer)
		fmt.Printf(strings.Repeat("-", 70) + "\n")

		// Save reasoning to memory
		err = agent.SaveToMemory(question, answer)
		if err != nil {
			log.Printf("⚠️  Failed to save to memory: %v\n", err)
		}

		// Show reasoning history
		fmt.Println("\n📝 Reasoning History:")
		fmt.Println(agent.GetReasoningHistory())
	}

	// 5. Test Vector Memory Search (if enabled)
	if useVectorMemory {
		fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
		fmt.Println("🔍 Testing Semantic Search in Vector Memory")
		fmt.Printf(strings.Repeat("=", 70) + "\n\n")

		if vectorMem, ok := mem.(types.AdvancedMemory); ok {
			ctx := context.Background()

			// Search for math-related memories
			fmt.Println("Searching for: 'mathematics calculations'")
			results, err := vectorMem.SearchSemantic(ctx, "mathematics calculations", 3)
			if err != nil {
				log.Printf("Search error: %v\n", err)
			} else {
				fmt.Printf("Found %d related memories:\n\n", len(results))
				for i, msg := range results {
					fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, truncate(msg.Content, 100))
				}
			}

			// Get memory stats
			fmt.Println("\n📊 Memory Statistics:")
			stats, err := vectorMem.GetStats(ctx)
			if err != nil {
				log.Printf("Stats error: %v\n", err)
			} else {
				fmt.Printf("  Total Messages: %d\n", stats.TotalMessages)
				fmt.Printf("  Vector Count: %d\n", stats.VectorCount)
			}
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Println("🎉 Demo completed!")
	fmt.Printf(strings.Repeat("=", 70) + "\n")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
