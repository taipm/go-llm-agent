// Simple example showing how easy it is to use built-in tools with the builtin package
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/pkg/builtin"
)

func main() {
	fmt.Println("=== Built-in Tools - Simple Example ===\n")

	// That's it! One line to get all 10 built-in tools with sensible defaults
	registry := builtin.GetRegistry()

	fmt.Printf("✅ Registered %d tools automatically:\n\n", registry.Count())

	// Show all registered tools
	for _, name := range registry.Names() {
		tool := registry.Get(name)
		safety := "✅ safe"
		if !tool.IsSafe() {
			safety = "⚠️  unsafe"
		}
		fmt.Printf("  [%s] %-18s - %s\n", safety, name, tool.Description())
	}

	fmt.Println("\n--- Example 1: Get current time ---")
	ctx := context.Background()
	result, err := registry.Execute(ctx, "datetime_now", map[string]interface{}{
		"format":   "RFC3339",
		"timezone": "Asia/Ho_Chi_Minh",
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Current time in Vietnam: %v\n", result)

	fmt.Println("\n--- Example 2: Format conversion ---")
	result, err = registry.Execute(ctx, "datetime_format", map[string]interface{}{
		"datetime":      "2024-10-27T15:30:00Z",
		"from_format":   "RFC3339",
		"to_format":     "RFC1123",
		"from_timezone": "UTC",
		"to_timezone":   "Asia/Tokyo",
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Formatted time in Tokyo: %v\n", result)

	fmt.Println("\n--- Example 3: List current directory ---")
	result, err = registry.Execute(ctx, "file_list", map[string]interface{}{
		"path":      ".",
		"recursive": false,
		"pattern":   "*.go",
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Go files: %v\n", result)

	fmt.Println("\n🎉 That's how simple it is to use built-in tools!")
	fmt.Println("💡 See examples/builtin_tools/main.go for advanced usage with LLM providers")
}
