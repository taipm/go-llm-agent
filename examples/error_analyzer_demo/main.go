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
	embedModel     = "nomic-embed-text"
	collectionName = "error_analyzer_demo"
)

func main() {
	fmt.Println("🔍 Error Pattern Detection Demo")
	fmt.Println("=================================")
	fmt.Println()

	// Setup agent with VectorMemory for learning
	ctx := context.Background()
	ag := setupAgent(ctx)

	// Phase 1: Generate errors to create patterns
	fmt.Println("📋 Phase 1: Generating Intentional Errors")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()

	// Scenario 1: Wrong tool usage (5 similar errors)
	fmt.Println("Scenario 1: Using wrong tool for calculations (5x)")
	errorQueries1 := []string{
		"Search the web for 123 * 456",  // Force web search instead of calculator
		"Find online the result of 789 + 321",
		"Look up 555 - 222 on the internet",
		"Google what is 999 / 3",
		"Search for 88 * 12",
	}
	stats1 := runQueries(ctx, ag, errorQueries1, "Wrong Tool")

	time.Sleep(1 * time.Second)

	// Scenario 2: Invalid operations (4 similar errors)
	fmt.Println("\nScenario 2: Invalid mathematical operations (4x)")
	errorQueries2 := []string{
		"Calculate 10 divided by zero",
		"Compute 5 divided by 0",
		"What is 100 / 0",
		"Find 50 divided by zero",
	}
	stats2 := runQueries(ctx, ag, errorQueries2, "Invalid Op")

	time.Sleep(1 * time.Second)

	// Scenario 3: Malformed queries (3 similar errors)
	fmt.Println("\nScenario 3: Ambiguous calculation requests (3x)")
	errorQueries3 := []string{
		"Calculate something with numbers",
		"Do some math for me",
		"Compute a result",
	}
	stats3 := runQueries(ctx, ag, errorQueries3, "Ambiguous")

	// Phase 2: Show agent status after errors
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Phase 2: Agent Status After Errors")
	fmt.Println(strings.Repeat("=", 60))
	showAgentStatus(ag)

	// Phase 3: Get learning report
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎓 Phase 3: Learning Report")
	fmt.Println(strings.Repeat("=", 60))
	showLearningReport(ctx, ag)

	// Phase 4: Test with correct queries to see if agent learned
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 Phase 4: Testing Learned Behavior")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("Now testing with CORRECT calculation queries:")
	correctQueries := []string{
		"Calculate 45 * 67",
		"What is 234 + 567",
		"Compute 890 - 321",
	}
	stats4 := runQueries(ctx, ag, correctQueries, "Correct")

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Summary")
	fmt.Println(strings.Repeat("=", 60))
	showSummary(stats1, stats2, stats3, stats4)

	fmt.Println("\n✅ Error pattern detection demo complete!")
	fmt.Println("\n💡 The agent has recorded all errors in its experience store.")
	fmt.Println("   Error patterns can be detected by clustering similar failures.")
	fmt.Println("   Future queries will benefit from learned error patterns.")
}

func setupAgent(ctx context.Context) *agent.Agent {
	fmt.Println("⚙️  Setting up agent with VectorMemory for error learning...")

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

	// Create agent with learning enabled (default)
	ag := agent.New(llm,
		agent.WithMemory(vectorMem),
	)

	status := ag.Status()
	fmt.Printf("✅ Agent ready:\n")
	fmt.Printf("   Memory: %s\n", status.Memory.Type)
	fmt.Printf("   Learning: %v\n", status.Learning.Enabled)
	fmt.Printf("   Experience Store: %v\n", status.Learning.ExperienceStoreReady)
	fmt.Printf("   Error Analyzer: %v\n", status.Learning.ErrorAnalyzerReady)
	fmt.Println()

	return ag
}

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
		fmt.Printf("  [%s %d/%d] %s\n", phase, i+1, len(queries), query)

		start := time.Now()
		_, err := ag.Chat(ctx, query)
		latency := time.Since(start)

		stats.TotalLatency += latency

		if err != nil {
			stats.FailureCount++
			fmt.Printf("    ❌ Error (expected, %v)\n", latency)
		} else {
			stats.SuccessCount++
			fmt.Printf("    ✅ Success (%v)\n", latency)
		}

		// Small delay
		time.Sleep(300 * time.Millisecond)
	}

	if stats.TotalQueries > 0 {
		stats.AverageLatency = stats.TotalLatency / time.Duration(stats.TotalQueries)
	}

	return stats
}

