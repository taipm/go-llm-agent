package gmail

import (
	"context"
	"fmt"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// ListTool lists email messages with filters
type ListTool struct {
	tools.BaseTool
	config     GmailConfig
	authHelper *AuthHelper
}

// NewListTool creates a new Gmail list tool
func NewListTool(config GmailConfig) *ListTool {
	return &ListTool{
		BaseTool: tools.NewBaseTool(
			"gmail_list",
			"List email messages with optional filters. Returns message IDs, thread IDs, and snippets. Use gmail_read to get full message content.",
			tools.CategoryEmail,
			true, // requires auth
			true, // safe operation (read-only)
		),
		config:     config,
		authHelper: NewAuthHelper(config),
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *ListTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"query": {
				Type:        "string",
				Description: "Optional: Gmail search query (e.g., 'is:unread', 'from:user@example.com', 'subject:hello'). Leave empty for all messages.",
			},
			"max_results": {
				Type:        "integer",
				Description: "Optional: Maximum number of messages to return (default: 10, max: 500)",
			},
			"label_ids": {
				Type:        "array",
				Description: "Optional: Filter by label IDs (e.g., ['INBOX', 'UNREAD'])",
			},
			"include_spam_trash": {
				Type:        "boolean",
				Description: "Optional: Include messages from SPAM and TRASH (default: false)",
			},
		},
		Required: []string{},
	}
}

// Execute lists email messages
func (t *ListTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Validate credentials
	if err := t.authHelper.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("Gmail credentials not configured: %w", err)
	}

	// Extract parameters
	query, maxResults, labelIDs, includeSpamTrash, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Get Gmail service
	service, err := t.authHelper.GetService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Build list call
	listCall := service.Users.Messages.List("me")

	if query != "" {
		listCall = listCall.Q(query)
	}
	if maxResults > 0 {
		listCall = listCall.MaxResults(maxResults)
	}
	if len(labelIDs) > 0 {
		listCall = listCall.LabelIds(labelIDs...)
	}
	if includeSpamTrash {
		listCall = listCall.IncludeSpamTrash(true)
	}

	// Execute list request
	response, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	// Format results
	messages := make([]map[string]interface{}, 0, len(response.Messages))
	for _, msg := range response.Messages {
		messages = append(messages, map[string]interface{}{
			"id":        msg.Id,
			"thread_id": msg.ThreadId,
		})
	}

	return map[string]interface{}{
		"success":         true,
		"messages":        messages,
		"result_count":    len(messages),
		"total_estimate":  response.ResultSizeEstimate,
		"next_page_token": response.NextPageToken,
	}, nil
}

// extractParams extracts and validates parameters
func (t *ListTool) extractParams(params map[string]interface{}) (string, int64, []string, bool, error) {
	query := ""
	if queryParam, ok := params["query"].(string); ok {
		query = queryParam
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
	} else if maxResults > 500 {
		maxResults = 500
	}

	labelIDs := []string{}
	if labelParam, ok := params["label_ids"].([]interface{}); ok {
		for _, label := range labelParam {
			if labelStr, ok := label.(string); ok {
				labelIDs = append(labelIDs, labelStr)
			}
		}
	}

	includeSpamTrash := false
	if spamParam, ok := params["include_spam_trash"].(bool); ok {
		includeSpamTrash = spamParam
	}

	return query, maxResults, labelIDs, includeSpamTrash, nil
}
