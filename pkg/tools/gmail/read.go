package gmail

import (
	"context"
	"fmt"
	"strings"

	gmailapi "google.golang.org/api/gmail/v1"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// ReadTool reads email messages by ID
type ReadTool struct {
	tools.BaseTool
	config     GmailConfig
	authHelper *AuthHelper
}

// NewReadTool creates a new Gmail read tool
func NewReadTool(config GmailConfig) *ReadTool {
	return &ReadTool{
		BaseTool: tools.NewBaseTool(
			"gmail_read",
			"Read a specific email message by its ID. Returns the email's subject, sender, date, body, and other metadata.",
			tools.CategoryEmail,
			true, // requires auth
			true, // safe operation (read-only)
		),
		config:     config,
		authHelper: NewAuthHelper(config),
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *ReadTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"message_id": {
				Type:        "string",
				Description: "The ID of the message to read (obtained from gmail_list or gmail_search)",
			},
			"format": {
				Type:        "string",
				Description: "Optional: Message format - 'full' (default, includes body), 'metadata' (headers only), or 'minimal' (ID and labels only)",
			},
		},
		Required: []string{"message_id"},
	}
}

// Execute reads an email message
func (t *ReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Validate credentials
	if err := t.authHelper.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("Gmail credentials not configured: %w", err)
	}

	// Extract parameters
	messageID, format, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Get Gmail service
	service, err := t.authHelper.GetService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Get message
	message, err := service.Users.Messages.Get("me", messageID).Format(format).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	// Parse message
	result := t.parseMessage(message, format)
	result["success"] = true
	result["message_id"] = messageID

	return result, nil
}

// extractParams extracts and validates parameters
func (t *ReadTool) extractParams(params map[string]interface{}) (string, string, error) {
	messageID, ok := params["message_id"].(string)
	if !ok || messageID == "" {
		return "", "", fmt.Errorf("message_id parameter is required and must be a non-empty string")
	}

	format := "full" // default
	if formatParam, ok := params["format"].(string); ok && formatParam != "" {
		format = strings.ToLower(formatParam)
		if format != "full" && format != "metadata" && format != "minimal" {
			return "", "", fmt.Errorf("invalid format: %s (must be 'full', 'metadata', or 'minimal')", formatParam)
		}
	}

	return messageID, format, nil
}

// parseMessage parses Gmail message into structured data
func (t *ReadTool) parseMessage(message *gmailapi.Message, format string) map[string]interface{} {
	result := map[string]interface{}{
		"id":        message.Id,
		"thread_id": message.ThreadId,
		"label_ids": message.LabelIds,
		"snippet":   message.Snippet,
	}

	if format == "minimal" {
		return result
	}

	// Parse headers
	headers := make(map[string]string)
	for _, header := range message.Payload.Headers {
		headers[header.Name] = header.Value
	}

	result["from"] = headers["From"]
	result["to"] = headers["To"]
	result["subject"] = headers["Subject"]
	result["date"] = headers["Date"]

	if cc := headers["Cc"]; cc != "" {
		result["cc"] = cc
	}
	if bcc := headers["Bcc"]; bcc != "" {
		result["bcc"] = bcc
	}

	// Get body (only in 'full' format)
	if format == "full" {
		body := t.extractBody(message.Payload)
		if body != "" {
			result["body"] = body
		}

		// Check for attachments
		if len(message.Payload.Parts) > 0 {
			attachments := t.extractAttachments(message.Payload.Parts)
			if len(attachments) > 0 {
				result["attachments"] = attachments
			}
		}
	}

	return result
}

// extractBody extracts email body from message payload
func (t *ReadTool) extractBody(payload *gmailapi.MessagePart) string {
	// If body data is directly in payload
	if payload.Body != nil && payload.Body.Data != "" {
		return payload.Body.Data
	}

	// Search in parts
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" || part.MimeType == "text/html" {
			if part.Body != nil && part.Body.Data != "" {
				return part.Body.Data
			}
		}

		// Recursive search in multipart messages
		if len(part.Parts) > 0 {
			if body := t.extractBody(part); body != "" {
				return body
			}
		}
	}

	return ""
}

// extractAttachments lists attachments from message parts
func (t *ReadTool) extractAttachments(parts []*gmailapi.MessagePart) []map[string]interface{} {
	var attachments []map[string]interface{}

	for _, part := range parts {
		if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
			attachments = append(attachments, map[string]interface{}{
				"filename":      part.Filename,
				"mime_type":     part.MimeType,
				"attachment_id": part.Body.AttachmentId,
				"size":          part.Body.Size,
			})
		}

		// Recursive search
		if len(part.Parts) > 0 {
			attachments = append(attachments, t.extractAttachments(part.Parts)...)
		}
	}

	return attachments
}
