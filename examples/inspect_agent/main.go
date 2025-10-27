package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
	fmt.Println("🔍 Agent Inspection - What's Inside agent.New(llm)?")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Create agent with zero configuration
	llm := ollama.New("http://localhost:11434", "qwen3:1.7b")
	ag := agent.New(llm)

	fmt.Println("✅ Agent created with: agent.New(llm)")
	fmt.Println()

	// Get agent status
	status := ag.Status()

	// 1. Configuration
	fmt.Println("📋 1. CONFIGURATION")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("System Prompt:       %s\n", truncate(status.Configuration.SystemPrompt, 50))
	fmt.Printf("Temperature:         %.1f\n", status.Configuration.Temperature)
	fmt.Printf("Max Tokens:          %d\n", status.Configuration.MaxTokens)
	fmt.Printf("Max Iterations:      %d\n", status.Configuration.MaxIterations)
	fmt.Printf("Min Confidence:      %.1f%%\n", status.Configuration.MinConfidence*100)
	fmt.Printf("Reflection Enabled:  %v\n", status.Configuration.EnableReflection)
	fmt.Printf("Learning Enabled:    %v\n", status.Configuration.EnableLearning)
	fmt.Println()

	// 2. Reasoning Capabilities
	fmt.Println("🧠 2. REASONING CAPABILITIES")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Auto Reasoning:      %v\n", status.Reasoning.AutoReasoningEnabled)
	fmt.Printf("CoT Available:       %v (Chain-of-Thought - step-by-step reasoning)\n", status.Reasoning.CoTAvailable)
	fmt.Printf("ReAct Available:     %v (Reason + Act with tools)\n", status.Reasoning.ReActAvailable)
	fmt.Printf("Reflection:          %v (Self-review and validation)\n", status.Reasoning.ReflectionAvailable)
	fmt.Println()

	// 3. Memory System
	fmt.Println("💾 3. MEMORY SYSTEM")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Type:                %s\n", status.Memory.Type)
	fmt.Printf("Supports Search:     %v\n", status.Memory.SupportsSearch)
	fmt.Printf("Supports Vectors:    %v\n", status.Memory.SupportsVectors)
	fmt.Printf("Message Count:       %d\n", status.Memory.MessageCount)
	
	if status.Memory.Type == "vector" {
		fmt.Println("✓ VectorMemory: Semantic search enabled (Qdrant detected)")
	} else {
		fmt.Println("ℹ Buffer Memory: Simple storage (Qdrant not available)")
		fmt.Println("  💡 To enable VectorMemory: docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant")
	}
	fmt.Println()

	// 4. Learning System
	fmt.Println("📚 4. LEARNING SYSTEM")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Learning Enabled:    %v\n", status.Learning.Enabled)
	fmt.Printf("Experience Store:    %v\n", status.Learning.ExperienceStoreReady)
	fmt.Printf("Tool Selector:       %v\n", status.Learning.ToolSelectorReady)
	fmt.Printf("Conversation ID:     %s\n", status.Learning.ConversationID)
	
	if status.Learning.Enabled && status.Learning.ExperienceStoreReady {
		fmt.Println("✓ Full learning active: ε-greedy algorithm (90% exploit, 10% explore)")
	} else if status.Learning.Enabled {
		fmt.Println("⚡ Learning enabled but waiting for VectorMemory")
	} else {
		fmt.Println("✗ Learning disabled")
	}
	fmt.Println()

	// 5. Built-in Tools
	fmt.Println("🔧 5. BUILT-IN TOOLS")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Total Tools:         %d\n", status.Tools.TotalCount)
	fmt.Println("\nAvailable Tools:")
	
	// Group tools by category
	toolsByCategory := groupTools(status.Tools.ToolNames)
	for category, tools := range toolsByCategory {
		fmt.Printf("\n  %s (%d):\n", category, len(tools))
		for _, tool := range tools {
			fmt.Printf("    • %s\n", tool)
		}
	}
	fmt.Println()

	// 6. Test Learning Report (if available)
	if status.Learning.Enabled && status.Learning.ExperienceStoreReady {
		fmt.Println("🧠 6. AGENT SELF-ASSESSMENT")
		fmt.Println(strings.Repeat("-", 70))
		
		ctx := context.Background()
		report, err := ag.GetLearningReport(ctx)
		if err == nil && report != nil {
			fmt.Printf("Total Experiences:   %d\n", report.TotalExperiences)
			fmt.Printf("Learning Stage:      %s\n", report.LearningStage)
			fmt.Printf("Production Ready:    %v\n", report.ReadyForProduction)
			
			if len(report.Insights) > 0 {
				fmt.Println("\nAgent Insights:")
				for _, insight := range report.Insights {
					fmt.Printf("  • %s\n", insight)
				}
			}
			
			if len(report.Warnings) > 0 {
				fmt.Println("\nWarnings:")
				for _, warning := range report.Warnings {
					fmt.Printf("  ⚠ %s\n", warning)
				}
			}
		} else {
			fmt.Printf("Learning report not available yet (needs first interaction)\n")
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("📊 SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("✅ Reasoning Modes:  %d (CoT, ReAct, Reflection)\n", 3)
	fmt.Printf("✅ Built-in Tools:   %d tools ready to use\n", status.Tools.TotalCount)
	fmt.Printf("✅ Memory:           %s\n", status.Memory.Type)
	fmt.Printf("✅ Learning:         %v\n", status.Learning.Enabled)
	fmt.Printf("✅ Auto Reasoning:   %v\n", status.Reasoning.AutoReasoningEnabled)
	fmt.Println()
	fmt.Println("🎯 READY TO USE: Just call ag.Chat(ctx, \"your question\")")
	fmt.Println("   The agent will automatically:")
	fmt.Println("   - Select best reasoning mode (CoT/ReAct)")
	fmt.Println("   - Use appropriate tools")
	fmt.Println("   - Self-reflect on answers")
	fmt.Println("   - Learn from experience")
	fmt.Println(strings.Repeat("=", 70))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func groupTools(toolNames []string) map[string][]string {
	groups := make(map[string][]string)
	
	for _, name := range toolNames {
		category := "Other"
		
		if strings.HasPrefix(name, "file_") {
			category = "File Operations"
		} else if strings.HasPrefix(name, "web_") {
			category = "Web Tools"
		} else if strings.HasPrefix(name, "dns_") || strings.HasPrefix(name, "ping") || 
			strings.HasPrefix(name, "whois_") || strings.HasPrefix(name, "ssl_") || 
			strings.HasPrefix(name, "ip_") {
			category = "Network Tools"
		} else if strings.HasPrefix(name, "gmail_") {
			category = "Gmail Tools"
		} else if strings.HasPrefix(name, "datetime_") {
			category = "DateTime Tools"
		} else if strings.HasPrefix(name, "system_") {
			category = "System Tools"
		} else if name == "calculator" || name == "stats" {
			category = "Math Tools"
		} else if strings.HasPrefix(name, "mongodb_") {
			category = "MongoDB Tools"
		}
		
		groups[category] = append(groups[category], name)
	}
	
	return groups
}
