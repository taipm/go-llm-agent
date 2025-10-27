package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

const (
	qdrantURL      = "localhost:6334"
	ollamaURL      = "http://localhost:11434"
	modelName      = "qwen3:1.7b"
	embedModel     = "nomic-embed-text:latest"
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

	// Show phase summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Phase 1 Summary: %d/%d success, avg latency: %v\n",
		initialStats.SuccessCount, initialStats.TotalQueries, initialStats.AverageLatency)
	showDetailedToolStats(ctx, ag, "After Phase 1")
	fmt.Println(strings.Repeat("=", 60))

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

	// Show phase summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Phase 2 Summary: %d/%d success, avg latency: %v\n",
		learnedStats.SuccessCount, learnedStats.TotalQueries, learnedStats.AverageLatency)
	showDetailedToolStats(ctx, ag, "After Phase 2")
	fmt.Println(strings.Repeat("=", 60))

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

	// Show phase summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Phase 3 Summary: %d/%d success, avg latency: %v\n",
		expertStats.SuccessCount, expertStats.TotalQueries, expertStats.AverageLatency)
	showDetailedToolStats(ctx, ag, "After Phase 3 (Final)")
	fmt.Println(strings.Repeat("=", 60))

	// Show improvement summary
	fmt.Println()
	fmt.Println("📊 Learning Progress Summary")
	fmt.Println("============================")
	showProgressSummary(initialStats, learnedStats, expertStats)

	// Get agent's self-assessment
	fmt.Println()
	fmt.Println("🤖 Agent Self-Assessment")
	fmt.Println("============================")
	showAgentSelfAssessment(ctx, ag)

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
		fmt.Printf("\n[%s %d/%d] %s\n", phase, i+1, len(queries), query)

		start := time.Now()
		response, err := ag.Chat(ctx, query)
		latency := time.Since(start)

		stats.TotalLatency += latency

		if err != nil {
			stats.FailureCount++
			fmt.Printf("  ❌ Failed (latency: %v)\n", latency)
		} else {
			stats.SuccessCount++
			// Show shortened response
			shortResp := response
			if len(shortResp) > 80 {
				shortResp = shortResp[:80] + "..."
			}
			fmt.Printf("  ✅ %s (%v)\n", shortResp, latency)
		}

		// Small delay to avoid overwhelming the system
		time.Sleep(500 * time.Millisecond)
	}

	if stats.TotalQueries > 0 {
		stats.AverageLatency = stats.TotalLatency / time.Duration(stats.TotalQueries)
	}

	return stats
}

func showDetailedToolStats(ctx context.Context, ag *agent.Agent, title string) {
	fmt.Printf("\n📈 %s - Learning Metrics:\n", title)
	fmt.Println(strings.Repeat("-", 60))

	// Get calculator stats
	calcStats, err := ag.GetToolStats(ctx, "calculator", "calculation")
	if err == nil && calcStats != nil {
		fmt.Printf("Calculator Performance:\n")
		fmt.Printf("  Total Calls:     %d\n", calcStats.TotalCalls)
		fmt.Printf("  Successes:       %d\n", calcStats.Successes)
		fmt.Printf("  Failures:        %d\n", calcStats.Failures)
		fmt.Printf("  Success Rate:    %.1f%%\n", calcStats.SuccessRate*100)
		fmt.Printf("  Avg Latency:     %dms\n", calcStats.AvgLatency)

		// Show learning trend
		if calcStats.TotalCalls >= 3 {
			fmt.Printf("  Learning Status: ✅ Enough data for exploitation\n")
		} else {
			fmt.Printf("  Learning Status: 🔍 Still exploring (need %d more samples)\n", 3-calcStats.TotalCalls)
		}
	}

	// Get tool recommendation to show decision making
	recommendation, err := ag.GetToolRecommendation(ctx, "calculate 100 * 200", "calculation")
	if err == nil && recommendation != nil {
		fmt.Printf("\nDecision Strategy:\n")
		fmt.Printf("  Strategy:        %s\n", recommendation.DecisionStrategy)
		fmt.Printf("  Confidence:      %.1f%%\n", recommendation.Confidence*100)
		if recommendation.IsExploration {
			fmt.Printf("  Mode:            🔍 Exploration (discovering patterns)\n")
		} else {
			fmt.Printf("  Mode:            💡 Exploitation (using learned knowledge)\n")
		}
		fmt.Printf("  Sample Size:     %d experiences\n", recommendation.SampleSize)
	}

	fmt.Println(strings.Repeat("-", 60))
}

