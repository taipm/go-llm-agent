# Built-in Tools Package

The `builtin` package provides a convenient way to use all built-in tools in go-llm-agent without manually registering each one.

## Overview

This package offers:
- **10 Production-Ready Tools**: File operations (4), Web requests (3), DateTime utilities (3)
- **Sensible Defaults**: Pre-configured with security and safety in mind
- **Easy Customization**: Override configurations as needed
- **Zero Boilerplate**: Get started with just one line of code

## Quick Start

### Simplest Usage

```go
import "github.com/taipm/go-llm-agent/pkg/builtin"

// Get registry with all 10 built-in tools
registry := builtin.GetRegistry()

// Use tools
result, err := registry.Execute(ctx, "file_read", map[string]interface{}{
    "path": "example.txt",
})
```

That's it! The registry is pre-configured with:
- ✅ File tools with 10MB size limit, backup enabled
- ✅ Web tools with 30s timeout, SSRF protection
- ✅ DateTime tools with timezone support

### Custom Configuration

```go
// Start with default config and customize
config := builtin.DefaultConfig()

// Customize file tools
config.File.Base.AllowedPaths = []string{"/custom/path"}
config.File.Base.MaxFileSize = 5 * 1024 * 1024 // 5MB

// Customize web tools
config.Web.Fetch.UserAgent = "MyApp/1.0"
config.Web.Fetch.Timeout = 60 * time.Second

// Skip certain categories
config.NoWeb = true // Don't register web tools

// Create registry with custom config
registry := builtin.GetRegistryWithConfig(config)
```

### Category-Specific Tools

```go
// Get only file tools
fileTools := builtin.GetFileTools(nil)

// Get only web tools
webTools := builtin.GetWebTools(nil)

// Get only datetime tools
datetimeTools := builtin.GetDateTimeTools()

// Get tools by category from registry
registry := builtin.GetRegistry()
fileTools := registry.ByCategory(tools.CategoryFile)
```

## Available Tools

### File Tools (4)

| Tool | Description | Safe |
|------|-------------|------|
| `file_read` | Read file contents | ✅ |
| `file_list` | List directory contents | ✅ |
| `file_write` | Write to file with backup | ❌ |
| `file_delete` | Delete file with protection | ❌ |

**Default Configuration:**
- Allowed paths: current directory, `/tmp`, system temp
- Max file size: 10MB
- Backup enabled for write operations
- Protected system paths for delete

### Web Tools (3)

| Tool | Description | Safe |
|------|-------------|------|
| `web_fetch` | HTTP GET requests | ✅ |
| `web_post` | HTTP POST with JSON/form data | ✅ |
| `web_scrape` | HTML scraping with CSS selectors | ✅ |

**Default Configuration:**
- Timeout: 30 seconds
- Max response: 1MB (5MB for scraping)
- SSRF protection enabled
- Private IPs blocked
- Rate limiting: 1 second between requests (scrape only)

### DateTime Tools (3)

| Tool | Description | Safe |
|------|-------------|------|
| `datetime_now` | Get current time in any timezone | ✅ |
| `datetime_format` | Convert between datetime formats | ✅ |
| `datetime_calc` | Add/subtract/diff datetime | ✅ |

**Features:**
- RFC3339, RFC1123, Unix timestamps, custom formats
- Timezone conversion (IANA timezones)
- Duration arithmetic with day support

## API Reference

### Main Functions

#### `GetRegistry() *tools.Registry`
Returns a registry with all 10 built-in tools using default configuration.

```go
registry := builtin.GetRegistry()
```

#### `GetRegistryWithConfig(config Config) *tools.Registry`
Returns a registry with custom configuration.

```go
config := builtin.DefaultConfig()
config.NoWeb = true
registry := builtin.GetRegistryWithConfig(config)
```

#### `DefaultConfig() Config`
Returns the default configuration for all tools.

```go
config := builtin.DefaultConfig()
// Modify as needed
config.File.Base.MaxFileSize = 20 * 1024 * 1024 // 20MB
```

### Helper Functions

#### `GetAllTools() []tools.Tool`
Returns all 10 tools as a slice.

```go
allTools := builtin.GetAllTools()
for _, tool := range allTools {
    fmt.Println(tool.Name())
}
```

#### `GetToolsByCategory(category tools.ToolCategory) []tools.Tool`
Returns tools in a specific category.

```go
fileTools := builtin.GetToolsByCategory(tools.CategoryFile)
webTools := builtin.GetToolsByCategory(tools.CategoryWeb)
datetimeTools := builtin.GetToolsByCategory(tools.CategoryDateTime)
```

#### `GetFileTools(config *FileConfig) []tools.Tool`
Returns file tools with optional custom config.

```go
fileTools := builtin.GetFileTools(nil) // Use defaults
```

#### `GetWebTools(config *WebConfig) []tools.Tool`
Returns web tools with optional custom config.

