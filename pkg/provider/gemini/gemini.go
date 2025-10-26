package gemini

import (
	"context"
	"fmt"

	"github.com/taipm/go-llm-agent/pkg/types"
	"google.golang.org/genai"
)

// Provider implements the LLMProvider interface for Google Gemini
type Provider struct {
	client *genai.Client
	model  string
}

// New creates a new Gemini provider with API key
func New(ctx context.Context, apiKey, model string) (*Provider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Provider{
		client: client,
		model:  model,
	}, nil
}

// NewWithVertexAI creates a new Gemini provider for Vertex AI
func NewWithVertexAI(ctx context.Context, projectID, location, model string) (*Provider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
	}

	return &Provider{
		client: client,
		model:  model,
	}, nil
}

// Chat implements the LLMProvider interface
func (p *Provider) Chat(ctx context.Context, messages []types.Message, options *types.ChatOptions) (*types.Response, error) {
	// Convert messages to Gemini format
	contents, systemInstruction := toGeminiContents(messages)

	// Build config
	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInstruction,
	}

	// Apply options if provided
	if options != nil {
		if options.Temperature > 0 {
			temp := float32(options.Temperature)
			config.Temperature = &temp
		}
		if options.MaxTokens > 0 {
			config.MaxOutputTokens = int32(options.MaxTokens)
		}
		if len(options.Tools) > 0 {
			config.Tools = toGeminiTools(options.Tools)
		}
	}

	// Generate content
	response, err := p.client.Models.GenerateContent(ctx, p.model, contents, config)
	if err != nil {
		return nil, &ProviderError{
			Provider: "gemini",
			Message:  "failed to generate content",
			Original: err,
		}
	}

	// Convert response
	return fromGeminiResponse(response)
}

// Stream implements the LLMProvider interface for streaming responses
func (p *Provider) Stream(ctx context.Context, messages []types.Message, options *types.ChatOptions, handler types.StreamHandler) error {
	// Convert messages to Gemini format
	contents, systemInstruction := toGeminiContents(messages)

	// Build config
	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInstruction,
	}

	// Apply options if provided
	if options != nil {
		if options.Temperature > 0 {
			temp := float32(options.Temperature)
			config.Temperature = &temp
		}
		if options.MaxTokens > 0 {
			config.MaxOutputTokens = int32(options.MaxTokens)
		}
		if len(options.Tools) > 0 {
			config.Tools = toGeminiTools(options.Tools)
		}
	}

	// Stream content
	stream := p.client.Models.GenerateContentStream(ctx, p.model, contents, config)

	// Accumulate tool calls across chunks
	toolCallsMap := make(map[int]*types.ToolCall)
	var fullContent string

	for chunk, err := range stream {
		if err != nil {
			return &ProviderError{
				Provider: "gemini",
				Message:  "streaming error",
				Original: err,
			}
		}

		if len(chunk.Candidates) == 0 {
			continue
		}

		candidate := chunk.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			continue
		}

		// Process each part in the chunk
		for _, part := range candidate.Content.Parts {
			streamChunk := types.StreamChunk{}

			// Handle text content
			if part.Text != "" {
				streamChunk.Content = part.Text
				fullContent += part.Text
			}

			// Handle function calls (tool calls)
			if part.FunctionCall != nil {
				index := len(toolCallsMap)
				toolCall := types.ToolCall{
					ID: fmt.Sprintf("call_%d", index),
					Function: types.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: part.FunctionCall.Args,
					},
				}
				toolCallsMap[index] = &toolCall
				streamChunk.ToolCalls = []types.ToolCall{toolCall}
			}

			// Send chunk to handler
			if err := handler(streamChunk); err != nil {
				return err
			}
		}
	}

	// Send final chunk with done flag
	finalToolCalls := make([]types.ToolCall, 0, len(toolCallsMap))
	for i := 0; i < len(toolCallsMap); i++ {
		if tc, ok := toolCallsMap[i]; ok {
			finalToolCalls = append(finalToolCalls, *tc)
		}
	}

	return handler(types.StreamChunk{
		Done:      true,
		ToolCalls: finalToolCalls,
	})
}

// Close closes the Gemini client (no-op as genai.Client doesn't have Close)
func (p *Provider) Close() error {
	return nil
} // ProviderError wraps Gemini API errors
type ProviderError struct {
	Provider string
	Message  string
	Original error
}

func (e *ProviderError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s provider error: %s: %v", e.Provider, e.Message, e.Original)
	}
	return fmt.Sprintf("%s provider error: %s", e.Provider, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Original
}
