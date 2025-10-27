package reasoning

import (
	"context"
	"testing"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// MockProvider for testing
type MockProvider struct {
	responses []string
	callCount int
}

func (m *MockProvider) Chat(ctx context.Context, messages []types.Message, options *types.ChatOptions) (*types.Response, error) {
	if m.callCount >= len(m.responses) {
		return &types.Response{Content: "Final answer"}, nil
	}
	response := m.responses[m.callCount]
	m.callCount++
	return &types.Response{Content: response}, nil
}

func (m *MockProvider) Stream(ctx context.Context, messages []types.Message, options *types.ChatOptions, handler types.StreamHandler) error {
	return nil
}

func (m *MockProvider) GetModel() string {
	return "mock-model"
}

func (m *MockProvider) GetProviderName() string {
	return "mock"
}

// MockMemory for testing
type MockMemory struct {
	messages []types.Message
}

func (m *MockMemory) Add(message types.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *MockMemory) GetHistory(limit int) ([]types.Message, error) {
	return m.messages, nil
}

func (m *MockMemory) Clear() error {
	m.messages = []types.Message{}
	return nil
}

func (m *MockMemory) Size() int {
	return len(m.messages)
}

func TestNewReActAgent(t *testing.T) {
	provider := &MockProvider{}
	memory := &MockMemory{}
	agent := NewReActAgent(provider, memory, 5)
	
	if agent == nil {
		t.Fatal("NewReActAgent returned nil")
	}
	if agent.maxSteps != 5 {
		t.Errorf("Expected maxSteps=5, got %d", agent.maxSteps)
	}
	if !agent.verbose {
		t.Error("Expected verbose=true by default")
	}
}

func TestParseReActResponse(t *testing.T) {
	provider := &MockProvider{}
	memory := &MockMemory{}
	agent := NewReActAgent(provider, memory, 5)
	
	response := "Thought: I need to check\nAction: get_weather"
	thought, action, _, _ := agent.parseReActResponse(response)
	
	if thought != "I need to check" {
		t.Errorf("Thought: got %q, want 'I need to check'", thought)
	}
	if action != "get_weather" {
		t.Errorf("Action: got %q, want 'get_weather'", action)
	}
}