func showAgentStatus(ag *agent.Agent) {
	status := ag.Status()

	fmt.Println("\n🎓 Learning System Status:")
	fmt.Printf("   Experience Store Ready: %v\n", status.Learning.ExperienceStoreReady)
	fmt.Printf("   Tool Selector Ready:    %v\n", status.Learning.ToolSelectorReady)
	fmt.Printf("   Error Analyzer Ready:   %v\n", status.Learning.ErrorAnalyzerReady)
	fmt.Printf("   Total Experiences:      %d\n", status.Learning.TotalExperiences)
	fmt.Printf("   Learning Stage:         %s\n", status.Learning.LearningStage)
	fmt.Printf("   Success Rate:           %.1f%%\n", status.Learning.OverallSuccessRate*100)
	fmt.Printf("   Production Ready:       %v\n", status.Learning.ReadyForProduction)

	if len(status.Learning.TopPerformingTools) > 0 {
		fmt.Println("\n✅ Top Performing Tools:")
		for _, tool := range status.Learning.TopPerformingTools {
			fmt.Printf("   • %s\n", tool)
		}
	}

	if len(status.Learning.ProblematicTools) > 0 {
		fmt.Println("\n❌ Problematic Tools:")
		for _, tool := range status.Learning.ProblematicTools {
			fmt.Printf("   • %s\n", tool)
		}
	}
}

func showLearningReport(ctx context.Context, ag *agent.Agent) {
	report, err := ag.GetLearningReport(ctx)
	if err != nil {
		fmt.Printf("Unable to get learning report: %v\n", err)
		return
	}

	fmt.Printf("\nTotal Experiences:   %d\n", report.TotalExperiences)
	fmt.Printf("Learning Stage:      %s\n", report.LearningStage)
	fmt.Printf("Production Ready:    %v\n", report.ReadyForProduction)

	if len(report.Insights) > 0 {
		fmt.Println("\n💡 Insights:")
		for i, insight := range report.Insights {
			fmt.Printf("  %d. %s\n", i+1, insight)
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Println("\n⚠️  Warnings:")
		for i, warning := range report.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
	}

	if len(report.ToolPerformance) > 0 {
		fmt.Println("\n📊 Tool Performance:")
		for toolIntent, stats := range report.ToolPerformance {
			fmt.Printf("  %s: %d calls, %.1f%% success\n",
				toolIntent, stats.TotalCalls, stats.SuccessRate*100)
		}
	}
}

func showSummary(stats1, stats2, stats3, stats4 QueryStats) {
	fmt.Println("\n📊 Overall Statistics:")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Printf("%-25s | Queries | Success | Failure\n", "Scenario")
	fmt.Println(strings.Repeat("-", 60))

	printStats := func(name string, stats QueryStats) {
		fmt.Printf("%-25s | %7d | %7d | %7d\n",
			name, stats.TotalQueries, stats.SuccessCount, stats.FailureCount)
	}

	printStats("Wrong Tool (expected fail)", stats1)
	printStats("Invalid Op (expected fail)", stats2)
	printStats("Ambiguous (expected fail)", stats3)
	printStats("Correct Queries", stats4)

	fmt.Println(strings.Repeat("-", 60))

	totalQueries := stats1.TotalQueries + stats2.TotalQueries + stats3.TotalQueries + stats4.TotalQueries
	totalSuccess := stats1.SuccessCount + stats2.SuccessCount + stats3.SuccessCount + stats4.SuccessCount
	totalFailure := stats1.FailureCount + stats2.FailureCount + stats3.FailureCount + stats4.FailureCount

	fmt.Printf("%-25s | %7d | %7d | %7d\n", "TOTAL", totalQueries, totalSuccess, totalFailure)

	successRate := float64(totalSuccess) / float64(totalQueries) * 100
	fmt.Printf("\nOverall Success Rate: %.1f%%\n", successRate)
	fmt.Printf("\n💡 Note: Error scenarios were intentional to generate patterns for learning.\n")
	fmt.Printf("   The agent recorded %d failures as learning experiences.\n", totalFailure)
	fmt.Printf("   These patterns can now be used to prevent similar errors.\n")
}

