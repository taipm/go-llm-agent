package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/agent"
	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// PersonalAssistant represents an intelligent personal assistant
type PersonalAssistant struct {
	agent  *agent.Agent
	ctx    context.Context
	logger logger.Logger
}

// NewPersonalAssistant creates a new personal assistant with LLM and tools
func NewPersonalAssistant() (*PersonalAssistant, error) {
	// Load .env file (ignore error if not exists)
	godotenv.Load()

	ctx := context.Background()

	// Read configuration from environment
	llmProvider := getEnv("LLM_PROVIDER", "ollama")
	llmModel := getEnv("LLM_MODEL", "qwen2.5:3b")
	ollamaURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
	openaiURL := getEnv("OPENAI_BASE_URL", "https://api.openai.com")

	useVectorMemory := getEnvBool("USE_VECTOR_MEMORY", true)
	qdrantURL := getEnv("QDRANT_URL", "localhost:6334")
	collectionName := getEnv("QDRANT_COLLECTION", "personal_assistant")
	embeddingModel := getEnv("EMBEDDING_MODEL", "nomic-embed-text:latest")
	embeddingURL := getEnv("EMBEDDING_BASE_URL", "http://localhost:11434")
	cacheSize := getEnvInt("MEMORY_CACHE_SIZE", 100)

	enableLearning := getEnvBool("ENABLE_LEARNING", true)
	enableReflection := getEnvBool("ENABLE_REFLECTION", true)
	minConfidence := getEnvFloat("MIN_CONFIDENCE", 0.7)
	temperature := getEnvFloat("TEMPERATURE", 0.7)
	maxTokens := getEnvInt("MAX_TOKENS", 2000)
	maxIterations := getEnvInt("MAX_ITERATIONS", 10)
	logLevel := getEnv("LOG_LEVEL", "INFO")
	systemPrompt := getEnv("SYSTEM_PROMPT", "")

	// Initialize LLM provider config
	providerConfig := provider.Config{
		Type:  provider.ProviderType(llmProvider),
		Model: llmModel,
	}

	// Set BaseURL only for non-standard endpoints
	// For OpenAI official API, DON'T set BaseURL (SDK adds /v1 automatically)
	// For Ollama, set custom URL
	// For Gemini, no BaseURL needed
	if llmProvider == "ollama" {
		providerConfig.BaseURL = ollamaURL
	} else if llmProvider == "openai" && openaiURL != "" && openaiURL != "https://api.openai.com" {
		// Only set BaseURL for Azure OpenAI or custom endpoints
		providerConfig.BaseURL = openaiURL
	}

	// Add API keys based on provider
	if llmProvider == "openai" {
		providerConfig.APIKey = os.Getenv("OPENAI_API_KEY")
	} else if llmProvider == "gemini" {
		providerConfig.APIKey = os.Getenv("GEMINI_API_KEY")
	}

	llm, err := provider.New(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// Create memory based on configuration
	var mem types.Memory
	if useVectorMemory {
		vectorMem, err := memory.NewVectorMemory(ctx, memory.VectorMemoryConfig{
			QdrantURL:      qdrantURL,
			CollectionName: collectionName,
			Embedder:       memory.NewOllamaEmbedder(embeddingURL, embeddingModel),
			CacheSize:      cacheSize,
		})

		if err != nil {
			fmt.Printf("[WARN] Failed to create VectorMemory: %v. Using BufferMemory instead.\n", err)
			mem = memory.NewBuffer(cacheSize)
		} else {
			mem = vectorMem
		}
	} else {
		mem = memory.NewBuffer(cacheSize)
	}

	// Parse log level
	var logLvl logger.LogLevel
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		logLvl = logger.LogLevelDebug
	case "INFO":
		logLvl = logger.LogLevelInfo
	case "WARN":
		logLvl = logger.LogLevelWarn
	case "ERROR":
		logLvl = logger.LogLevelError
	default:
		logLvl = logger.LogLevelInfo
	}

	// Build agent options
	opts := []agent.Option{
		agent.WithMemory(mem),
		agent.WithLearning(enableLearning),
		agent.WithReflection(enableReflection),
		agent.WithMinConfidence(minConfidence),
		agent.WithTemperature(temperature),
		agent.WithMaxTokens(maxTokens),
		agent.WithLogLevel(logLvl),
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		opts = append(opts, agent.WithSystemPrompt(systemPrompt))
	}

	// Create agent with configured options
	ag := agent.New(llm, opts...)

	// Note: maxIterations is handled internally by agent
	_ = maxIterations

	return &PersonalAssistant{
		agent:  ag,
		ctx:    ctx,
		logger: logger.NewConsoleLogger(),
	}, nil
}

// Ask sends a request to the assistant
func (pa *PersonalAssistant) Ask(question string) (string, error) {
	response, err := pa.agent.Chat(pa.ctx, question)
	if err != nil {
		return "", err
	}
	return response, nil
}

