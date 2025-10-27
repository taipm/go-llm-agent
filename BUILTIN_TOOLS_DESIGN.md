# Built-in Tools Design Document

## Overview

This document describes the design and implementation of built-in tools for go-llm-agent. Built-in tools provide ready-to-use functionality that agents can leverage without requiring custom implementation.

## Design Principles

1. **Modular**: Each tool category in its own package
2. **Consistent Interface**: All tools implement the same interface
3. **Provider-Agnostic**: Works with all LLM providers
4. **Safe by Default**: Input validation and error handling
5. **Well-Documented**: Clear descriptions and examples
6. **Testable**: Comprehensive unit tests for each tool

## Directory Structure

```
pkg/
├── tools/                      # Built-in tools package
│   ├── tools.go               # Tool interface and base types
│   ├── registry.go            # Tool registry for managing tools
│   ├── file/                  # File operations tools
│   │   ├── read.go           # Read file content
│   │   ├── write.go          # Write to file
│   │   ├── list.go           # List directory contents
│   │   ├── delete.go         # Delete files
│   │   └── file_test.go      # Tests
│   ├── web/                   # Web operations tools
│   │   ├── fetch.go          # HTTP GET request
│   │   ├── post.go           # HTTP POST request
│   │   ├── scrape.go         # Web scraping (basic)
│   │   └── web_test.go       # Tests
│   ├── system/                # System information tools
│   │   ├── execute.go        # Execute shell commands
│   │   ├── env.go            # Get environment variables
│   │   ├── info.go           # System information
│   │   └── system_test.go    # Tests
│   ├── data/                  # Data processing tools
│   │   ├── json.go           # JSON parsing/manipulation
│   │   ├── csv.go            # CSV processing
│   │   ├── xml.go            # XML processing
│   │   └── data_test.go      # Tests
│   ├── math/                  # Mathematical tools
│   │   ├── calculator.go     # Basic calculations
│   │   ├── statistics.go     # Statistical functions
│   │   └── math_test.go      # Tests
│   └── datetime/              # Date/time tools
│       ├── current.go        # Current time
│       ├── format.go         # Format conversion
│       ├── calc.go           # Date calculations
│       └── datetime_test.go  # Tests
│
examples/
└── builtin_tools/             # Examples using built-in tools
    ├── file_operations/
    │   └── main.go
    ├── web_scraper/
    │   └── main.go
    ├── data_processor/
    │   └── main.go
    └── complete_agent/        # Agent using multiple tools
        └── main.go
```

## Tool Interface

```go
package tools

import (
    "context"
    "github.com/taipm/go-llm-agent/pkg/types"
)

// Tool is the interface that all tools must implement
type Tool interface {
    // Name returns the unique name of the tool
    Name() string
    
    // Description returns a description of what the tool does
    Description() string
    
    // Parameters returns the JSON schema for the tool's parameters
    Parameters() *types.JSONSchema
    
    // Execute runs the tool with the given parameters
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
    
    // Category returns the tool category (file, web, system, etc.)
    Category() ToolCategory
    
    // RequiresAuth returns true if the tool requires authentication
    RequiresAuth() bool
}

// ToolCategory represents the category of a tool
type ToolCategory string

const (
    CategoryFile     ToolCategory = "file"
    CategoryWeb      ToolCategory = "web"
    CategorySystem   ToolCategory = "system"
    CategoryData     ToolCategory = "data"
    CategoryMath     ToolCategory = "math"
    CategoryDateTime ToolCategory = "datetime"
)

// BaseTool provides common functionality for all tools
type BaseTool struct {
    name         string
    description  string
    category     ToolCategory
    requiresAuth bool
}

func (b *BaseTool) Name() string          { return b.name }
func (b *BaseTool) Description() string   { return b.description }
func (b *BaseTool) Category() ToolCategory { return b.category }
func (b *BaseTool) RequiresAuth() bool    { return b.requiresAuth }
```

## Built-in Tools Catalog

### 1. File Operations (Priority: HIGH)

#### 1.1 Read File
- **Name**: `file_read`
- **Description**: Read content from a file
- **Parameters**: `path` (string), `encoding` (optional, default: utf-8)
- **Returns**: File content as string
- **Safety**: Validate path, check file exists, limit file size

