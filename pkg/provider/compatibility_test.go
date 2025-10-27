package provider

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/taipm/go-llm-agent/pkg/types"
)

const (
	skipNoAPIKeys     = "No API keys available for provider testing"
	skipNoProviders   = "No providers available for testing"
	skipNoToolSupport = "Provider does not support tool calling"
)

// TestCompatibilityChat tests that all providers handle basic chat requests consistently
func TestCompatibilityChat(t *testing.T) {
	// Skip if no API keys are available
	if !hasAnyProvider() {
		t.Skip(skipNoAPIKeys)
	}

	tests := []struct {
		name     string
		question string
		validate func(t *testing.T, response string, provider string)
	}{
		{
			name:     "simple_math",
			question: "What is 2+2? Answer with just the number.",
			validate: func(t *testing.T, response string, provider string) {
				if !strings.Contains(response, "4") {
					t.Errorf("%s: Expected response to contain '4', got: %s", provider, response)
				}
			},
		},
		{
			name:     "capital_city",
			question: "What is the capital of France? Answer with just the city name.",
			validate: func(t *testing.T, response string, provider string) {
				if !strings.Contains(strings.ToLower(response), "paris") {
					t.Errorf("%s: Expected response to contain 'Paris', got: %s", provider, response)
				}
			},
		},
		{
			name:     "yes_no_question",
			question: "Is the Earth round? Answer with just yes or no.",
			validate: func(t *testing.T, response string, provider string) {
				lower := strings.ToLower(response)
				if !strings.Contains(lower, "yes") {
					t.Errorf("%s: Expected response to contain 'yes', got: %s", provider, response)
				}
			},
		},
	}

	providers := getAvailableProviders(t)
	if len(providers) == 0 {
		t.Skip(skipNoProviders)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testChatQuestion(t, providers, tt.question, tt.validate)
		})
	}
}

// testChatQuestion is a helper to test a question across all providers
func testChatQuestion(t *testing.T, providers map[string]types.LLMProvider, question string, validate func(t *testing.T, response string, provider string)) {
	ctx := context.Background()
	results := make(map[string]chatResult)

	// Test each provider
	for name, provider := range providers {
		messages := []types.Message{
			{Role: types.RoleUser, Content: question},
		}

		start := time.Now()
		response, err := provider.Chat(ctx, messages, nil)
		duration := time.Since(start)

		results[name] = chatResult{
			response: response,
			err:      err,
			duration: duration,
		}

		if err != nil {
			t.Logf("%s: Error: %v", name, err)
			continue
		}

		t.Logf("%s: Response (%.2fs): %s", name, duration.Seconds(), truncate(response.Content, 100))
		validate(t, response.Content, name)
	}

	// Log comparison
	t.Logf("\nComparison:")
	for name, result := range results {
		if result.err != nil {
			t.Logf("  %s: ERROR - %v", name, result.err)
		} else {
			t.Logf("  %s: %.2fs - %s", name, result.duration.Seconds(), truncate(result.response.Content, 80))
		}
	}
}

// TestCompatibilityChatWithHistory tests conversation history handling
func TestCompatibilityChatWithHistory(t *testing.T) {
	if !hasAnyProvider() {
		t.Skip(skipNoAPIKeys)
	}

	providers := getAvailableProviders(t)
	if len(providers) == 0 {
		t.Skip(skipNoProviders)
	}

	ctx := context.Background()

	for name, provider := range providers {
		t.Run(name, func(t *testing.T) {
			// Build conversation history
			history := []types.Message{
				{Role: types.RoleUser, Content: "My name is Alice."},
				{Role: types.RoleAssistant, Content: "Nice to meet you, Alice!"},
				{Role: types.RoleUser, Content: "What is my name?"},
			}

			response, err := provider.Chat(ctx, history, nil)
			if err != nil {
				t.Fatalf("Chat with history failed: %v", err)
			}
			t.Logf("Response: %s", truncate(response.Content, 100))

			// Verify the provider remembered the name
			if !strings.Contains(strings.ToLower(response.Content), "alice") {
				t.Errorf("Expected response to contain 'Alice', got: %s", response.Content)
			}
		})
	}
}

