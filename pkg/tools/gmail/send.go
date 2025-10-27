package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	gmailapi "google.golang.org/api/gmail/v1"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// SendTool sends emails via Gmail API
type SendTool struct {
	tools.BaseTool
	config     GmailConfig
	authHelper *AuthHelper
}

// NewSendTool creates a new Gmail send tool
func NewSendTool(config GmailConfig) *SendTool {
	return &SendTool{
		BaseTool: tools.NewBaseTool(
			"gmail_send",
			"Send an email via Gmail. Supports to, cc, bcc, subject, and body (plain text or HTML).",
			tools.CategoryEmail,
			true, // requires auth
			true, // safe operation (sending email is reversible via drafts)
		),
		config:     config,
		authHelper: NewAuthHelper(config),
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *SendTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"to": {
				Type:        "string",
				Description: "Recipient email address(es), comma-separated (e.g., 'user@example.com' or 'user1@example.com,user2@example.com')",
			},
			"subject": {
				Type:        "string",
				Description: "Email subject line",
			},
			"body": {
				Type:        "string",
				Description: "Email body content (plain text or HTML)",
			},
			"cc": {
				Type:        "string",
				Description: "Optional: CC email address(es), comma-separated",
			},
			"bcc": {
				Type:        "string",
				Description: "Optional: BCC email address(es), comma-separated",
			},
			"html": {
				Type:        "boolean",
				Description: "Optional: Set to true if body contains HTML (default: false for plain text)",
			},
		},
		Required: []string{"to", "subject", "body"},
	}
}

// Execute sends an email via Gmail
func (t *SendTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Validate credentials first
	if err := t.authHelper.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("Gmail credentials not configured: %w", err)
	}

	// Extract parameters
	to, subject, body, cc, bcc, isHTML, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Get Gmail service
	service, err := t.authHelper.GetService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Build email message
	message := t.buildMessage(to, cc, bcc, subject, body, isHTML)

	// Send email
	sent, err := service.Users.Messages.Send("me", &gmailapi.Message{
		Raw: message,
	}).Do()

	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"message_id": sent.Id,
		"thread_id":  sent.ThreadId,
		"to":         to,
		"subject":    subject,
	}, nil
}

// extractParams extracts and validates parameters
func (t *SendTool) extractParams(params map[string]interface{}) (string, string, string, string, string, bool, error) {
	to, ok := params["to"].(string)
	if !ok || to == "" {
		return "", "", "", "", "", false, fmt.Errorf("to parameter is required and must be a non-empty string")
	}

	subject, ok := params["subject"].(string)
	if !ok || subject == "" {
		return "", "", "", "", "", false, fmt.Errorf("subject parameter is required and must be a non-empty string")
	}

	body, ok := params["body"].(string)
	if !ok || body == "" {
		return "", "", "", "", "", false, fmt.Errorf("body parameter is required and must be a non-empty string")
	}

	cc := ""
	if ccParam, ok := params["cc"].(string); ok {
		cc = ccParam
	}

	bcc := ""
	if bccParam, ok := params["bcc"].(string); ok {
		bcc = bccParam
	}

	isHTML := false
	if htmlParam, ok := params["html"].(bool); ok {
		isHTML = htmlParam
	}

	return to, subject, body, cc, bcc, isHTML, nil
}

// buildMessage creates RFC 2822 compliant email message
func (t *SendTool) buildMessage(to, cc, bcc, subject, body string, isHTML bool) string {
	var message strings.Builder

	// Headers
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if cc != "" {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	if bcc != "" {
		message.WriteString(fmt.Sprintf("Bcc: %s\r\n", bcc))
	}
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	// Content-Type
	if isHTML {
		message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	} else {
		message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	}

	// Empty line between headers and body
	message.WriteString("\r\n")

	// Body
	message.WriteString(body)

	// Encode to base64url (required by Gmail API)
	return base64.URLEncoding.EncodeToString([]byte(message.String()))
}
