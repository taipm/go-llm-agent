package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/builtin"
	"github.com/taipm/go-llm-agent/pkg/provider"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║          GO-LLM-AGENT - Standalone Demo v0.1.1                ║")
	fmt.Println("║  Copy this example to any project to test the library         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	ctx := context.Background()

	// Auto-detect LLM provider from environment
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("❌ Failed to initialize LLM provider: %v\n", err)
	}

	// Display provider info
	fmt.Printf("✓ LLM Provider: %s\n", llm)
	fmt.Printf("✓ Model: %s\n", getModelName())
	fmt.Println()

	// Get all built-in tools from registry
	registry := builtin.GetRegistry()
	tools := registry.All()
	fmt.Printf("✓ Loaded %d built-in tools\n\n", len(tools))

	// Show tools by category
	categories := map[string][]string{
		"DateTime": {},
		"Math":     {},
		"File":     {},
		"Web":      {},
		"System":   {},
		"Database": {},
	}

	for _, tool := range tools {
		category := tool.Category()
		name := tool.Name()
		switch category {
		case "datetime":
			categories["DateTime"] = append(categories["DateTime"], name)
		case "math":
			categories["Math"] = append(categories["Math"], name)
		case "file":
			categories["File"] = append(categories["File"], name)
		case "web":
			categories["Web"] = append(categories["Web"], name)
		case "system":
			categories["System"] = append(categories["System"], name)
		case "database":
			categories["Database"] = append(categories["Database"], name)
		}
	}

	for category, items := range categories {
		if len(items) > 0 {
			fmt.Printf("✓ %s tools (%d): %s\n", category, len(items), strings.Join(items, ", "))
		}
	}
	fmt.Println()

	// Create agent (memory and logging initialized automatically)
	var a *agent.Agent
	a = agent.New(llm)

	// Register all tools
	for _, tool := range tools {
		a.AddTool(tool)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    INTERACTIVE CHAT MODE                       ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Println("║ The agent has access to 20 tools across 6 categories:         ║")
	fmt.Println("║  • DateTime - Get current time, calculate dates, format        ║")
	fmt.Println("║  • Math     - Evaluate expressions, calculate statistics       ║")
	fmt.Println("║  • File     - Read, write, list files                          ║")
	fmt.Println("║  • Web      - Fetch web pages, search                          ║")
	fmt.Println("║  • System   - Execute commands, get env vars, system info      ║")
	fmt.Println("║  • MongoDB  - Database operations (connect, CRUD, aggregate)   ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ Logging symbols:                                               ║")
	fmt.Println("║  👤 = User message    🤔 = Agent thinking    🔧 = Tool call    ║")
	fmt.Println("║  ✓ = Success          💬 = Response         💾 = Memory       ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ Try asking:                                                    ║")
	fmt.Println("║  • What time is it now?                                        ║")
	fmt.Println("║  • Calculate 15% of 350                                        ║")
	fmt.Println("║  • What's the mean of [10, 20, 30, 40, 50]?                    ║")
	fmt.Println("║  • How many days until Christmas 2025?                         ║")
	fmt.Println("║  • List files in current directory                             ║")
	fmt.Println("║  • What's my operating system?                                 ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ Vietnamese examples:                                           ║")
	fmt.Println("║  • Tôi sinh ngày 15/03/1990, năm nay tôi bao nhiêu tuổi?       ║")
	fmt.Println("║  • Tính 25 * 4 + 100                                           ║")
	fmt.Println("║  • Còn bao nhiêu ngày nữa đến Tết Nguyên Đán 2026?            ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ Commands:                                                      ║")
	fmt.Println("║  • Type 'quit' or 'exit' to stop                               ║")
	fmt.Println("║  • Type 'clear' to clear conversation history                  ║")
	fmt.Println("║  • Type 'help' to see this menu again                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Interactive chat loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("👤 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle commands
		switch strings.ToLower(input) {
		case "quit", "exit":
			fmt.Println("\n👋 Goodbye! Thanks for testing go-llm-agent!")
			return

		case "clear":
			a = agent.New(llm)
			// Re-register all tools
			for _, tool := range tools {
				a.AddTool(tool)
			}
			fmt.Println("✓ Conversation history cleared\n")
			continue

		case "help":
			showHelp()
			continue
		}

		// Chat with agent
		response, err := a.Chat(ctx, input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\n🤖 Agent: %s\n\n", response)
		fmt.Println("─────────────────────────────────────────────────────────────────")
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("❌ Scanner error: %v\n", err)
	}
}

func getProviderName() string {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "openai" // default
	}
	return strings.ToUpper(provider)
}

func getModelName() string {
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		// Default models based on provider
		switch strings.ToLower(os.Getenv("LLM_PROVIDER")) {
		case "ollama":
			model = "llama2"
		case "gemini":
			model = "gemini-pro"
		default:
			model = "gpt-3.5-turbo"
		}
	}
	return model
}

func showHelp() {
	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         HELP MENU                              ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Println("║ Available Tools (20 total):                                    ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ DateTime Tools:                                                ║")
	fmt.Println("║  • datetime_now      - Get current date/time                   ║")
	fmt.Println("║  • datetime_calc     - Calculate date differences              ║")
	fmt.Println("║  • datetime_format   - Format dates                            ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ Math Tools:                                                    ║")
	fmt.Println("║  • math_calculate    - Evaluate expressions (15% of 350)       ║")
	fmt.Println("║  • math_stats        - Statistics (mean, median, std dev)      ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ File Tools:                                                    ║")
	fmt.Println("║  • file_read         - Read file contents                      ║")
	fmt.Println("║  • file_write        - Write to file                           ║")
	fmt.Println("║  • file_list         - List directory contents                 ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ Web Tools:                                                     ║")
	fmt.Println("║  • web_fetch         - Fetch web page content                  ║")
	fmt.Println("║  • web_search        - Search the web                          ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ System Tools:                                                  ║")
	fmt.Println("║  • system_exec       - Execute shell commands                  ║")
	fmt.Println("║  • system_env        - Get environment variables               ║")
	fmt.Println("║  • system_info       - Get system information                  ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ MongoDB Tools:                                                 ║")
	fmt.Println("║  • mongodb_connect   - Connect to database                     ║")
	fmt.Println("║  • mongodb_insert    - Insert documents                        ║")
	fmt.Println("║  • mongodb_find      - Find documents                          ║")
	fmt.Println("║  • mongodb_update    - Update documents                        ║")
	fmt.Println("║  • mongodb_delete    - Delete documents                        ║")
	fmt.Println("║  • mongodb_aggregate - Run aggregation pipelines               ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║ Commands:                                                      ║")
	fmt.Println("║  • quit/exit  - Exit the program                               ║")
	fmt.Println("║  • clear      - Clear conversation history                     ║")
	fmt.Println("║  • help       - Show this menu                                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}