// TestCompatibilityStream tests streaming behavior across providers
func TestCompatibilityStream(t *testing.T) {
	if !hasAnyProvider() {
		t.Skip(skipNoAPIKeys)
	}

	providers := getAvailableProviders(t)
	if len(providers) == 0 {
		t.Skip(skipNoProviders)
	}

	ctx := context.Background()
	question := "Count from 1 to 5, one number per line."

	for name, provider := range providers {
		t.Run(name, func(t *testing.T) {
			messages := []types.Message{
				{Role: types.RoleUser, Content: question},
			}

			fullText := ""
			chunkCount := 0
			start := time.Now()

			err := provider.Stream(ctx, messages, nil, func(chunk types.StreamChunk) error {
				if chunk.Error != nil {
					return chunk.Error
				}

				chunkCount++
				fullText += chunk.Content

				// Log first few chunks
				if chunkCount <= 3 {
					t.Logf("Chunk %d: %q", chunkCount, chunk.Content)
				}

				return nil
			})

			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Stream error: %v", err)
			}

			t.Logf("Received %d chunks in %.2fs", chunkCount, duration.Seconds())
			t.Logf("Full text: %s", truncate(fullText, 100))

			// Validate that we got some chunks
			if chunkCount == 0 {
				t.Error("Expected at least one chunk")
			}

			// Validate that the full text contains expected numbers
			for i := 1; i <= 5; i++ {
				numStr := string(rune('0' + i))
				if !strings.Contains(fullText, numStr) {
					t.Errorf("Expected response to contain '%s'", numStr)
				}
			}
		})
	}
}

// TestCompatibilityStreamWithHistory tests streaming with conversation history
func TestCompatibilityStreamWithHistory(t *testing.T) {
	if !hasAnyProvider() {
		t.Skip(skipNoAPIKeys)
	}

	providers := getAvailableProviders(t)
	if len(providers) == 0 {
		t.Skip(skipNoProviders)
	}

	ctx := context.Background()

	for name, provider := range providers {
		t.Run(name, func(t *testing.T) {
			history := []types.Message{
				{Role: types.RoleUser, Content: "My favorite color is blue."},
				{Role: types.RoleAssistant, Content: "That's nice! Blue is a great color."},
				{Role: types.RoleUser, Content: "What is my favorite color?"},
			}

			fullText := ""
			err := provider.Stream(ctx, history, nil, func(chunk types.StreamChunk) error {
				if chunk.Error != nil {
					return chunk.Error
				}
				fullText += chunk.Content
				return nil
			})

			if err != nil {
				t.Fatalf("Stream error: %v", err)
			}

			t.Logf("Response: %s", truncate(fullText, 100))

			// Verify the provider remembered the color
			if !strings.Contains(strings.ToLower(fullText), "blue") {
				t.Errorf("Expected response to contain 'blue', got: %s", fullText)
			}
		})
	}
}

// TestCompatibilityToolCalling tests tool/function calling across providers
func TestCompatibilityToolCalling(t *testing.T) {
	if !hasAnyProvider() {
		t.Skip(skipNoAPIKeys)
	}

	providers := getAvailableProviders(t)
	if len(providers) == 0 {
		t.Skip(skipNoProviders)
	}

	// Define a simple weather tool
	tools := []types.ToolDefinition{
		{
			Type: "function",
			Function: types.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather for a location",
				Parameters: &types.JSONSchema{
					Type: "object",
					Properties: map[string]*types.JSONSchema{
						"location": {
							Type:        "string",
							Description: "The city name, e.g. Tokyo, Paris",
						},
						"unit": {
							Type:        "string",
							Enum:        []interface{}{"celsius", "fahrenheit"},
							Description: "The temperature unit",
						},
					},
					Required: []string{"location"},
				},
			},
		},
	}

	ctx := context.Background()
	question := "What's the weather in Tokyo?"

	for name, provider := range providers {
		t.Run(name, func(t *testing.T) {
			messages := []types.Message{
				{Role: types.RoleUser, Content: question},
			}

			options := &types.ChatOptions{
				Tools: tools,
			}

			response, err := provider.Chat(ctx, messages, options)
			if err != nil {
				// Log error but don't fail - tool support is optional
				t.Logf("%s does not support tool calling or error occurred: %v", name, err)
				return
			}

			t.Logf("Response: %s", truncate(response.Content, 100))

			// If tool calls were made, validate them
			if len(response.ToolCalls) > 0 {
				t.Logf("Tool calls made: %d", len(response.ToolCalls))
				validateToolCalls(t, response.ToolCalls)
			} else {
				t.Logf("No tool calls made (provider may have answered directly)")
			}
		})
	}
}