```go
webTools := builtin.GetWebTools(nil) // Use defaults
```

#### `GetDateTimeTools() []tools.Tool`
Returns datetime tools (no configuration needed).

```go
datetimeTools := builtin.GetDateTimeTools()
```

#### `ToolCount() int`
Returns the total number of built-in tools (always 10).

```go
count := builtin.ToolCount() // Returns 10
```

## Configuration Structure

### Config

```go
type Config struct {
    File   FileConfig
    Web    WebConfig
    NoFile bool // Skip file tools
    NoWeb  bool // Skip web tools
    NoTime bool // Skip datetime tools
}
```

### FileConfig

```go
type FileConfig struct {
    Base   file.Config       // Config for read/list
    Write  file.WriteConfig  // Config for write
    Delete file.DeleteConfig // Config for delete
}
```

### WebConfig

```go
type WebConfig struct {
    Fetch  web.Config       // Config for fetch
    Post   web.PostConfig   // Config for post
    Scrape web.ScrapeConfig // Config for scrape
}
```

## Usage Examples

### With LLM Provider

```go
import (
    "github.com/taipm/go-llm-agent/pkg/builtin"
    "github.com/taipm/go-llm-agent/pkg/provider"
)

// Initialize provider
llm, _ := provider.FromEnv()

// Get registry with all tools
registry := builtin.GetRegistry()

// Convert to tool definitions for LLM
toolDefs := registry.ToToolDefinitions()

// Chat with tools
response, _ := llm.Chat(ctx, messages, &types.ChatOptions{
    Tools: toolDefs,
})

// Execute tool calls
for _, tc := range response.ToolCalls {
    result, _ := registry.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
    // Handle result...
}
```

### Selective Registration

```go
// Only file and datetime tools
config := builtin.DefaultConfig()
config.NoWeb = true
registry := builtin.GetRegistryWithConfig(config)

// Or manually select
registry := tools.NewRegistry()
for _, tool := range builtin.GetFileTools(nil) {
    registry.Register(tool)
}
for _, tool := range builtin.GetDateTimeTools() {
    registry.Register(tool)
}
```

### Security-Focused Configuration

```go
config := builtin.DefaultConfig()

// Strict file access
config.File.Base.AllowedPaths = []string{"/app/data"}
config.File.Base.AllowSymlinks = false
config.File.Base.MaxFileSize = 1024 * 1024 // 1MB only

// Disable unsafe operations
config.File.Write.CreateDirs = false
config.File.Delete.RequireConfirmation = true

// Strict web access
config.Web.Fetch.AllowPrivateIPs = false
config.Web.Fetch.Timeout = 10 * time.Second
config.Web.Scrape.RateLimit = 5 * time.Second

registry := builtin.GetRegistryWithConfig(config)
```

### Production Best Practices

```go
import (
    "os"
    "time"
    "github.com/taipm/go-llm-agent/pkg/builtin"
)

func createProductionRegistry() *tools.Registry {
    config := builtin.DefaultConfig()
    
    // File tools: restrict to app directories
    config.File.Base.AllowedPaths = []string{
        os.Getenv("APP_DATA_DIR"),
        os.TempDir(),
    }
    config.File.Base.MaxFileSize = 50 * 1024 * 1024 // 50MB
    config.File.Write.Backup = true
    config.File.Delete.RequireConfirmation = true
    
    // Web tools: production timeouts
    config.Web.Fetch.Timeout = 30 * time.Second
    config.Web.Fetch.UserAgent = "MyApp/" + os.Getenv("APP_VERSION")
    config.Web.Fetch.AllowPrivateIPs = false
    
    return builtin.GetRegistryWithConfig(config)
}
```

## Testing

The builtin package includes comprehensive tests:

```bash
go test ./pkg/builtin/... -v
```

Tests cover:
- Default configuration
- Registry creation
- Selective tool registration (NoFile, NoWeb, NoTime)
- Category filtering
- Tool counts and validation
- Custom configurations

## Migration Guide

### Before (Manual Registration)

```go
registry := tools.NewRegistry()

// File tools
fileConfig := file.Config{...}
registry.Register(file.NewReadTool(fileConfig))
registry.Register(file.NewListTool(fileConfig))
// ... 40+ lines of configuration

// Web tools
webConfig := web.Config{...}
registry.Register(web.NewFetchTool(webConfig))
// ... more configuration

// DateTime tools
registry.Register(datetime.NewNowTool())
// ... more registration
```

### After (Builtin Package)

```go
// Use defaults (one line!)
registry := builtin.GetRegistry()

// Or customize
config := builtin.DefaultConfig()
config.File.Base.AllowedPaths = []string{"/custom/path"}
registry := builtin.GetRegistryWithConfig(config)
```

## See Also

- [Examples](../../examples/builtin_tools/) - Complete working examples
- [Tools Documentation](../tools/) - Individual tool details
- [API Documentation](../../README.md) - Main library documentation