#### 1.2 Write File
- **Name**: `file_write`
- **Description**: Write content to a file
- **Parameters**: `path` (string), `content` (string), `append` (bool, optional)
- **Returns**: Success confirmation with bytes written
- **Safety**: Validate path, create directories if needed, backup if exists

#### 1.3 List Directory
- **Name**: `file_list`
- **Description**: List files and directories in a path
- **Parameters**: `path` (string), `recursive` (bool, optional), `pattern` (optional)
- **Returns**: Array of file/directory information
- **Safety**: Validate path, limit depth for recursive

#### 1.4 Delete File
- **Name**: `file_delete`
- **Description**: Delete a file or directory
- **Parameters**: `path` (string), `recursive` (bool, for directories)
- **Returns**: Success confirmation
- **Safety**: Confirm before delete, no system files

### 2. Web Operations (Priority: HIGH)

#### 2.1 HTTP Fetch
- **Name**: `web_fetch`
- **Description**: Fetch content from a URL via HTTP GET
- **Parameters**: `url` (string), `headers` (map, optional), `timeout` (int, optional)
- **Returns**: Response body, status code, headers
- **Safety**: Validate URL, timeout limit, size limit

#### 2.2 HTTP Post
- **Name**: `web_post`
- **Description**: Send HTTP POST request
- **Parameters**: `url` (string), `body` (string/json), `headers` (map)
- **Returns**: Response body, status code
- **Safety**: Validate URL, timeout, size limits

#### 2.3 Web Scrape
- **Name**: `web_scrape`
- **Description**: Extract structured data from HTML
- **Parameters**: `url` (string), `selectors` (array of CSS selectors)
- **Returns**: Extracted data as JSON
- **Safety**: Rate limiting, respect robots.txt

### 3. System Operations (Priority: MEDIUM)

#### 3.1 Execute Command
- **Name**: `system_exec`
- **Description**: Execute a shell command
- **Parameters**: `command` (string), `args` (array), `timeout` (int)
- **Returns**: stdout, stderr, exit code
- **Safety**: Whitelist commands, timeout, no sudo

#### 3.2 Environment Variables
- **Name**: `system_env`
- **Description**: Get environment variable value
- **Parameters**: `key` (string)
- **Returns**: Value or empty string
- **Safety**: No sensitive vars (passwords, tokens)

#### 3.3 System Info
- **Name**: `system_info`
- **Description**: Get system information
- **Parameters**: `type` (cpu, memory, disk, os)
- **Returns**: System information as JSON
- **Safety**: Read-only operations

### 4. Data Processing (Priority: MEDIUM)

#### 4.1 JSON Parse
- **Name**: `data_json_parse`
- **Description**: Parse and query JSON data
- **Parameters**: `json` (string), `query` (JSONPath, optional)
- **Returns**: Parsed JSON or queried value
- **Safety**: Validate JSON, handle errors

#### 4.2 CSV Parse
- **Name**: `data_csv_parse`
- **Description**: Parse CSV data
- **Parameters**: `csv` (string), `delimiter` (optional), `headers` (bool)
- **Returns**: Array of objects
- **Safety**: Handle large files, validate format

#### 4.3 XML Parse
- **Name**: `data_xml_parse`
- **Description**: Parse XML data
- **Parameters**: `xml` (string), `xpath` (optional)
- **Returns**: Parsed XML or queried value
- **Safety**: Validate XML, handle errors

### 5. Math Operations (Priority: LOW)

#### 5.1 Calculator
- **Name**: `math_calculate`
- **Description**: Perform mathematical calculations
- **Parameters**: `expression` (string) or `operation` + `operands`
- **Returns**: Calculation result
- **Safety**: Validate expression, prevent injection

#### 5.2 Statistics
- **Name**: `math_stats`
- **Description**: Calculate statistics on dataset
- **Parameters**: `data` (array), `operation` (mean, median, mode, stddev)
- **Returns**: Statistical result
- **Safety**: Validate data type and size

### 6. Date/Time Operations (Priority: MEDIUM)