// validateToolCalls validates tool call responses
func validateToolCalls(t *testing.T, toolCalls []types.ToolCall) {
	for i, tc := range toolCalls {
		t.Logf("  Tool %d: ID=%s, Type=%s, Function=%s", i+1, tc.ID, tc.Type, tc.Function.Name)

		// Validate that the tool call is for weather
		if tc.Function.Name != "get_weather" {
			t.Errorf("Expected tool name 'get_weather', got: %s", tc.Function.Name)
		}

		// Validate that the function has arguments
		if tc.Function.Arguments == nil || len(tc.Function.Arguments) == 0 {
			t.Errorf("Expected tool arguments, got none")
		} else {
			t.Logf("    Arguments: %v", tc.Function.Arguments)
		}
	}
}

// TestCompatibilityErrorHandling tests consistent error handling
func TestCompatibilityErrorHandling(t *testing.T) {
	if !hasAnyProvider() {
		t.Skip(skipNoAPIKeys)
	}

	providers := getAvailableProviders(t)
	if len(providers) == 0 {
		t.Skip(skipNoProviders)
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		testFunc func(t *testing.T, provider types.LLMProvider) error
	}{
		{
			name: "empty_message",
			testFunc: func(t *testing.T, provider types.LLMProvider) error {
				_, err := provider.Chat(ctx, []types.Message{{Role: types.RoleUser, Content: ""}}, nil)
				return err
			},
		},
		{
			name: "empty_history",
			testFunc: func(t *testing.T, provider types.LLMProvider) error {
				_, err := provider.Chat(ctx, []types.Message{}, nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, provider := range providers {
				err := tt.testFunc(t, provider)
				if err != nil {
					t.Logf("%s: Error (as expected): %v", name, err)
				} else {
					t.Logf("%s: No error (provider handled gracefully)", name)
				}
			}
		})
	}
}

// Helper types and functions

type chatResult struct {
	response *types.Response
	err      error
	duration time.Duration
}

// hasAnyProvider checks if any provider API keys are available
func hasAnyProvider() bool {
	return os.Getenv("OPENAI_API_KEY") != "" ||
		os.Getenv("GEMINI_API_KEY") != "" ||
		os.Getenv("OLLAMA_BASE_URL") != "" ||
		checkOllamaDefault()
}

// checkOllamaDefault checks if Ollama is available at default URL
func checkOllamaDefault() bool {
	// We'll assume Ollama might be available locally
	// The actual test will fail gracefully if not
	return true
}

// getAvailableProviders creates providers for all available API keys
func getAvailableProviders(t *testing.T) map[string]types.LLMProvider {
	providers := make(map[string]types.LLMProvider)

	// Try Ollama (local, no API key needed)
	addOllamaProvider(t, providers)

	// Try OpenAI
	addOpenAIProvider(t, providers)

	// Try Gemini
	addGeminiProvider(t, providers)

	return providers
}

// addOllamaProvider tries to add Ollama provider
func addOllamaProvider(t *testing.T, providers map[string]types.LLMProvider) {
	ollamaURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "qwen3:1.7b" // default test model
	}

	provider, err := New(Config{
		Type:    ProviderOllama,
		BaseURL: ollamaURL,
		Model:   model,
	})
	if err == nil {
		providers["ollama"] = provider
		t.Logf("Added Ollama provider: %s @ %s", model, ollamaURL)
	}
}

// addOpenAIProvider tries to add OpenAI provider
func addOpenAIProvider(t *testing.T, providers map[string]types.LLMProvider) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini" // default test model
	}

	provider, err := New(Config{
		Type:   ProviderOpenAI,
		APIKey: apiKey,
		Model:  model,
	})
	if err == nil {
		providers["openai"] = provider
		t.Logf("Added OpenAI provider: %s", model)
	}
}

// addGeminiProvider tries to add Gemini provider
func addGeminiProvider(t *testing.T, providers map[string]types.LLMProvider) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash" // default test model
	}

	provider, err := New(Config{
		Type:   ProviderGemini,
		APIKey: apiKey,
		Model:  model,
	})
	if err == nil {
		providers["gemini"] = provider
		t.Logf("Added Gemini provider: %s", model)
	}
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
