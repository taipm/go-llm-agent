package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

const (
	qdrantURL      = "localhost:6334"
	ollamaURL      = "http://localhost:11434"
	modelName      = "qwen2.5:7b"
	embedModel     = "nomic-embed-text"
	collectionName = "learning_demo"
)

func main() {
	fmt.Println("🧠 Agent Learning Demo - Watch the Agent Improve Over Time")
	fmt.Println("=============================================================")
	fmt.Println()

	// Setup: Create agent with VectorMemory for full learning
	ctx := context.Background()
	ag := setupAgent(ctx)

	// Demo scenario: Agent learns to handle calculation queries better
	fmt.Println("📚 Scenario: Agent learns which tools work best for calculations")
	fmt.Println()

	// Phase 1: Initial queries (exploration phase)
	fmt.Println("Phase 1: Initial Learning (First 5 queries)")
	fmt.Println("-------------------------------------------")
	queries := []string{
		"Calculate 123 * 456",
		"What is 789 + 321?",
		"Compute 555 - 222",
		"Calculate 999 / 3",
		"What is 88 * 12?",
	}

	initialStats := runQueries(ctx, ag, queries, "Initial")

	// Phase 2: More queries (agent starts exploiting learned knowledge)
	fmt.Println()
	fmt.Println("Phase 2: Learning in Action (Next 10 queries)")
	fmt.Println("----------------------------------------------")
	moreQueries := []string{
		"Calculate 234 * 567",
		"What is 890 + 123?",
		"Compute 666 - 333",
		"Calculate 1000 / 4",
		"What is 99 * 11?",
		"Calculate 345 * 678",
		"What is 456 + 789?",
		"Compute 777 - 444",
		"Calculate 2000 / 5",
		"What is 77 * 13?",
	}

	learnedStats := runQueries(ctx, ag, moreQueries, "Learned")

	// Phase 3: Final queries (high exploitation, agent is confident)
	fmt.Println()
	fmt.Println("Phase 3: Expert Mode (Final 5 queries)")
	fmt.Println("---------------------------------------")
	finalQueries := []string{
		"Calculate 111 * 222",
		"What is 333 + 444?",
		"Compute 888 - 555",
		"Calculate 3000 / 6",
		"What is 66 * 14?",
	}

	expertStats := runQueries(ctx, ag, finalQueries, "Expert")

	// Show improvement summary
	fmt.Println()
	fmt.Println("📊 Learning Progress Summary")
	fmt.Println("============================")
	showProgressSummary(initialStats, learnedStats, expertStats)

	// Show tool statistics
	fmt.Println()
	fmt.Println("🔧 Tool Performance Analysis")
	fmt.Println("============================")
	showToolStats(ctx, ag)

	fmt.Println()
	fmt.Println("✨ Demo complete! The agent learned to:")
	fmt.Println("  1. Identify calculation queries faster")
	fmt.Println("  2. Select the best tool (calculator) consistently")
	fmt.Println("  3. Reduce latency through experience")
	fmt.Println("  4. Balance exploration (10%) vs exploitation (90%)")
}

func setupAgent(ctx context.Context) *agent.Agent {
	fmt.Println("⚙️  Setting up agent with VectorMemory for learning...")

	// Create LLM provider
	llm := ollama.New(ollamaURL, modelName)

	// Create VectorMemory for semantic search
	vectorMem, err := memory.NewVectorMemory(ctx, memory.VectorMemoryConfig{
		QdrantURL:      qdrantURL,
		CollectionName: collectionName,
		Embedder:       memory.NewOllamaEmbedder(ollamaURL, embedModel),
		CacheSize:      100,
	})
	if err != nil {
		log.Fatalf("❌ Failed to create VectorMemory: %v", err)
		log.Fatalf("   Make sure Qdrant is running: docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant")
	}

	// Create agent with learning enabled
	ag := agent.New(llm,
		agent.WithMemory(vectorMem),
		// Learning auto-enabled by default!
	)

	status := ag.Status()
	fmt.Printf("✅ Agent ready: %s memory, Learning: %v\n", status.Memory.Type, status.Learning.Enabled)
	fmt.Println()

	return ag
}

