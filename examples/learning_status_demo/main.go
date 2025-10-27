package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
	fmt.Println("🎓 Agent Learning Status Demo")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Create agent
	llm := ollama.New("http://localhost:11434", "qwen3:1.7b")
	ag := agent.New(llm)
	ctx := context.Background()

	// Helper to show status
	showStatus := func(title string) {
		fmt.Println(strings.Repeat("=", 70))
		fmt.Printf("📊 %s\n", title)
		fmt.Println(strings.Repeat("=", 70))

		status := ag.Status()

		fmt.Printf("🧠 Learning Stage:      %s\n", status.Learning.LearningStage)
		fmt.Printf("📚 Total Experiences:   %d\n", status.Learning.TotalExperiences)
		fmt.Printf("✅ Success Rate:        %.1f%%\n", status.Learning.OverallSuccessRate)
		fmt.Printf("🚀 Production Ready:    %v\n", status.Learning.ReadyForProduction)

		if len(status.Learning.TopPerformingTools) > 0 {
			fmt.Println("\n⭐ Top Performing Tools:")
			for _, tool := range status.Learning.TopPerformingTools {
				fmt.Printf("   %s\n", tool)
			}
		}

		if len(status.Learning.ProblematicTools) > 0 {
			fmt.Println("\n⚠️  Problematic Tools:")
			for _, tool := range status.Learning.ProblematicTools {
				fmt.Printf("   %s\n", tool)
			}
		}

		if len(status.Learning.KnowledgeAreas) > 0 {
			fmt.Println("\n📖 Knowledge Areas:")
			for area, count := range status.Learning.KnowledgeAreas {
				fmt.Printf("   %s: %d experiences\n", area, count)
			}
		}

		if len(status.Learning.RecentImprovements) > 0 {
			fmt.Println("\n📈 Recent Improvements:")
			for _, improvement := range status.Learning.RecentImprovements {
				fmt.Printf("   %s\n", improvement)
			}
		}

		fmt.Println()
	}

	// Initial status
	showStatus("AGENT STATUS - Before Learning")

	// Run some tasks to let agent learn
	tasks := []struct {
		description string
		query       string
	}{
		{"Math calculation", "Calculate 123 * 456"},
		{"File operation", "Create a file called demo.txt with content 'Learning demo'"},
		{"Another calculation", "What is 789 + 321?"},
		{"File read", "Read the file demo.txt"},
		{"Complex math", "Calculate (100 + 50) * 3 - 25"},
		{"System info", "What are my system specs?"},
		{"Date/time", "What's the current date and time?"},
		{"Another math", "What is 999 / 3?"},
		{"File list", "List all files in current directory"},
		{"Final calculation", "Calculate 88 * 12"},
	}

	fmt.Println("🏃 Running tasks to train the agent...")
	fmt.Println(strings.Repeat("-", 70))

	for i, task := range tasks {
		fmt.Printf("\n[%d/%d] %s: %s\n", i+1, len(tasks), task.description, task.query)
		_, err := ag.Chat(ctx, task.query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("✅ Completed\n")
		}

		// Show status after every 3 tasks
		if (i+1)%3 == 0 {
			showStatus(fmt.Sprintf("AGENT STATUS - After %d Tasks", i+1))
		}
	}

	// Final status
	showStatus("AGENT STATUS - Final (After All Tasks)")

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("✨ Demo Complete!")
	fmt.Println()
	fmt.Println("Key Observations:")
	fmt.Println("• Learning stage progresses: exploring → learning → expert")
	fmt.Println("• Success rate improves over time")
	fmt.Println("• Agent identifies top performing and problematic tools")
	fmt.Println("• Knowledge areas show what agent has learned")
	fmt.Println("• Recent improvements track actual progress")
	fmt.Println(strings.Repeat("=", 70))
}
