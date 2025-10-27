package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
	fmt.Println("🚀 Zero-Config Agent Demo")
	fmt.Println("==================================================")
	fmt.Println()

	// Just this! Learning auto-enabled, VectorMemory auto-configured
	llm := ollama.New("http://localhost:11434", "qwen2.5:7b")
	ag := agent.New(llm)

	fmt.Println("✅ Agent created with just: agent.New(llm)")
	fmt.Println()
	fmt.Println("Behind the scenes:")
	fmt.Println("  - Learning enabled by default ✅")
	fmt.Println("  - Auto-tries VectorMemory (requires Qdrant) ✅")
	fmt.Println("  - Gracefully falls back to BufferMemory ✅")
	fmt.Println()

	// Example 1: Simple calculation
	fmt.Println("📝 Example 1: Calculate 123 * 456")
	ctx := context.Background()
	
	response, err := ag.Chat(ctx, "Calculate 123 * 456")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Printf("Response: %s\n", response)
	fmt.Println()

	// Example 2: File operations
	fmt.Println("📝 Example 2: Create a test file")
	response, err = ag.Chat(ctx, "Create a file called test.txt with content 'Hello, World!'")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Printf("Response: %s\n", response)
	fmt.Println()

	// Example 3: Another calculation (agent learns from first one)
	fmt.Println("📝 Example 3: Another calculation (agent may learn from first)")
	response, err = ag.Chat(ctx, "What is 789 * 321?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Printf("Response: %s\n", response)
	fmt.Println()

	// Get agent status
	status := ag.Status()
	fmt.Println("📊 Agent Status:")
	fmt.Printf("  Memory Type: %s\n", status.Memory.Type)
	fmt.Printf("  Total Messages: %d\n", status.Memory.MessageCount)
	fmt.Printf("  Learning Enabled: %v\n", status.Learning.Enabled)
	fmt.Printf("  Experience Store: %v\n", status.Learning.ExperienceStoreReady)
	fmt.Printf("  Tool Selector: %v\n", status.Learning.ToolSelectorReady)
	fmt.Println()

	fmt.Println("✨ That's it! Simple as agent.New(llm)")
	fmt.Println()
	fmt.Println("💡 To enable full learning with semantic search:")
	fmt.Println("   docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant")
	fmt.Println("   Then restart this program - it will auto-detect!")
}
