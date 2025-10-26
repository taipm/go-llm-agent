package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/taipm/go-llm-agent/pkg/memory"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
	"github.com/taipm/go-llm-agent/pkg/types"
)

const (
	testBaseURL = "http://localhost:11434"
	testModel   = "qwen3:1.7b"
)

// Helper to check if Ollama is running
func isOllamaRunning() bool {
	provider := ollama.New(testBaseURL, testModel)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	messages := []types.Message{
		{Role: types.RoleUser, Content: "test"},
	}

	_, err := provider.Chat(ctx, messages, nil)
	return err == nil || !strings.Contains(err.Error(), "connection refused")
}

// Simple test tool for testing
type TestTool struct{}

func (t *TestTool) Name() string {
	return "test_tool"
}

func (t *TestTool) Description() string {
	return "A simple test tool that returns a test message"
}

func (t *TestTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"message": {
				Type:        "string",
				Description: "Test message parameter",
			},
		},
		Required: []string{"message"},
	}
}

func (t *TestTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	msg, ok := params["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message parameter must be a string")
	}
	return fmt.Sprintf("Test tool received: %s", msg), nil
}

func TestNewAgent(t *testing.T) {
	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)

	if ag == nil {
		t.Fatal("Expected agent to be created")
	}

	if ag.provider == nil {
		t.Error("Expected provider to be set")
	}

	if ag.tools == nil {
		t.Error("Expected tools registry to be initialized")
	}
}

func TestNewAgentWithMemory(t *testing.T) {
	provider := ollama.New(testBaseURL, testModel)
	mem := memory.NewBuffer(10)
	ag := New(provider, WithMemory(mem))

	if ag.memory == nil {
		t.Error("Expected memory to be set")
	}
}

func TestAddTool(t *testing.T) {
	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)

	testTool := &TestTool{}
	ag.AddTool(testTool)

	// Verify tool was added
	tools := ag.tools.GetDefinitions()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	if tools[0].Function.Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got %s", tools[0].Function.Name)
	}
}

func TestChat(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)
	ctx := context.Background()

	response, err := ag.Chat(ctx, "Say 'Hello Test'")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Response: %s", response)
}

func TestChatWithMemory(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	mem := memory.NewBuffer(100)
	ag := New(provider, WithMemory(mem))
	ctx := context.Background()

	// First message
	_, err := ag.Chat(ctx, "My favorite color is blue")
	if err != nil {
		t.Fatalf("First chat failed: %v", err)
	}

	// Check memory has messages
	history, err := mem.GetHistory(10)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) < 2 { // user + assistant
		t.Errorf("Expected at least 2 messages in history, got %d", len(history))
	}

	// Second message - should remember context
	response, err := ag.Chat(ctx, "What is my favorite color?")
	if err != nil {
		t.Fatalf("Second chat failed: %v", err)
	}

	if !strings.Contains(strings.ToLower(response), "blue") {
		t.Errorf("Expected response to mention 'blue', got: %s", response)
	}

	t.Logf("Response: %s", response)
}

func TestChatStream(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)
	ctx := context.Background()

	var chunks []string
	var finalContent string
	chunkCount := 0

	handler := func(chunk types.StreamChunk) error {
		chunkCount++

		if chunk.Content != "" {
			chunks = append(chunks, chunk.Content)
			finalContent += chunk.Content
		}

		return nil
	}

	err := ag.ChatStream(ctx, "Count from 1 to 3", handler)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	if chunkCount == 0 {
		t.Error("Expected at least one chunk")
	}

	if finalContent == "" {
		t.Error("Expected non-empty final content")
	}

	t.Logf("Received %d chunks", chunkCount)
	t.Logf("Final content: %s", finalContent)
}