func showProgressSummary(initial, learned, expert QueryStats) {
	fmt.Println("\n📊 Overall Performance by Phase:")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Printf("%-15s | Queries | Success | Avg Latency | Success Rate\n", "Phase")
	fmt.Println(strings.Repeat("-", 70))

	initialRate := float64(initial.SuccessCount) / float64(initial.TotalQueries) * 100
	learnedRate := float64(learned.SuccessCount) / float64(learned.TotalQueries) * 100
	expertRate := float64(expert.SuccessCount) / float64(expert.TotalQueries) * 100

	fmt.Printf("%-15s | %7d | %7d | %11v | %10.1f%%\n",
		"Initial", initial.TotalQueries, initial.SuccessCount, initial.AverageLatency, initialRate)
	fmt.Printf("%-15s | %7d | %7d | %11v | %10.1f%%\n",
		"Learned", learned.TotalQueries, learned.SuccessCount, learned.AverageLatency, learnedRate)
	fmt.Printf("%-15s | %7d | %7d | %11v | %10.1f%%\n",
		"Expert", expert.TotalQueries, expert.SuccessCount, expert.AverageLatency, expertRate)

	fmt.Println(strings.Repeat("-", 70))

	// Calculate improvements
	totalSuccess := initial.SuccessCount + learned.SuccessCount + expert.SuccessCount
	totalQueries := initial.TotalQueries + learned.TotalQueries + expert.TotalQueries
	overallRate := float64(totalSuccess) / float64(totalQueries) * 100

	fmt.Printf("\n🎯 Overall Statistics:\n")
	fmt.Printf("   Total Queries:    %d\n", totalQueries)
	fmt.Printf("   Total Successes:  %d\n", totalSuccess)
	fmt.Printf("   Success Rate:     %.1f%%\n", overallRate)

	if initial.AverageLatency > 0 && expert.AverageLatency > 0 {
		latencyImprovement := float64(initial.AverageLatency-expert.AverageLatency) / float64(initial.AverageLatency) * 100
		fmt.Printf("\n📈 Learning Improvements:\n")
		fmt.Printf("   Latency Reduction:  %.1f%% faster (Initial: %v → Expert: %v)\n",
			latencyImprovement, initial.AverageLatency, expert.AverageLatency)

		if latencyImprovement > 0 {
			fmt.Printf("   ✅ Agent learned to respond %.1f%% faster!\n", latencyImprovement)
		}
	}

	// Success rate trend
	if expertRate > initialRate {
		improvement := expertRate - initialRate
		fmt.Printf("   Success Rate Gain:  +%.1f%% (Initial: %.1f%% → Expert: %.1f%%)\n",
			improvement, initialRate, expertRate)
		fmt.Printf("   ✅ Agent improved success rate by %.1f%%!\n", improvement)
	} else if expertRate == initialRate && expertRate >= 95.0 {
		fmt.Printf("   ✅ Maintained high success rate: %.1f%%\n", expertRate)
	}
}

func showAgentSelfAssessment(ctx context.Context, ag *agent.Agent) {
	report, err := ag.GetLearningReport(ctx)
	if err != nil {
		fmt.Printf("Unable to get learning report: %v\n", err)
		return
	}

	fmt.Printf("Learning Stage:      %s\n", report.LearningStage)
	fmt.Printf("Total Experiences:   %d\n", report.TotalExperiences)
	fmt.Printf("Production Ready:    %v\n", report.ReadyForProduction)

	if len(report.Insights) > 0 {
		fmt.Println("\n💡 Agent's Insights:")
		for i, insight := range report.Insights {
			fmt.Printf("  %d. %s\n", i+1, insight)
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Println("\n⚠️  Agent's Warnings:")
		for i, warning := range report.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
	}

	if len(report.ToolPerformance) > 0 {
		fmt.Println("\n📊 Tool Performance Summary:")
		for toolIntent, stats := range report.ToolPerformance {
			fmt.Printf("  %s: %d calls, %.1f%% success, avg %dms\n",
				toolIntent, stats.TotalCalls, stats.SuccessRate*100, stats.AvgLatency)
		}
	}
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
