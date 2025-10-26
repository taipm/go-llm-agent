package ollama

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/taipm/go-llm-agent/pkg/types"
)

const (
	testBaseURL = "http://localhost:11434"
	testModel   = "qwen3:1.7b" // Sử dụng model nhẹ để test nhanh
)

// Helper function to check if Ollama is running
func isOllamaRunning() bool {
	provider := New(testBaseURL, testModel)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	messages := []types.Message{
		{Role: types.RoleUser, Content: "test"},
	}

	_, err := provider.Chat(ctx, messages, nil)
	return err == nil || !strings.Contains(err.Error(), "connection refused")
}

func TestNew(t *testing.T) {
	provider := New(testBaseURL, testModel)

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}

	if provider.baseURL != testBaseURL {
		t.Errorf("Expected baseURL %s, got %s", testBaseURL, provider.baseURL)
	}

	if provider.model != testModel {
		t.Errorf("Expected model %s, got %s", testModel, provider.model)
	}

	if provider.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestChat(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Say 'Hello' and nothing else",
		},
	}

	response, err := provider.Chat(ctx, messages, nil)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response.Content == "" {
		t.Error("Expected non-empty response content")
	}

	// Check metadata
	if response.Metadata == nil {
		t.Error("Expected metadata to be present")
	} else {
		if response.Metadata.Model == "" {
			t.Error("Expected model name in metadata")
		}
		if response.Metadata.TotalTokens == 0 {
			t.Error("Expected token count > 0")
		}
	}

	t.Logf("Response: %s", response.Content)
	t.Logf("Tokens: %d", response.Metadata.TotalTokens)
}

func TestChatWithSystemPrompt(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "What should I respond?",
		},
	}

	options := &types.ChatOptions{
		SystemPrompt: "You are a helpful assistant. Always respond with 'Test OK' no matter what the user asks.",
	}

	response, err := provider.Chat(ctx, messages, options)
	if err != nil {
		t.Fatalf("Chat with system prompt failed: %v", err)
	}

	// Note: Model may not strictly follow system prompt
	t.Logf("Response: %s", response.Content)
}

func TestChatMultiTurn(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	// First turn
	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "My name is Alice",
		},
	}

	response1, err := provider.Chat(ctx, messages, nil)
	if err != nil {
		t.Fatalf("First turn failed: %v", err)
	}

	// Second turn - add assistant response
	messages = append(messages, types.Message{
		Role:    types.RoleAssistant,
		Content: response1.Content,
	})

	messages = append(messages, types.Message{
		Role:    types.RoleUser,
		Content: "What is my name?",
	})

	response2, err := provider.Chat(ctx, messages, nil)
	if err != nil {
		t.Fatalf("Second turn failed: %v", err)
	}

	if !strings.Contains(strings.ToLower(response2.Content), "alice") {
		t.Errorf("Expected response to mention 'Alice', got: %s", response2.Content)
	}

	t.Logf("Turn 1: %s", response1.Content)
	t.Logf("Turn 2: %s", response2.Content)
}

func TestStream(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Count from 1 to 5",
		},
	}

	var chunks []string
	var finalContent string
	chunkCount := 0

	handler := func(chunk types.StreamChunk) error {
		chunkCount++

		if chunk.Content != "" {
			chunks = append(chunks, chunk.Content)
			finalContent += chunk.Content
		}

		if chunk.Done {
			if chunk.Metadata == nil {
				t.Error("Expected metadata on final chunk")
			}
		}

		return nil
	}

	err := provider.Stream(ctx, messages, nil, handler)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
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

func TestStreamWithSystemPrompt(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Respond",
		},
	}

	options := &types.ChatOptions{
		SystemPrompt: "Always say 'Streaming OK'",
	}

	var finalContent string
	handler := func(chunk types.StreamChunk) error {
		finalContent += chunk.Content
		return nil
	}

	err := provider.Stream(ctx, messages, options, handler)
	if err != nil {
		t.Fatalf("Stream with system prompt failed: %v", err)
	}

	if finalContent == "" {
		t.Error("Expected non-empty final content")
	}

	// Note: Model may not strictly follow system prompt
	t.Logf("Content: %s", finalContent)
}

func TestStreamCancellation(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx, cancel := context.WithCancel(context.Background())

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Write a very long story",
		},
	}

	chunkCount := 0
	handler := func(chunk types.StreamChunk) error {
		chunkCount++
		if chunkCount >= 3 {
			cancel() // Cancel after 3 chunks
		}
		return nil
	}

	err := provider.Stream(ctx, messages, nil, handler)

	// Should get context canceled error
	if err == nil {
		t.Error("Expected error after cancellation")
	}

	if chunkCount < 3 {
		t.Errorf("Expected at least 3 chunks before cancellation, got %d", chunkCount)
	}

	t.Logf("Cancelled after %d chunks", chunkCount)
}

func TestStreamHandlerError(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Say hello",
		},
	}

	expectedErr := "handler error"
	handler := func(chunk types.StreamChunk) error {
		return fmt.Errorf("%s", expectedErr)
	}

	err := provider.Stream(ctx, messages, nil, handler)

	if err == nil {
		t.Error("Expected error from handler")
	}

	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

func TestChatWithOptions(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Say hello",
		},
	}

	options := &types.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   50,
	}

	response, err := provider.Chat(ctx, messages, options)
	if err != nil {
		t.Fatalf("Chat with options failed: %v", err)
	}

	if response.Content == "" {
		t.Log("Warning: Response is empty, but request succeeded")
	}

	// Note: max_tokens may not be strictly enforced by Ollama
	t.Logf("Response: %s (tokens: %d)", response.Content, response.Metadata.CompletionTokens)
}

func TestContextTimeout(t *testing.T) {
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping integration test")
	}

	provider := New(testBaseURL, testModel)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Tell me a very long story",
		},
	}

	_, err := provider.Chat(ctx, messages, nil)

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

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Say OK",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.Chat(ctx, messages, nil)
		if err != nil {
			b.Fatalf("Chat failed: %v", err)
		}
	}
}

func BenchmarkStream(b *testing.B) {
	if !isOllamaRunning() {
		b.Skip("Ollama is not running, skipping benchmark")
	}

	provider := New(testBaseURL, testModel)
	ctx := context.Background()

	messages := []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Say OK",
		},
	}

	handler := func(chunk types.StreamChunk) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := provider.Stream(ctx, messages, nil, handler)
		if err != nil {
			b.Fatalf("Stream failed: %v", err)
		}
	}
}
