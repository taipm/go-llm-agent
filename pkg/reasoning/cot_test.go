package reasoning

import (
	"context"
	"strings"
	"testing"

	"github.com/taipm/go-llm-agent/pkg/types"
)

func TestNewCoTAgent(t *testing.T) {
	provider := &MockProvider{}
	memory := &MockMemory{}

	agent := NewCoTAgent(provider, memory, 5)

	if agent == nil {
		t.Fatal("Expected non-nil agent")
	}

	if agent.maxSteps != 5 {
		t.Errorf("Expected maxSteps 5, got %d", agent.maxSteps)
	}

	// Test default maxSteps
	agent2 := NewCoTAgent(provider, memory, 0)
	if agent2.maxSteps != 10 {
		t.Errorf("Expected default maxSteps 10, got %d", agent2.maxSteps)
	}
}

func TestBuildCoTPrompt(t *testing.T) {
	provider := &MockProvider{}
	memory := &MockMemory{}
	agent := NewCoTAgent(provider, memory, 5)

	question := "What is 2 + 2?"
	prompt := agent.buildCoTPrompt(question)

	// Check that prompt contains key elements
	if !strings.Contains(prompt, "step-by-step") {
		t.Error("Prompt should mention step-by-step reasoning")
	}

	if !strings.Contains(prompt, question) {
		t.Error("Prompt should contain the original question")
	}

	if !strings.Contains(prompt, "Step 1:") {
		t.Error("Prompt should include format example")
	}

	if !strings.Contains(prompt, "Answer:") {
		t.Error("Prompt should include answer format")
	}
}

func TestParseCoTResponse(t *testing.T) {
	provider := &MockProvider{}
	memory := &MockMemory{}
	agent := NewCoTAgent(provider, memory, 5)

	tests := []struct {
		name           string
		response       string
		expectedSteps  int
		expectedAnswer string
	}{
		{
			name: "Structured response",
			response: `Step 1: First, we add 2 + 2
Step 2: The result is 4

Answer: 4`,
			expectedSteps:  2,
			expectedAnswer: "4",
		},
		{
			name: "Numbered list response",
			response: `1. Calculate 10 * 2 = 20
2. Add 5 to get 25

The answer is 25`,
			expectedSteps:  2,
			expectedAnswer: "The answer is 25",
		},
		{
			name: "Case insensitive",
			response: `STEP 1: Calculate the sum
STEP 2: Get result

ANSWER: 42`,
			expectedSteps:  2,
			expectedAnswer: "42",
		},
		{
			name: "Multi-line answer",
			response: `Step 1: First step
Step 2: Second step

Answer: The final answer is
that we need multiple lines
to express`,
			expectedSteps:  2,
			expectedAnswer: "The final answer is that we need multiple lines to express",
		},
		{
			name: "No explicit answer",
			response: `Step 1: Calculate 5 + 3 = 8
Step 2: Therefore the result is 8`,
			expectedSteps:  2,
			expectedAnswer: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, answer, err := agent.parseCoTResponse(tt.response)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(steps) != tt.expectedSteps {
				t.Errorf("Expected %d steps, got %d", tt.expectedSteps, len(steps))
			}

			// Check step numbering
			for i, step := range steps {
				expectedNum := i + 1
				if step.StepNumber != expectedNum {
					t.Errorf("Step %d: expected number %d, got %d", i, expectedNum, step.StepNumber)
				}
			}

			// Check answer if expected
			if tt.expectedAnswer != "" && !strings.Contains(answer, tt.expectedAnswer) {
				t.Errorf("Expected answer to contain '%s', got '%s'", tt.expectedAnswer, answer)
			}
		})
	}
}

