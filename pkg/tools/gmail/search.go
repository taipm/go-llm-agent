package gmail

import (
	"context"
	"fmt"
	"strings"

	gmailapi "google.golang.org/api/gmail/v1"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// SearchTool searches emails with advanced queries
type SearchTool struct {
	tools.BaseTool
	config     GmailConfig
	authHelper *AuthHelper
}

// NewSearchTool creates a new Gmail search tool
func NewSearchTool(config GmailConfig) *SearchTool {
	return &SearchTool{
		BaseTool: tools.NewBaseTool(
			"gmail_search",
			"Search emails using Gmail's advanced search syntax. Returns matching messages with metadata. Supports from, to, subject, has:attachment, is:unread, etc.",
			tools.CategoryEmail,
			true, // requires auth
			true, // safe operation (read-only)
		),
		config:     config,
		authHelper: NewAuthHelper(config),
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *SearchTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"query": {
				Type:        "string",
				Description: "Gmail search query. Examples: 'from:user@example.com', 'subject:meeting', 'is:unread', 'has:attachment', 'after:2024/01/01', 'newer_than:7d'",
			},
			"max_results": {
				Type:        "integer",
				Description: "Optional: Maximum number of results to return (default: 10, max: 100)",
			},
			"include_metadata": {
				Type:        "boolean",
				Description: "Optional: Include message metadata (from, to, subject, date) in results (default: true)",
			},
		},
		Required: []string{"query"},
	}
}

// Execute searches emails
func (t *SearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Validate credentials
	if err := t.authHelper.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("Gmail credentials not configured: %w", err)
	}

	// Extract parameters
	query, maxResults, includeMetadata, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Get Gmail service
	service, err := t.authHelper.GetService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Search messages
	listCall := service.Users.Messages.List("me").Q(query).MaxResults(maxResults)

	response, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Format results
	messages := make([]map[string]interface{}, 0, len(response.Messages))

	for _, msg := range response.Messages {
		messageData := map[string]interface{}{
			"id":        msg.Id,
			"thread_id": msg.ThreadId,
		}

		// Fetch metadata if requested
		if includeMetadata {
			fullMsg, err := service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
			if err == nil {
				metadata := t.extractMetadata(fullMsg)
				for k, v := range metadata {
					messageData[k] = v
				}
			}
		}

		messages = append(messages, messageData)
	}

	return map[string]interface{}{
		"success":        true,
		"query":          query,
		"messages":       messages,
		"result_count":   len(messages),
		"total_estimate": response.ResultSizeEstimate,
	}, nil
}

// extractParams extracts and validates parameters
func (t *SearchTool) extractParams(params map[string]interface{}) (string, int64, bool, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return "", 0, false, fmt.Errorf("query parameter is required and must be a non-empty string")
	}

	maxResults := int64(10) // default
	if maxParam, ok := params["max_results"].(float64); ok {
		maxResults = int64(maxParam)
	} else if maxParam, ok := params["max_results"].(int); ok {
		maxResults = int64(maxParam)
	} else if maxParam, ok := params["max_results"].(int64); ok {
		maxResults = maxParam
	}

	// Validate max results
	if maxResults < 1 {
		maxResults = 10
	} else if maxResults > 100 {
		maxResults = 100
	}

	includeMetadata := true // default
	if metaParam, ok := params["include_metadata"].(bool); ok {
		includeMetadata = metaParam
	}

	return query, maxResults, includeMetadata, nil
}

// extractMetadata extracts key metadata from message
func (t *SearchTool) extractMetadata(message *gmailapi.Message) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Parse headers
	headers := make(map[string]string)
	if message.Payload != nil {
		for _, header := range message.Payload.Headers {
			key := strings.ToLower(header.Name)
			headers[key] = header.Value
		}
	}

	if from := headers["from"]; from != "" {
		metadata["from"] = from
	}
	if to := headers["to"]; to != "" {
		metadata["to"] = to
	}
	if subject := headers["subject"]; subject != "" {
		metadata["subject"] = subject
	}
	if date := headers["date"]; date != "" {
		metadata["date"] = date
	}

	if message.Snippet != "" {
		metadata["snippet"] = message.Snippet
	}

	if len(message.LabelIds) > 0 {
		metadata["labels"] = message.LabelIds
	}

	return metadata
}
