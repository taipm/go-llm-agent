package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// Provider implements the LLMProvider interface for Ollama
type Provider struct {
	baseURL string
	model   string
	client  *http.Client
}

// New creates a new Ollama provider
func New(baseURL, model string) *Provider {
	return &Provider{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// WithHTTPClient sets a custom HTTP client
func (p *Provider) WithHTTPClient(client *http.Client) *Provider {
	p.client = client
	return p
}

// ollamaMessage represents a message in Ollama's format
type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	Images    []string         `json:"images,omitempty"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

// ollamaToolCall represents a tool call in Ollama's format
type ollamaToolCall struct {
	Function ollamaFunction `json:"function"`
}

// ollamaFunction represents a function call in Ollama's format
type ollamaFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ollamaTool represents a tool definition in Ollama's format
type ollamaTool struct {
	Type     string            `json:"type"`
	Function ollamaFunctionDef `json:"function"`
}

// ollamaFunctionDef represents a function definition in Ollama's format
type ollamaFunctionDef struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  *types.JSONSchema `json:"parameters"`
}

// ollamaRequest represents the request body for Ollama API
type ollamaRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Tools    []ollamaTool           `json:"tools,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ollamaResponse represents the response from Ollama API
type ollamaResponse struct {
	Model           string        `json:"model"`
	CreatedAt       time.Time     `json:"created_at"`
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	TotalDuration   int64         `json:"total_duration,omitempty"`
	LoadDuration    int64         `json:"load_duration,omitempty"`
	PromptEvalCount int           `json:"prompt_eval_count,omitempty"`
	EvalCount       int           `json:"eval_count,omitempty"`
}

// Chat implements the LLMProvider interface
func (p *Provider) Chat(ctx context.Context, messages []types.Message, options *types.ChatOptions) (*types.Response, error) {
	// Convert messages to Ollama format
	ollamaMessages := make([]ollamaMessage, 0, len(messages))

	// Add system prompt if provided
	if options != nil && options.SystemPrompt != "" {
		ollamaMessages = append(ollamaMessages, ollamaMessage{
			Role:    "system",
			Content: options.SystemPrompt,
		})
	}

	for _, msg := range messages {
		ollamaMsg := ollamaMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			ollamaMsg.ToolCalls = make([]ollamaToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				ollamaMsg.ToolCalls[i] = ollamaToolCall{
					Function: ollamaFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		ollamaMessages = append(ollamaMessages, ollamaMsg)
	}

	// Build request
	reqBody := ollamaRequest{
		Model:    p.model,
		Messages: ollamaMessages,
		Stream:   false,
	}

	// Add tools if provided
	if options != nil && len(options.Tools) > 0 {
		reqBody.Tools = make([]ollamaTool, len(options.Tools))
		for i, tool := range options.Tools {
			reqBody.Tools[i] = ollamaTool{
				Type: tool.Type,
				Function: ollamaFunctionDef{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
	}

	// Add options
	if options != nil {
		reqBody.Options = make(map[string]interface{})
		if options.Temperature > 0 {
			reqBody.Options["temperature"] = options.Temperature
		}
		if options.TopP > 0 {
			reqBody.Options["top_p"] = options.TopP
		}
		if options.MaxTokens > 0 {
			reqBody.Options["num_predict"] = options.MaxTokens
		}
		if len(options.Stop) > 0 {
			reqBody.Options["stop"] = options.Stop
		}
	}

	// Marshal request
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert response to our format
	response := &types.Response{
		Content: ollamaResp.Message.Content,
		Metadata: &types.Metadata{
			Model:            ollamaResp.Model,
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
	}

	// Convert tool calls if present
	if len(ollamaResp.Message.ToolCalls) > 0 {
		response.ToolCalls = make([]types.ToolCall, len(ollamaResp.Message.ToolCalls))
		for i, tc := range ollamaResp.Message.ToolCalls {
			response.ToolCalls[i] = types.ToolCall{
				ID:   fmt.Sprintf("call_%d", i), // Ollama doesn't provide IDs
				Type: "function",
				Function: types.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return response, nil
}

// Stream implements streaming chat completion
func (p *Provider) Stream(ctx context.Context, messages []types.Message, options *types.ChatOptions, handler types.StreamHandler) error {
	// Convert messages to Ollama format
	ollamaMessages := make([]ollamaMessage, 0, len(messages))

	// Add system prompt if provided
	if options != nil && options.SystemPrompt != "" {
		ollamaMessages = append(ollamaMessages, ollamaMessage{
			Role:    "system",
			Content: options.SystemPrompt,
		})
	}

	for _, msg := range messages {
		ollamaMsg := ollamaMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			ollamaMsg.ToolCalls = make([]ollamaToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				ollamaMsg.ToolCalls[i] = ollamaToolCall{
					Function: ollamaFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		ollamaMessages = append(ollamaMessages, ollamaMsg)
	}

	// Build request with streaming enabled
	reqBody := ollamaRequest{
		Model:    p.model,
		Messages: ollamaMessages,
		Stream:   true, // Enable streaming
	}

	// Add tools if provided
	if options != nil && len(options.Tools) > 0 {
		reqBody.Tools = make([]ollamaTool, len(options.Tools))
		for i, tool := range options.Tools {
			reqBody.Tools[i] = ollamaTool{
				Type: tool.Type,
				Function: ollamaFunctionDef{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
	}

	// Add options
	if options != nil {
		reqBody.Options = make(map[string]interface{})
		if options.Temperature > 0 {
			reqBody.Options["temperature"] = options.Temperature
		}
		if options.TopP > 0 {
			reqBody.Options["top_p"] = options.TopP
		}
		if options.MaxTokens > 0 {
			reqBody.Options["num_predict"] = options.MaxTokens
		}
		if len(options.Stop) > 0 {
			reqBody.Options["stop"] = options.Stop
		}
	}

	// Marshal request
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read streaming response line by line
	decoder := json.NewDecoder(resp.Body)

	for {
		var ollamaResp ollamaResponse
		if err := decoder.Decode(&ollamaResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode response: %w", err)
		}

		// Create stream chunk
		chunk := types.StreamChunk{
			Content: ollamaResp.Message.Content,
			Done:    ollamaResp.Done,
		}

		// Add metadata on final chunk
		if ollamaResp.Done {
			chunk.Metadata = &types.Metadata{
				Model:            ollamaResp.Model,
				PromptTokens:     ollamaResp.PromptEvalCount,
				CompletionTokens: ollamaResp.EvalCount,
				TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
			}
		}

		// Convert tool calls if present
		if len(ollamaResp.Message.ToolCalls) > 0 {
			chunk.ToolCalls = make([]types.ToolCall, len(ollamaResp.Message.ToolCalls))
			for i, tc := range ollamaResp.Message.ToolCalls {
				chunk.ToolCalls[i] = types.ToolCall{
					ID:   fmt.Sprintf("call_%d", i),
					Type: "function",
					Function: types.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		// Call handler
		if err := handler(chunk); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}

		// Break if done
		if ollamaResp.Done {
			break
		}
	}

	return nil
}