func TestChatStreamWithMemory(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	mem := memory.NewBuffer(100)
	ag := New(provider, WithMemory(mem))
	ctx := context.Background()

	var finalContent string
	handler := func(chunk types.StreamChunk) error {
		finalContent += chunk.Content
		return nil
	}

	// First message
	err := ag.ChatStream(ctx, "My name is Bob", handler)
	if err != nil {
		t.Fatalf("First ChatStream failed: %v", err)
	}

	// Check memory
	history, err := mem.GetHistory(10)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) < 2 {
		t.Errorf("Expected at least 2 messages in history, got %d", len(history))
	}

	// Second message - should remember
	finalContent = ""
	err = ag.ChatStream(ctx, "What is my name?", handler)
	if err != nil {
		t.Fatalf("Second ChatStream failed: %v", err)
	}

	if !strings.Contains(strings.ToLower(finalContent), "bob") {
		t.Errorf("Expected response to mention 'Bob', got: %s", finalContent)
	}

	t.Logf("Response: %s", finalContent)
}

func TestChatStreamCancellation(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)
	ctx, cancel := context.WithCancel(context.Background())

	chunkCount := 0
	handler := func(chunk types.StreamChunk) error {
		chunkCount++
		if chunkCount >= 3 {
			cancel()
		}
		return nil
	}

	err := ag.ChatStream(ctx, "Tell me a long story", handler)

	if err == nil {
		t.Error("Expected error after cancellation")
	}

	if chunkCount < 3 {
		t.Errorf("Expected at least 3 chunks before cancellation, got %d", chunkCount)
	}

	t.Logf("Cancelled after %d chunks", chunkCount)
}

func TestChatStreamHandlerError(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)
	ctx := context.Background()

	expectedErr := "handler error"
	handler := func(chunk types.StreamChunk) error {
		return fmt.Errorf("%s", expectedErr)
	}

	err := ag.ChatStream(ctx, "Say hello", handler)

	if err == nil {
		t.Error("Expected error from handler")
	}

	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

func TestChatWithSystemPrompt(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider, WithSystemPrompt("You are a helpful assistant. Always respond with 'System OK'"))
	ctx := context.Background()

	response, err := ag.Chat(ctx, "Respond")
	if err != nil {
		t.Fatalf("Chat with system prompt failed: %v", err)
	}

	// Note: Model may not strictly follow system prompt
	t.Logf("Response: %s", response)
}

func TestChatWithOptions(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider, WithTemperature(0.7), WithMaxTokens(50))
	ctx := context.Background()

	response, err := ag.Chat(ctx, "Say hello")
	if err != nil {
		t.Fatalf("Chat with options failed: %v", err)
	}

	if response == "" {
		t.Log("Warning: Response is empty")
	}

	t.Logf("Response: %s", response)
}

func TestChatMultipleCalls(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	mem := memory.NewBuffer(100)
	ag := New(provider, WithMemory(mem))
	ctx := context.Background()

	// Make 3 calls
	for i := 1; i <= 3; i++ {
		response, err := ag.Chat(ctx, fmt.Sprintf("This is message number %d", i))
		if err != nil {
			t.Fatalf("Chat %d failed: %v", i, err)
		}
		t.Logf("Response %d: %s", i, response)
	}

	// Verify memory has all messages
	history, err := mem.GetHistory(100)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) < 6 { // 3 user + 3 assistant
		t.Errorf("Expected at least 6 messages, got %d", len(history))
	}
}

func TestContextTimeout(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := ag.Chat(ctx, "Tell me a very long story")

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline error, got: %v", err)
	}
}

func BenchmarkChat(b *testing.B) {
	if !isOllamaRunning() {
		b.Skip("Ollama is not running, skipping benchmark")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ag.Chat(ctx, "Say OK")
		if err != nil {
			b.Fatalf("Chat failed: %v", err)
		}
	}
}

func BenchmarkChatStream(b *testing.B) {
	if !isOllamaRunning() {
		b.Skip("Ollama is not running, skipping benchmark")
	}

	provider := ollama.New(testBaseURL, testModel)
	ag := New(provider)
	ctx := context.Background()

	handler := func(chunk types.StreamChunk) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ag.ChatStream(ctx, "Say OK", handler)
		if err != nil {
			b.Fatalf("ChatStream failed: %v", err)
		}
	}
}