#### 6.1 Current Time
- **Name**: `datetime_now`
- **Description**: Get current date/time
- **Parameters**: `timezone` (optional), `format` (optional)
- **Returns**: Formatted datetime string
- **Safety**: Validate timezone and format

#### 6.2 Format Conversion
- **Name**: `datetime_format`
- **Description**: Convert datetime format
- **Parameters**: `datetime` (string), `from_format`, `to_format`
- **Returns**: Converted datetime string
- **Safety**: Validate formats

#### 6.3 Date Calculation
- **Name**: `datetime_calc`
- **Description**: Add/subtract time from date
- **Parameters**: `datetime` (string), `operation` (add/sub), `amount`, `unit`
- **Returns**: Calculated datetime
- **Safety**: Validate inputs

## Implementation Priority

### Phase 1 (Immediate - for v0.3.0)
1. **File Operations** (4 tools)
   - Essential for file-based agents
   - Low complexity, high value
   
2. **Web Operations** (3 tools)
   - Enable web-scraping agents
   - HTTP fetch most important

3. **Date/Time** (3 tools)
   - Commonly needed
   - Low complexity

### Phase 2 (Near-term)
4. **System Operations** (3 tools)
   - Shell execution needs careful security
   - Environment and info are safer
   
5. **Data Processing** (3 tools)
   - JSON most important
   - CSV for data agents

### Phase 3 (Future)
6. **Math Operations** (2 tools)
   - Nice to have
   - LLMs can often do this themselves

## Security Considerations

### 1. File Operations
- **Path Validation**: Prevent directory traversal (../)
- **Sandbox**: Limit operations to specific directories
- **Size Limits**: Prevent reading/writing huge files
- **Type Checking**: Validate file types for certain operations

### 2. Web Operations
- **URL Validation**: Prevent SSRF attacks
- **Rate Limiting**: Prevent abuse
- **Timeout**: Prevent hanging requests
- **Size Limits**: Limit response size
- **HTTPS Preferred**: Warn on HTTP

### 3. System Operations
- **Command Whitelist**: Only allow safe commands
- **No Privileged Access**: No sudo, no root
- **Timeout**: Prevent infinite loops
- **Output Limits**: Prevent memory exhaustion
- **Environment Filter**: Hide sensitive variables

### 4. General
- **Input Validation**: All parameters validated
- **Error Handling**: Graceful error messages
- **Logging**: Log all tool executions
- **Audit Trail**: Track what agents do

## Configuration

```go
// ToolConfig provides configuration for built-in tools
type ToolConfig struct {
    // File operations
    AllowedPaths      []string          // Whitelisted paths
    MaxFileSize       int64             // Maximum file size in bytes
    
    // Web operations
    HTTPTimeout       time.Duration     // HTTP request timeout
    MaxResponseSize   int64             // Maximum response size
    UserAgent         string            // Custom user agent
    
    // System operations
    AllowedCommands   []string          // Whitelisted commands
    CommandTimeout    time.Duration     // Command execution timeout
    
    // General
    EnableAuditLog    bool              // Enable audit logging
    AuditLogPath      string            // Path to audit log
    DangerousMode     bool              // Allow dangerous operations (default: false)
}
```

## Usage Examples

### Example 1: File Reading Agent

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/tools"
    "github.com/taipm/go-llm-agent/pkg/tools/file"
    "github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
    // Create provider
    llm, _ := provider.FromEnv()
    
    // Create file read tool
    fileRead := file.NewReadTool(file.Config{
        AllowedPaths: []string{"/tmp", "/home/user/documents"},
        MaxFileSize:  10 * 1024 * 1024, // 10MB
    })
    
    // Convert tool to LLM tool definition
    toolDef := tools.ToToolDefinition(fileRead)
    
    ctx := context.Background()
    messages := []types.Message{
        {
            Role:    types.RoleUser,
            Content: "Read the file /tmp/data.txt and summarize its content",
        },
    }
    
    // Chat with tools
    response, _ := llm.Chat(ctx, messages, &types.ChatOptions{
        Tools: []types.ToolDefinition{toolDef},
    })
    
    // Execute tool if requested
    if len(response.ToolCalls) > 0 {
        for _, tc := range response.ToolCalls {
            result, _ := fileRead.Execute(ctx, tc.Function.Arguments)
            fmt.Printf("File content: %v\n", result)
        }
    }
}
```

### Example 2: Web Scraper Agent

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/tools/web"
    "github.com/taipm/go-llm-agent/pkg/types"
)

func main() {
    llm, _ := provider.FromEnv()
    
    // Create web fetch tool
    webFetch := web.NewFetchTool(web.Config{
        Timeout:         30 * time.Second,
        MaxResponseSize: 1 * 1024 * 1024, // 1MB
        UserAgent:       "go-llm-agent/0.2.0",
    })
    
    ctx := context.Background()
    messages := []types.Message{
        {
            Role:    types.RoleUser,
            Content: "Fetch https://example.com and extract the title",
        },
    }
    
    response, _ := llm.Chat(ctx, messages, &types.ChatOptions{
        Tools: []types.ToolDefinition{tools.ToToolDefinition(webFetch)},
    })
    
    // Handle tool calls...
}
```

