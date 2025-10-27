package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/types"
)

const separator = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

func main() {
	// Load .env
	godotenv.Load("../.env")

	fmt.Println("=== Vector Memory Agent - Semantic Search Demo ===\n")
	fmt.Println("This demo shows how agent remembers past conversations")
	fmt.Println("and retrieves relevant context using semantic search.\n")

	// Setup LLM provider
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

	fmt.Printf("📡 Using %s with model %s\n\n", llmProvider, llmModel)

	// Create LLM provider
	llm, err := provider.New(provider.Config{
		Type:    provider.ProviderType(llmProvider),
		Model:   llmModel,
		BaseURL: baseURL,
	})
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	// Check if Qdrant is available
	fmt.Println("🔍 Checking Qdrant availability...")
	ctx := context.Background()

	// Create vector memory with Qdrant
	vectorMem, err := memory.NewVectorMemory(ctx, memory.VectorMemoryConfig{
		QdrantURL:      "localhost:6334",
		CollectionName: "agent_demo",
		Embedder:       memory.NewOllamaEmbedder(baseURL, "nomic-embed-text"),
		CacheSize:      50,
	})

	if err != nil {
		fmt.Printf("⚠️  Qdrant not available: %v\n", err)
		fmt.Println("💡 Falling back to BufferMemory (no semantic search)")
		fmt.Println("\nTo use vector memory:")
		fmt.Println("  docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant\n")

		// Fallback to buffer memory
		bufferMem := memory.NewBuffer(100)
		runDemo(ctx, llm, bufferMem, false)
		return
	}
	defer vectorMem.Close()

	fmt.Println("✅ Qdrant connected successfully!\n")

	// Run demo with vector memory
	runDemo(ctx, llm, vectorMem, true)
}

func runDemo(ctx context.Context, llm types.LLMProvider, mem types.Memory, hasVectorSearch bool) {
	// Create agent with vector memory
	ag := agent.New(llm,
		agent.WithMemory(mem),
		agent.WithLogLevel(logger.LogLevelInfo),
	)

	fmt.Println("🤖 Agent created with advanced memory\n")
	fmt.Println(separator)
	fmt.Println("PHASE 1: Teaching agent about different topics")
	fmt.Println(separator + "\n")

	// Teach agent different topics
	topics := []struct {
		question string
		context  string
	}{
		{
			question: "What is Go programming language?",
			context:  "learning about Go",
		},
		{
			question: "How does vector search work?",
			context:  "learning about vector databases",
		},
		{
			question: "What are the benefits of microservices?",
			context:  "learning about architecture patterns",
		},
	}

	for i, topic := range topics {
		fmt.Printf("📚 Topic %d: %s\n", i+1, topic.context)
		answer, err := ag.Chat(ctx, topic.question)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("💡 Learned: %s\n\n", truncate(answer, 100))
		time.Sleep(500 * time.Millisecond) // Give time for memory to process
	}

	fmt.Println("\n" + separator)
	fmt.Println("PHASE 2: Testing semantic memory recall")
	fmt.Println(separator + "\n")

	if hasVectorSearch {
		// Test semantic search
		fmt.Println("🔍 Semantic Search Test:")
		fmt.Println("Query: 'programming languages'\n")

		vectorMem := mem.(*memory.VectorMemory)
		results, err := vectorMem.SearchSemantic(ctx, "programming languages", 2)
		if err != nil {
			log.Printf("Search error: %v\n", err)
		} else {
			fmt.Printf("Found %d semantically similar conversations:\n", len(results))
			for i, msg := range results {
				fmt.Printf("  %d. %s: %s\n", i+1, msg.Role, truncate(msg.Content, 80))
			}
		}
		fmt.Println()

		// Test another semantic search
		fmt.Println("🔍 Semantic Search Test:")
		fmt.Println("Query: 'distributed systems design'\n")

		results, err = vectorMem.SearchSemantic(ctx, "distributed systems design", 2)
		if err != nil {
			log.Printf("Search error: %v\n", err)
		} else {
			fmt.Printf("Found %d semantically similar conversations:\n", len(results))
			for i, msg := range results {
				fmt.Printf("  %d. %s: %s\n", i+1, msg.Role, truncate(msg.Content, 80))
			}
		}
		fmt.Println()

		// Show memory stats
		stats, err := vectorMem.GetStats(ctx)
		if err != nil {
			log.Printf("Stats error: %v\n", err)
		} else {
			fmt.Println("📊 Memory Statistics:")
			fmt.Printf("  Total messages: %d\n", stats.TotalMessages)
			fmt.Printf("  Vector count: %d\n", stats.VectorCount)
			fmt.Println()
		}
	} else {
		fmt.Println("⚠️  Semantic search not available (using BufferMemory)")
		fmt.Println("Agent can only access recent conversation history\n")

		// Show buffer memory limitation
		history, _ := mem.GetHistory(10)
		fmt.Printf("📝 Recent history: %d messages\n", len(history))
	}

	fmt.Println("\n" + separator)
	fmt.Println("PHASE 3: Agent using memory context")
	fmt.Println(separator + "\n")

	// Ask agent to recall
	fmt.Println("💬 User: Tell me what we discussed about Go")
	answer, err := ag.Chat(ctx, "Tell me what we discussed about Go")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("🤖 Agent: %s\n\n", answer)
	}

	fmt.Println("✅ Demo Complete!\n")

	if hasVectorSearch {
		fmt.Println("✨ Key Features Demonstrated:")
		fmt.Println("  ✓ Semantic search - Find similar conversations by meaning")
		fmt.Println("  ✓ Vector embeddings - Mathematical representation of text")
		fmt.Println("  ✓ Persistent memory - Survives across sessions")
		fmt.Println("  ✓ Context retrieval - Agent recalls relevant past discussions")
	} else {
		fmt.Println("💡 With Qdrant, you would get:")
		fmt.Println("  ✓ Semantic search across all conversations")
		fmt.Println("  ✓ Automatic relevance ranking")
		fmt.Println("  ✓ Persistent storage")
		fmt.Println("  ✓ Scalable to millions of messages")
	}
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
