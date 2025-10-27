# Gmail Tools

Gmail tools provide email automation capabilities through the Gmail API. These tools allow AI agents to send, read, list, and search emails.

## Features

- **gmail_send**: Send emails with support for to, cc, bcc, HTML content
- **gmail_read**: Read email messages by ID with full content extraction
- **gmail_list**: List emails with filters and pagination
- **gmail_search**: Advanced email search using Gmail query syntax

## Authentication Setup

Gmail tools require OAuth2 authentication. Follow these steps to set up:

### 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Gmail API:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Gmail API"
   - Click "Enable"

### 2. Create OAuth2 Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth client ID"
3. If prompted, configure the OAuth consent screen:
   - User Type: External (for personal use) or Internal (for workspace)
   - Fill in app name, user support email, and developer contact
   - Add scopes: `https://www.googleapis.com/auth/gmail.modify`
4. Select "Desktop app" as application type
5. Download the credentials JSON file
6. Save it as `credentials.json` in your working directory

### 3. Configure the Tool

```go
import (
    "github.com/taipm/go-llm-agent/pkg/builtin"
    "github.com/taipm/go-llm-agent/pkg/tools/gmail"
)

// Option 1: Use default configuration (credentials.json and token.json in current directory)
registry := builtin.GetRegistryWithConfig(builtin.Config{
    NoGmail: false, // Enable Gmail tools
})

// Option 2: Custom credentials path
registry := builtin.GetRegistryWithConfig(builtin.Config{
    NoGmail: false,
    Gmail: builtin.GmailConfig{
        Config: gmail.GmailConfig{
            CredentialsFile: "/path/to/credentials.json",
            TokenFile:       "/path/to/token.json",
            Scopes: []string{gmail.GmailModifyScope},
        },
    },
})

// Option 3: Get Gmail tools separately
gmailTools := builtin.GetGmailTools()
for _, tool := range gmailTools {
    registry.Register(tool)
}
```

### 4. First-Time Authorization

On first use, the tool will:
1. Open a browser for Google account authorization
2. Request permission to access Gmail
3. Save the access token to `token.json` for future use

```go
// The authorization happens automatically on first tool execution
result, err := tool.Execute(ctx, params)
```

## Tool Usage Examples

### Send Email

```go
params := map[string]interface{}{
    "to":      "recipient@example.com",
    "subject": "Hello from AI Agent",
    "body":    "This is an automated email",
    "html":    true, // Optional: send as HTML
    "cc":      "cc@example.com", // Optional
    "bcc":     "bcc@example.com", // Optional
}
result, err := sendTool.Execute(ctx, params)
// Returns: {"success": true, "message_id": "...", "thread_id": "..."}
```

### Read Email

```go
params := map[string]interface{}{
    "message_id": "18d1e2f3c4b5a6d7",
    "format":     "full", // Options: full, metadata, minimal
}
result, err := readTool.Execute(ctx, params)
// Returns full message with headers, body, and attachments
```

### List Emails

```go
params := map[string]interface{}{
    "query":       "is:unread", // Optional: Gmail search query
    "max_results": 10,           // Optional: default 10, max 500
    "label_ids":   []string{"INBOX", "UNREAD"}, // Optional
}
result, err := listTool.Execute(ctx, params)
// Returns array of message IDs and snippets
```

### Search Emails

```go
params := map[string]interface{}{
    "query":            "from:boss@company.com subject:urgent",
    "max_results":      20,
    "include_metadata": true, // Optional: include from, to, subject, date
}
result, err := searchTool.Execute(ctx, params)
// Returns matching messages with metadata
```

## Gmail Search Syntax

The `gmail_list` and `gmail_search` tools support Gmail's powerful search operators:

- `from:sender@example.com` - Emails from specific sender
- `to:recipient@example.com` - Emails to specific recipient
- `subject:keyword` - Emails with keyword in subject
- `is:unread` - Unread emails
- `is:read` - Read emails
- `is:starred` - Starred emails
- `has:attachment` - Emails with attachments
- `after:2024/01/01` - Emails after date
- `before:2024/12/31` - Emails before date
- `newer_than:7d` - Emails newer than 7 days
- `older_than:1m` - Emails older than 1 month
- `label:work` - Emails with specific label

Combine operators with AND (space) or OR:
- `from:boss@company.com is:unread` - Unread emails from boss
- `from:alice OR from:bob` - Emails from Alice or Bob
- `subject:(report OR summary) after:2024/01/01` - Reports or summaries after Jan 1

## Security Considerations

1. **Credentials Storage**: Keep `credentials.json` and `token.json` secure
2. **Token Expiration**: Tokens are automatically refreshed
3. **Scope Limitation**: Default scope is `gmail.modify` (read and send only)
4. **Rate Limits**: Gmail API has usage limits (check Google Cloud Console)
5. **.gitignore**: Add `credentials.json` and `token.json` to `.gitignore`

## Troubleshooting

### "credentials.json not found"

Make sure the credentials file exists and the path is correct:

```go
config := gmail.GmailConfig{
    CredentialsFile: "/absolute/path/to/credentials.json",
}
```

### "Unable to get OAuth2 token"

1. Delete `token.json` to force re-authorization
2. Check that Gmail API is enabled in Google Cloud Console
3. Verify OAuth consent screen is configured

### "Access blocked: This app's request is invalid"

1. Check OAuth consent screen configuration
2. Add your email to test users (if app is in testing mode)
3. Verify redirect URIs in OAuth client settings

### "Insufficient permissions"

Ensure the OAuth2 scope includes Gmail access:

```go
Scopes: []string{gmailapi.GmailModifyScope}
```

## API Reference

### gmail_send

**Parameters:**
- `to` (string, required): Recipient email address
- `subject` (string, required): Email subject
- `body` (string, required): Email body
- `cc` (string, optional): CC recipients (comma-separated)
- `bcc` (string, optional): BCC recipients (comma-separated)
- `html` (boolean, optional): Send as HTML (default: false)

**Returns:**
- `success`: boolean
- `message_id`: string
- `thread_id`: string

### gmail_read

**Parameters:**
- `message_id` (string, required): Gmail message ID
- `format` (string, optional): Response format - "full" (default), "metadata", or "minimal"

**Returns:**
- `id`: message ID
- `thread_id`: thread ID
- `from`, `to`, `subject`, `date`: email headers
- `body`: email content (full format only)
- `attachments`: array of attachment metadata (full format only)
- `label_ids`: message labels
- `snippet`: message preview

### gmail_list

**Parameters:**
- `query` (string, optional): Gmail search query
- `max_results` (integer, optional): Maximum messages to return (default: 10, max: 500)
- `label_ids` (array, optional): Filter by label IDs
- `include_spam_trash` (boolean, optional): Include SPAM and TRASH (default: false)

**Returns:**
- `messages`: array of message objects with id, thread_id, snippet
- `result_size_estimate`: approximate total matching messages
- `next_page_token`: token for pagination (if more results exist)

### gmail_search

**Parameters:**
- `query` (string, required): Gmail search query
- `max_results` (integer, optional): Maximum messages to return (default: 10, max: 100)
- `include_metadata` (boolean, optional): Include message metadata (default: true)

**Returns:**
- Array of message objects with id, thread_id, snippet, and metadata (if enabled)

## Dependencies

- `google.golang.org/api/gmail/v1` - Official Gmail API client
- `golang.org/x/oauth2` - OAuth2 authentication

## License

Same as parent project.