### Example 3: Multi-Tool Agent

```go
package main

import (
    "github.com/taipm/go-llm-agent/pkg/tools"
    "github.com/taipm/go-llm-agent/pkg/tools/file"
    "github.com/taipm/go-llm-agent/pkg/tools/web"
    "github.com/taipm/go-llm-agent/pkg/tools/datetime"
)

func main() {
    // Create tool registry
    registry := tools.NewRegistry()
    
    // Add multiple tools
    registry.Register(file.NewReadTool(file.DefaultConfig))
    registry.Register(file.NewWriteTool(file.DefaultConfig))
    registry.Register(web.NewFetchTool(web.DefaultConfig))
    registry.Register(datetime.NewNowTool())
    
    // Get all tools as LLM definitions
    toolDefs := registry.ToToolDefinitions()
    
    // Use with provider
    llm, _ := provider.FromEnv()
    response, _ := llm.Chat(ctx, messages, &types.ChatOptions{
        Tools: toolDefs,
    })
    
    // Execute requested tools
    if len(response.ToolCalls) > 0 {
        for _, tc := range response.ToolCalls {
            tool := registry.Get(tc.Function.Name)
            result, _ := tool.Execute(ctx, tc.Function.Arguments)
            fmt.Printf("Tool %s result: %v\n", tc.Function.Name, result)
        }
    }
}
```

## Testing Strategy

### Unit Tests
- Each tool has comprehensive unit tests
- Mock external dependencies (filesystem, HTTP)
- Test error conditions
- Test input validation

### Integration Tests
- Test tools with real providers
- Test tool chaining
- Test security restrictions
- Test with actual agent

### Security Tests
- Path traversal attempts
- Command injection attempts
- SSRF attempts
- Size limit violations

## Documentation Requirements

Each tool must have:
1. **Clear description**: What it does
2. **Parameter documentation**: What inputs it accepts
3. **Return value documentation**: What it returns
4. **Usage examples**: How to use it
5. **Security notes**: What to be careful about
6. **Error conditions**: What can go wrong

## Migration from Examples

Current example tools in `examples/tools/` should be:
1. Moved to `pkg/tools/` as built-in tools
2. Enhanced with security features
3. Properly tested
4. Documented

## Success Metrics

For v0.3.0 release, built-in tools are successful if:
- ✅ 10+ tools implemented across 3+ categories
- ✅ 100% test coverage for all tools
- ✅ Security validation for all inputs
- ✅ Complete documentation with examples
- ✅ Tool registry working with all providers
- ✅ 3+ complete agent examples using tools

## Future Enhancements

### v0.4.0 and beyond
- **Database Tools**: SQL query, NoSQL operations
- **AI/ML Tools**: Image processing, embedding generation
- **Communication Tools**: Email, Slack, Discord
- **Search Tools**: Google Search, DuckDuckGo
- **Code Tools**: Code execution, linting, formatting
- **Document Tools**: PDF parsing, DOCX processing
- **Vector Tools**: Vector database operations
- **Authentication**: OAuth, API key management

---

**Status**: Design Document  
**Version**: 1.0  
**Date**: October 27, 2025  
**Author**: AI Assistant  
**Next Steps**: Implementation in Sprint 4