// ShowStatus displays agent's current status and capabilities
func (pa *PersonalAssistant) ShowStatus() {
	status := pa.agent.Status()

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📊 PERSONAL ASSISTANT STATUS")
	fmt.Println(strings.Repeat("=", 70))

	// Basic info
	fmt.Printf("🤖 Provider: %s\n", status.Provider)
	fmt.Printf("💬 Conversations: %d messages in memory\n", status.Memory.MessageCount)
	fmt.Printf("🔧 Available Tools: %d\n", status.Tools.TotalCount)

	// Learning status
	if status.Learning.Enabled {
		fmt.Println("\n🧠 LEARNING SYSTEM:")
		fmt.Printf("   ✅ Enabled: %v\n", status.Learning.Enabled)
		if status.Learning.ExperienceStoreReady {
			fmt.Printf("   📚 Total Experiences: %d\n", status.Learning.TotalExperiences)
			fmt.Printf("   ✨ Success Rate: %.1f%%\n", status.Learning.OverallSuccessRate*100)
		}
		if status.Learning.ToolSelectorReady {
			fmt.Println("   🎯 Tool Selector: Active")
		}
		if status.Learning.ErrorAnalyzerReady {
			fmt.Println("   🔍 Error Analyzer: Active")
		}
	} else {
		fmt.Println("\n🧠 LEARNING: Disabled (requires VectorMemory)")
	}

	// Tool categories
	fmt.Println("\n🔧 AVAILABLE TOOLS:")
	for i, toolName := range status.Tools.ToolNames {
		if i < 10 { // Show first 10
			fmt.Printf("   • %s\n", toolName)
		}
	}
	if status.Tools.TotalCount > 10 {
		fmt.Printf("   ... and %d more\n", status.Tools.TotalCount-10)
	}

	fmt.Println(strings.Repeat("=", 70) + "\n")
}

// RunDemo runs interactive demo scenarios
func (pa *PersonalAssistant) RunDemo() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 PERSONAL ASSISTANT DEMO - Practical Use Cases")
	fmt.Println(strings.Repeat("=", 70))

	scenarios := []struct {
		name        string
		description string
		query       string
	}{
		{
			name:        "1. Daily Information",
			description: "Get current date, time, and day of week",
			query:       "What's today's date and what day of the week is it?",
		},
		{
			name:        "2. Web Research",
			description: "Search and fetch information from the web",
			query:       "Search for the latest news about artificial intelligence and summarize the top result",
		},
		{
			name:        "3. Calculations",
			description: "Perform mathematical calculations",
			query:       "Calculate the compound interest for $10,000 at 5% annual rate over 3 years",
		},
		{
			name:        "4. File Management",
			description: "List and read files in current directory",
			query:       "Show me what files are in the current directory",
		},
		{
			name:        "5. System Information",
			description: "Check system resources and environment",
			query:       "What's the current system CPU and memory usage?",
		},
		{
			name:        "6. Network Tools",
			description: "DNS lookup and network diagnostics",
			query:       "Do a DNS lookup for google.com and tell me its IP addresses",
		},
		{
			name:        "7. Time Calculations",
			description: "Calculate time differences and deadlines",
			query:       "How many days until Christmas 2025?",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n%s\n", scenario.name)
		fmt.Printf("📋 %s\n", scenario.description)
		fmt.Printf("❓ Query: %s\n", scenario.query)
		fmt.Println(strings.Repeat("-", 70))

		response, err := pa.Ask(scenario.query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("💡 Answer: %s\n", response)
		}

		// Pause between scenarios (except last one)
		if i < len(scenarios)-1 {
			fmt.Print("\n⏸  Press Enter to continue to next scenario...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}
}

// RunInteractive runs interactive chat mode
func (pa *PersonalAssistant) RunInteractive() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("💬 INTERACTIVE MODE - Chat with your Personal Assistant")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\nTips:")
	fmt.Println("  • Ask me to search the web, do calculations, check files, etc.")
	fmt.Println("  • Type 'status' to see my current status and capabilities")
	fmt.Println("  • Type 'help' for example questions")
	fmt.Println("  • Type 'exit' to quit\n")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)

		// Handle special commands
		switch strings.ToLower(input) {
		case "exit", "quit", "q":
			fmt.Println("\n👋 Goodbye! Thanks for using Personal Assistant!")
			return
		case "status":
			pa.ShowStatus()
			continue
		case "help":
			pa.showHelp()
			continue
		case "":
			continue
		}

		// Process query
		fmt.Println()
		response, err := pa.Ask(input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
		} else {
			fmt.Printf("🤖 Assistant: %s\n\n", response)
		}
	}
}

func (pa *PersonalAssistant) showHelp() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📖 EXAMPLE QUESTIONS")
	fmt.Println(strings.Repeat("=", 70))

	examples := []string{
		"What's the current time and date?",
		"Search the web for Python tutorials",
		"Calculate 15% of 250",
		"What files are in the current directory?",
		"Check system CPU usage",
		"Do a DNS lookup for github.com",
		"How many days between today and New Year 2026?",
		"Read the content of README.md file",
		"What's the weather API endpoint format?",
		"Calculate compound interest for $5000 at 3% for 5 years",
	}

	for _, example := range examples {
		fmt.Printf("  • %s\n", example)
	}
	fmt.Println(strings.Repeat("=", 70) + "\n")
}

func main() {
	fmt.Println("\n🚀 Starting Personal Assistant...")

	// Create assistant
	assistant, err := NewPersonalAssistant()
	if err != nil {
		fmt.Printf("❌ Failed to create assistant: %v\n", err)
		os.Exit(1)
	}

	// Show initial status
	time.Sleep(500 * time.Millisecond) // Let initialization complete
	assistant.ShowStatus()

	// Ask user what to do
	fmt.Println("Choose mode:")
	fmt.Println("  1. Run Demo Scenarios (recommended for first-time users)")
	fmt.Println("  2. Interactive Chat Mode")
	fmt.Print("\nYour choice (1 or 2): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		assistant.RunDemo()

		// After demo, ask if want to continue with interactive mode
		fmt.Print("\n\n💬 Would you like to continue with interactive chat? (y/n): ")
		cont, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(cont)) == "y" {
			assistant.RunInteractive()
		}
	case "2":
		assistant.RunInteractive()
	default:
		fmt.Println("Invalid choice. Running demo scenarios...")
		assistant.RunDemo()
	}

	fmt.Println("\n✨ Thank you for using Personal Assistant!")
}

// Helper functions for reading environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
		if value == "false" || value == "0" || value == "no" {
			return false
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}
