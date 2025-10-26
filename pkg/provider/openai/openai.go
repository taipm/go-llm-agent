package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// Provider implements the LLMProvider interface for OpenAI
type Provider struct {
	client openai.Client
	model  string
}

// New creates a new OpenAI provider
// apiKey: OpenAI API key (get from https://platform.openai.com/api-keys)
// model: Model name (e.g., "gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo")
func New(apiKey, model string) *Provider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Provider{
		client: client,
		model:  model,
	}
}

// NewWithBaseURL creates a new OpenAI provider with custom base URL
// Useful for Azure OpenAI or other OpenAI-compatible endpoints
func NewWithBaseURL(apiKey, model, baseURL string) *Provider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	return &Provider{
		client: client,
		model:  model,
	}
}

// Chat sends a message and returns a response
func (p *Provider) Chat(ctx context.Context, messages []types.Message, options *types.ChatOptions) (*types.Response, error) {
	// Convert messages to OpenAI format
	oaiMessages := toOpenAIMessages(messages)

	// Build request parameters
	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(p.model),
		Messages: oaiMessages,
	}

	// Apply options if provided
	if options != nil {
		if options.Temperature > 0 {
			params.Temperature = param.NewOpt(options.Temperature)
		}
		if options.MaxTokens > 0 {
			params.MaxCompletionTokens = param.NewOpt(int64(options.MaxTokens))
		}
		if options.TopP > 0 {
			params.TopP = param.NewOpt(options.TopP)
		}
		if len(options.Stop) > 0 {
			params.Stop = openai.ChatCompletionNewParamsStopUnion{
				OfStringArray: options.Stop,
			}
		}
		if len(options.Tools) > 0 {
			params.Tools = toOpenAITools(options.Tools)
		}
	}

	// Call OpenAI API
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, &ProviderError{
			Op:  "Chat",
			Err: err,
		}
	}

	// Convert response
	return fromOpenAICompletion(completion)
}

// Stream sends messages and streams the response
func (p *Provider) Stream(ctx context.Context, messages []types.Message, options *types.ChatOptions, handler types.StreamHandler) error {
	// Convert messages to OpenAI format
	oaiMessages := toOpenAIMessages(messages)

	// Build request parameters (same as Chat but with stream=true)
	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(p.model),
		Messages: oaiMessages,
	}

	// Apply options
	if options != nil {
		if options.Temperature > 0 {
			params.Temperature = param.NewOpt(options.Temperature)
		}
		if options.MaxTokens > 0 {
			params.MaxCompletionTokens = param.NewOpt(int64(options.MaxTokens))
		}
		if options.TopP > 0 {
			params.TopP = param.NewOpt(options.TopP)
		}
		if len(options.Stop) > 0 {
			params.Stop = openai.ChatCompletionNewParamsStopUnion{
				OfStringArray: options.Stop,
			}
		}
		if len(options.Tools) > 0 {
			params.Tools = toOpenAITools(options.Tools)
		}
	}

	// Start streaming
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	// Process stream chunks
	var currentToolCalls map[int]*types.ToolCall

	for stream.Next() {
		chunk := stream.Current()

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		delta := choice.Delta

		// Handle content
		if delta.Content != "" {
			if err := handler(types.StreamChunk{
				Content: delta.Content,
			}); err != nil {
				stream.Close()
				return err
			}
		}

		// Handle tool calls
		if len(delta.ToolCalls) > 0 {
			if currentToolCalls == nil {
				currentToolCalls = make(map[int]*types.ToolCall)
			}

			for _, tc := range delta.ToolCalls {
				idx := int(tc.Index)

				// Initialize or update tool call
				if currentToolCalls[idx] == nil {
					currentToolCalls[idx] = &types.ToolCall{
						ID:   tc.ID,
						Type: "function",
						Function: types.FunctionCall{
							Name:      "",
							Arguments: make(map[string]interface{}),
						},
					}
				}

				// Append function name and arguments
				if tc.Function.Name != "" {
					currentToolCalls[idx].Function.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					// Accumulate arguments (they come in chunks)
					existingArgs, _ := json.Marshal(currentToolCalls[idx].Function.Arguments)
					currentToolCalls[idx].Function.Arguments = appendJSONString(string(existingArgs), tc.Function.Arguments)
				}
			}
		}

		// Check if done
		if choice.FinishReason != "" {
			// Build final tool calls array
			var finalToolCalls []types.ToolCall
			if len(currentToolCalls) > 0 {
				finalToolCalls = make([]types.ToolCall, 0, len(currentToolCalls))
				for _, tc := range currentToolCalls {
					finalToolCalls = append(finalToolCalls, *tc)
				}
			}

			// Send done event
			if err := handler(types.StreamChunk{
				ToolCalls: finalToolCalls,
				Done:      true,
			}); err != nil {
				stream.Close()
				return err
			}
		}
	}

	if err := stream.Err(); err != nil {
		return &ProviderError{
			Op:  "Stream",
			Err: err,
		}
	}

	return stream.Close()
}

// appendJSONString is a helper to accumulate JSON string chunks
func appendJSONString(existing, new string) map[string]interface{} {
	combined := existing + new
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(combined), &result); err != nil {
		// If not valid JSON yet, return empty map
		return make(map[string]interface{})
	}
	return result
}

// ProviderError wraps OpenAI API errors
type ProviderError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("openai %s: %v", e.Op, e.Err)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}