func TestCoTThink(t *testing.T) {
	mockResponse := `Step 1: Calculate 15 * 23 = 345
Step 2: Add 47 to get 345 + 47 = 392

Answer: 392`

	provider := &MockProvider{
		responses: []string{mockResponse},
	}
	memory := &MockMemory{}
	agent := NewCoTAgent(provider, memory, 5)

	ctx := context.Background()
	answer, err := agent.Think(ctx, "What is 15 * 23 + 47?")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(answer, "392") {
		t.Errorf("Expected answer to contain '392', got '%s'", answer)
	}

	// Check that chain was created
	chain := agent.GetChain()
	if chain == nil {
		t.Fatal("Expected chain to be created")
	}

	if len(chain.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(chain.Steps))
	}

	if chain.Answer != "392" {
		t.Errorf("Expected final answer '392', got '%s'", chain.Answer)
	}
}

func TestCoTValidate(t *testing.T) {
	provider := &MockProvider{}
	memory := &MockMemory{}
	agent := NewCoTAgent(provider, memory, 5)

	// Test validation with no chain
	valid, issues := agent.Validate()
	if valid {
		t.Error("Expected validation to fail with no chain")
	}
	if len(issues) == 0 {
		t.Error("Expected validation issues")
	}

	// Create a valid chain
	agent.chain = &types.CoTChain{
		Query: "Test question",
		Steps: []types.CoTStep{
			{StepNumber: 1, Description: "First step"},
			{StepNumber: 2, Description: "Second step"},
		},
		Answer: "Final answer",
	}

	valid, issues = agent.Validate()
	if !valid {
		t.Errorf("Expected validation to pass, got issues: %v", issues)
	}

	// Test with invalid step numbering
	agent.chain.Steps[1].StepNumber = 5
	valid, issues = agent.Validate()
	if valid {
		t.Error("Expected validation to fail with wrong step numbering")
	}

	// Test with empty step
	agent.chain.Steps[1].StepNumber = 2
	agent.chain.Steps[1].Description = ""
	valid, issues = agent.Validate()
	if valid {
		t.Error("Expected validation to fail with empty step")
	}

	// Test with no answer
	agent.chain.Steps[1].Description = "Valid step"
	agent.chain.Answer = ""
	valid, issues = agent.Validate()
	if valid {
		t.Error("Expected validation to fail with no final answer")
	}
}

func TestCoTGetReasoningHistory(t *testing.T) {
	provider := &MockProvider{}
	memory := &MockMemory{}
	agent := NewCoTAgent(provider, memory, 5)

	// Test with no chain
	history := agent.GetReasoningHistory()
	if !strings.Contains(history, "No reasoning history") {
		t.Error("Expected message about no history")
	}

	// Create a chain
	agent.chain = &types.CoTChain{
		Query: "What is 2+2?",
		Steps: []types.CoTStep{
			{StepNumber: 1, Description: "Add 2 and 2"},
			{StepNumber: 2, Description: "Get 4"},
		},
		Answer: "4",
	}

	history = agent.GetReasoningHistory()

	// Check that history contains key elements
	if !strings.Contains(history, "Chain-of-Thought") {
		t.Error("History should contain title")
	}

	if !strings.Contains(history, "What is 2+2?") {
		t.Error("History should contain question")
	}

	if !strings.Contains(history, "Step 1") {
		t.Error("History should contain step 1")
	}

	if !strings.Contains(history, "Step 2") {
		t.Error("History should contain step 2")
	}

	if !strings.Contains(history, "Final Answer: 4") {
		t.Error("History should contain final answer")
	}
}

func TestCoTExceedsMaxSteps(t *testing.T) {
	// Create mock with many steps
	mockResponse := `Step 1: First
Step 2: Second
Step 3: Third
Step 4: Fourth
Step 5: Fifth
Step 6: Sixth

Answer: Too many steps`

	provider := &MockProvider{
		responses: []string{mockResponse},
	}
	memory := &MockMemory{}
	agent := NewCoTAgent(provider, memory, 3) // Max 3 steps

	ctx := context.Background()
	_, err := agent.Think(ctx, "Complex question")

	if err == nil {
		t.Error("Expected error for exceeding max steps")
	}

	if !strings.Contains(err.Error(), "exceeded max steps") {
		t.Errorf("Expected 'exceeded max steps' error, got: %v", err)
	}
}
