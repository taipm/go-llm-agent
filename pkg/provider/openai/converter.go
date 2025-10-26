package openai

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// toOpenAIMessages converts agent messages to OpenAI format
func toOpenAIMessages(messages []types.Message) []openai.ChatCompletionMessageParamUnion {
	oaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			oaiMessages = append(oaiMessages, openai.SystemMessage(msg.Content))

		case "user":
			oaiMessages = append(oaiMessages, openai.UserMessage(msg.Content))

		case "assistant":
			// If no tool calls, simple message
			if len(msg.ToolCalls) == 0 {
				oaiMessages = append(oaiMessages, openai.AssistantMessage(msg.Content))
			} else {
				// With tool calls, need to build full param
				toolCalls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					// Marshal arguments to JSON string
					argsJSON, err := json.Marshal(tc.Function.Arguments)
					if err != nil {
						continue // Skip invalid tool calls
					}

					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      tc.Function.Name,
								Arguments: string(argsJSON),
							},
						},
					})
				}

				oaiMessages = append(oaiMessages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content: openai.ChatCompletionAssistantMessageParamContentUnion{
							OfString: param.NewOpt(msg.Content),
						},
						ToolCalls: toolCalls,
					},
				})
			}

		case "tool":
			oaiMessages = append(oaiMessages, openai.ToolMessage(msg.Content, msg.ToolID))
		}
	}

	return oaiMessages
}

// toOpenAITools converts our tool definitions to OpenAI format
func toOpenAITools(tools []types.ToolDefinition) []openai.ChatCompletionToolUnionParam {
	result := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))

	for _, tool := range tools {
		// Convert parameters to shared.FunctionParameters
		funcDef := shared.FunctionDefinitionParam{
			Name:        tool.Function.Name,
			Description: param.NewOpt(tool.Function.Description),
		}

		// If parameters are provided, convert them to map
		if tool.Function.Parameters != nil {
			paramsJSON, err := json.Marshal(tool.Function.Parameters)
			if err == nil {
				var paramsMap map[string]interface{}
				if err := json.Unmarshal(paramsJSON, &paramsMap); err == nil {
					funcDef.Parameters = shared.FunctionParameters(paramsMap)
				}
			}
		}

		result = append(result, openai.ChatCompletionFunctionTool(funcDef))
	}

	return result
}

// fromOpenAICompletion converts OpenAI completion to our Response type
func fromOpenAICompletion(completion *openai.ChatCompletion) (*types.Response, error) {
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in completion")
	}

	choice := completion.Choices[0]
	message := choice.Message

	response := &types.Response{
		Content: message.Content,
	}

	// Convert tool calls if present
	if len(message.ToolCalls) > 0 {
		response.ToolCalls = make([]types.ToolCall, 0, len(message.ToolCalls))
		for _, tc := range message.ToolCalls {
			// Get function tool call
			funcTC := tc.AsFunction()

			// Parse arguments JSON
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(funcTC.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
			}

			response.ToolCalls = append(response.ToolCalls, types.ToolCall{
				ID:   funcTC.ID,
				Type: "function",
				Function: types.FunctionCall{
					Name:      funcTC.Function.Name,
					Arguments: args,
				},
			})
		}
	}

	// Add metadata
	if completion.Usage.PromptTokens > 0 {
		response.Metadata = &types.Metadata{
			Model:            completion.Model,
			PromptTokens:     int(completion.Usage.PromptTokens),
			CompletionTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:      int(completion.Usage.TotalTokens),
		}
	}

	return response, nil
}
