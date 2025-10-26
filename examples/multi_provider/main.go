package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/taipm/go-llm-agent/pkg/provider"
	"github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	fmt.Println("🤖 Multi-Provider Chat Demo")
	fmt.Println("============================")
	fmt.Println()

	// Show 3 ways to create provider
	demoProviderCreation()

	// Interactive chat using FromEnv()
	runInteractiveChat()
}

func demoProviderCreation() {
	fmt.Println("📦 Three Ways to Create a Provider:")
	fmt.Println()

	// Way 1: FromEnv() - Auto-detect from environment variables
	fmt.Println("1️⃣  Auto-detect from environment variables:")
	fmt.Println("   provider, err := provider.FromEnv()")
	fmt.Println()

	// Way 2: Manual config - Ollama
	fmt.Println("2️⃣  Manual config - Ollama:")
	fmt.Println("   provider, err := provider.New(provider.Config{")
	fmt.Println("       Type:    provider.ProviderOllama,")
	fmt.Println("       BaseURL: \"http://localhost:11434\",")
	fmt.Println("       Model:   \"llama3.2\",")
	fmt.Println("   })")
	fmt.Println()

	// Way 3: Manual config - OpenAI
	fmt.Println("3️⃣  Manual config - OpenAI:")
	fmt.Println("   provider, err := provider.New(provider.Config{")
	fmt.Println("       Type:   provider.ProviderOpenAI,")
	fmt.Println("       APIKey: os.Getenv(\"OPENAI_API_KEY\"),")
	fmt.Println("       Model:  \"gpt-4o\",")
	fmt.Println("   })")
	fmt.Println()

	// Way 4: Manual config - Gemini
	fmt.Println("4️⃣  Manual config - Gemini:")
	fmt.Println("   provider, err := provider.New(provider.Config{")
	fmt.Println("       Type:   provider.ProviderGemini,")
	fmt.Println("       APIKey: os.Getenv(\"GEMINI_API_KEY\"),")
	fmt.Println("       Model:  \"gemini-2.5-flash\",")
	fmt.Println("   })")
	fmt.Println()

	// Way 5: Manual config - Azure OpenAI
	fmt.Println("5️⃣  Manual config - Azure OpenAI:")
	fmt.Println("   provider, err := provider.New(provider.Config{")
	fmt.Println("       Type:    provider.ProviderOpenAI,")
	fmt.Println("       APIKey:  os.Getenv(\"OPENAI_API_KEY\"),")
	fmt.Println("       BaseURL: \"https://mycompany.openai.azure.com\",")
	fmt.Println("       Model:   \"gpt-4o\",")
	fmt.Println("   })")
	fmt.Println()

	// Way 6: Manual config - Vertex AI
	fmt.Println("6️⃣  Manual config - Vertex AI:")
	fmt.Println("   provider, err := provider.New(provider.Config{")
	fmt.Println("       Type:      provider.ProviderGemini,")
	fmt.Println("       ProjectID: \"my-gcp-project\",")
	fmt.Println("       Location:  \"us-central1\",")
	fmt.Println("       Model:     \"gemini-2.5-flash\",")
	fmt.Println("   })")
	fmt.Println()
	fmt.Println("============================")
	fmt.Println()
}

func runInteractiveChat() {
	// Create provider from environment variables
	llm, err := provider.FromEnv()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Detect which provider is being used
	providerType := os.Getenv("LLM_PROVIDER")
	if providerType == "" {
		providerType = "unknown"
	}
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "unknown"
	}

	fmt.Printf("✅ Using provider: %s (model: %s)\n", providerType, model)
	fmt.Println()
	fmt.Println("💬 Interactive Chat Mode")
	fmt.Println("Type your message and press Enter. Type 'quit' to exit.")
	fmt.Println()

	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)
	conversationHistory := []types.Message{}

	for {
		fmt.Print("You: ")
		userInput, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}

		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			continue
		}

		if strings.ToLower(userInput) == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		// Add user message to history
		conversationHistory = append(conversationHistory, types.Message{
			Role:    types.RoleUser,
			Content: userInput,
		})

		// Get response from LLM
		fmt.Print("Assistant: ")
		resp, err := llm.Chat(ctx, conversationHistory, nil)
		if err != nil {
			log.Printf("Error: %v\n", err)
			// Remove the failed message from history
			conversationHistory = conversationHistory[:len(conversationHistory)-1]
			continue
		}

		fmt.Println(resp.Content)
		fmt.Println()

		// Add assistant response to history
		conversationHistory = append(conversationHistory, types.Message{
			Role:    types.RoleAssistant,
			Content: resp.Content,
		})
	}
}
