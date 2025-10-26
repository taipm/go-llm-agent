package gemini

import (
	"fmt"

	"github.com/taipm/go-llm-agent/pkg/types"
	"google.golang.org/genai"
)

// toGeminiContents converts our messages to Gemini Content format
// Returns contents and optional system instruction
func toGeminiContents(messages []types.Message) ([]*genai.Content, *genai.Content) {
	var contents []*genai.Content
	var systemInstruction *genai.Content

	for _, msg := range messages {
		switch msg.Role {
		case types.RoleSystem:
			// Gemini uses separate SystemInstruction field
			systemInstruction = genai.NewContentFromText(msg.Content, genai.RoleUser)

		case types.RoleUser:
			contents = append(contents, genai.NewContentFromText(msg.Content, genai.RoleUser))

		case types.RoleAssistant:
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				parts := make([]*genai.Part, 0)

				// Add text content if present
				if msg.Content != "" {
					parts = append(parts, genai.NewPartFromText(msg.Content))
				}

				// Add function calls
				for _, tc := range msg.ToolCalls {
					parts = append(parts, genai.NewPartFromFunctionCall(
						tc.Function.Name,
						tc.Function.Arguments,
					))
				}

				contents = append(contents, genai.NewContentFromParts(parts, genai.RoleModel))
			} else {
				// Simple assistant message
				contents = append(contents, genai.NewContentFromText(msg.Content, genai.RoleModel))
			}

		case types.RoleTool:
			// Tool result message
			// Parse the content as response data
			responseData := map[string]any{
				"result": msg.Content,
			}

			// Find the tool name from ToolID (simplified - in real scenario might need to track)
			toolName := msg.ToolID
			if toolName == "" {
				toolName = "tool_response"
			}

			contents = append(contents, genai.NewContentFromFunctionResponse(
				toolName,
				responseData,
				genai.RoleUser,
			))
		}
	}

	return contents, systemInstruction
}

// toGeminiTools converts our tool definitions to Gemini Tool format
func toGeminiTools(tools []types.ToolDefinition) []*genai.Tool {
	if len(tools) == 0 {
		return nil
	}

	functionDecls := make([]*genai.FunctionDeclaration, 0, len(tools))

	for _, tool := range tools {
		if tool.Type != "function" {
			continue
		}

		funcDecl := &genai.FunctionDeclaration{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
		}

		// Convert JSONSchema to Gemini Schema
		if tool.Function.Parameters != nil {
			funcDecl.Parameters = toGeminiSchema(tool.Function.Parameters)
		}

		functionDecls = append(functionDecls, funcDecl)
	}

	if len(functionDecls) == 0 {
		return nil
	}

	return []*genai.Tool{
		{
			FunctionDeclarations: functionDecls,
		},
	}
}

// toGeminiSchema converts our JSONSchema to Gemini Schema
func toGeminiSchema(jsonSchema *types.JSONSchema) *genai.Schema {
	if jsonSchema == nil {
		return nil
	}

	schema := &genai.Schema{
		Type:        genai.Type(jsonSchema.Type),
		Description: jsonSchema.Description,
	}

	// Convert properties
	if len(jsonSchema.Properties) > 0 {
		schema.Properties = make(map[string]*genai.Schema)
		for name, propSchema := range jsonSchema.Properties {
			schema.Properties[name] = toGeminiSchema(propSchema)
		}
	}

	// Convert items for arrays
	if jsonSchema.Items != nil {
		schema.Items = toGeminiSchema(jsonSchema.Items)
	}

	// Required fields
	if len(jsonSchema.Required) > 0 {
		schema.Required = jsonSchema.Required
	}

	// Enum values
	if len(jsonSchema.Enum) > 0 {
		enumStrings := make([]string, 0, len(jsonSchema.Enum))
		for _, e := range jsonSchema.Enum {
			if s, ok := e.(string); ok {
				enumStrings = append(enumStrings, s)
			}
		}
		schema.Enum = enumStrings
	}

	return schema
}

// fromGeminiResponse converts Gemini response to our Response format
func fromGeminiResponse(geminiResp *genai.GenerateContentResponse) (*types.Response, error) {
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := geminiResp.Candidates[0]
	if candidate.Content == nil {
		return nil, fmt.Errorf("no content in candidate")
	}

	response := &types.Response{
		Content: "",
	}

	// Process parts
	for _, part := range candidate.Content.Parts {
		// Text content
		if part.Text != "" {
			response.Content += part.Text
		}

		// Function calls (tool calls)
		if part.FunctionCall != nil {
			toolCall := types.ToolCall{
				ID: fmt.Sprintf("call_%s", part.FunctionCall.Name),
				Function: types.FunctionCall{
					Name:      part.FunctionCall.Name,
					Arguments: part.FunctionCall.Args,
				},
			}
			response.ToolCalls = append(response.ToolCalls, toolCall)
		}
	}

	// Extract metadata
	if geminiResp.UsageMetadata != nil {
		response.Metadata = &types.Metadata{
			PromptTokens:     int(geminiResp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(geminiResp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(geminiResp.UsageMetadata.TotalTokenCount),
		}
	}

	return response, nil
}