// QueryStats tracks performance metrics for a set of queries
type QueryStats struct {
	TotalQueries   int
	SuccessCount   int
	FailureCount   int
	TotalLatency   time.Duration
	AverageLatency time.Duration
}

func runQueries(ctx context.Context, ag *agent.Agent, queries []string, phase string) QueryStats {
	stats := QueryStats{TotalQueries: len(queries)}

	for i, query := range queries {
		fmt.Printf("[%s %d/%d] Query: %s\n", phase, i+1, len(queries), query)

		start := time.Now()
		response, err := ag.Chat(ctx, query)
		latency := time.Since(start)

		stats.TotalLatency += latency

		if err != nil {
			stats.FailureCount++
			fmt.Printf("  ❌ Error: %v (latency: %v)\n", err, latency)
		} else {
			stats.SuccessCount++
			// Show shortened response
			shortResp := response
			if len(shortResp) > 100 {
				shortResp = shortResp[:100] + "..."
			}
			fmt.Printf("  ✅ Success: %s (latency: %v)\n", shortResp, latency)
		}

		// Small delay to avoid overwhelming the system
		time.Sleep(500 * time.Millisecond)
	}

	if stats.TotalQueries > 0 {
		stats.AverageLatency = stats.TotalLatency / time.Duration(stats.TotalQueries)
	}

	return stats
}

func showProgressSummary(initial, learned, expert QueryStats) {
	fmt.Printf("Phase 1 (Initial):  %d queries, %d success, avg latency: %v\n",
		initial.TotalQueries, initial.SuccessCount, initial.AverageLatency)
	fmt.Printf("Phase 2 (Learned):  %d queries, %d success, avg latency: %v\n",
		learned.TotalQueries, learned.SuccessCount, learned.AverageLatency)
	fmt.Printf("Phase 3 (Expert):   %d queries, %d success, avg latency: %v\n",
		expert.TotalQueries, expert.SuccessCount, expert.AverageLatency)

	// Calculate improvements
	if initial.AverageLatency > 0 {
		latencyImprovement := float64(initial.AverageLatency-expert.AverageLatency) / float64(initial.AverageLatency) * 100
		fmt.Printf("\n📈 Latency improvement: %.1f%% faster in Expert mode\n", latencyImprovement)
	}

	totalSuccess := initial.SuccessCount + learned.SuccessCount + expert.SuccessCount
	totalQueries := initial.TotalQueries + learned.TotalQueries + expert.TotalQueries
	successRate := float64(totalSuccess) / float64(totalQueries) * 100
	fmt.Printf("📈 Overall success rate: %.1f%% (%d/%d)\n", successRate, totalSuccess, totalQueries)
}

func showToolStats(ctx context.Context, ag *agent.Agent) {
	// Get tool statistics for calculation intent
	calcStats, err := ag.GetToolStats(ctx, "calculator", "calculation")
	if err == nil && calcStats != nil {
		fmt.Printf("Calculator (calculation queries):\n")
		fmt.Printf("  Success Rate: %.1f%%\n", calcStats.SuccessRate*100)
		fmt.Printf("  Average Latency: %dms\n", calcStats.AvgLatency)
		fmt.Printf("  Total Calls: %d\n", calcStats.TotalCalls)
		fmt.Printf("  Successes: %d, Failures: %d\n", calcStats.Successes, calcStats.Failures)
	}

	// Get recommendation for a calculation query
	recommendation, err := ag.GetToolRecommendation(ctx, "calculate 100 * 200", "calculation")
	if err == nil && recommendation != nil {
		fmt.Printf("\nRecommended tool for 'calculate 100 * 200':\n")
		fmt.Printf("  Tool: %s\n", recommendation.ToolName)
		fmt.Printf("  Confidence: %.1f%%\n", recommendation.Confidence*100)
		fmt.Printf("  Reasoning: %s\n", recommendation.Reasoning)
		fmt.Printf("  Strategy: %s\n", recommendation.DecisionStrategy)
		if recommendation.IsExploration {
			fmt.Printf("  Mode: Exploration (discovering new patterns)\n")
		} else {
			fmt.Printf("  Mode: Exploitation (using learned knowledge)\n")
		}
	}
}
