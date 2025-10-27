# Migration Guide: v0.1.0 → v0.2.0

This guide helps you migrate your codebase from **v0.1.0** (Ollama-only) to **v0.2.0** (multi-provider support with Ollama, OpenAI, and Gemini).

## 📋 Table of Contents

- [What's New in v0.2.0](#whats-new-in-v020)
- [Breaking Changes](#breaking-changes)
- [Migration Steps](#migration-steps)
- [Code Examples](#code-examples)
- [Provider Selection Guide](#provider-selection-guide)
- [FAQ](#faq)

## 🎯 What's New in v0.2.0

### Major Features

1. **Multi-Provider Support**
   - Ollama (existing, enhanced)
   - OpenAI (GPT-4o, GPT-4o-mini, GPT-3.5-turbo)
   - Gemini (gemini-2.5-flash, gemini-2.0-pro)

2. **Factory Pattern**
   - `provider.FromEnv()` - Auto-detect from environment
   - `provider.New(name, config)` - Manual configuration
   - Provider-agnostic code

3. **Enhanced Features**
   - Streaming support (all providers)
   - Tool calling (OpenAI, Gemini, select Ollama models)
   - Unified API across providers
   - Environment-based configuration

4. **New Documentation**
   - PROVIDER_COMPARISON.md - Detailed provider comparison
   - Updated examples with all providers
   - Multi-provider best practices

## ⚠️ Breaking Changes

### Good News: **ZERO Breaking Changes!** 🎉

v0.2.0 is **100% backward compatible** with v0.1.0. Your existing code will continue to work without modifications.

**Why?**
- All v0.1.0 APIs are preserved
- Ollama provider works exactly the same way
- Agent pattern still fully supported
- No changes to data structures

**Migration Strategy:**
- **Immediate**: Your code works as-is
- **Recommended**: Adopt new patterns gradually
- **Future**: v0.1.0 patterns will be supported long-term

## 🚀 Migration Steps

### Step 1: Update Dependencies

```bash
# Update to v0.2.0
go get -u github.com/taipm/go-llm-agent@v0.2.0

# Install optional dependencies for environment variables
go get github.com/joho/godotenv
```

### Step 2: Choose Your Migration Path

You have **3 migration options**:

#### Option A: No Changes (Keep v0.1.0 Pattern) ✅

Your existing code works as-is:

```go
// v0.1.0 code - STILL WORKS IN v0.2.0
import (
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
    provider := ollama.New("http://localhost:11434", "llama3.2")
    ag := agent.New(provider)
    
    response, _ := ag.Chat(ctx, "Hello!")
    fmt.Println(response)
}
```

**When to use:** Stable production code, Ollama-only deployments, no immediate need for other providers.

#### Option B: Gradual Migration (Recommended) 🔄

Keep existing code but add new provider capability:

```go
// v0.2.0 pattern - provider-agnostic
import (
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/types"
    _ "github.com/joho/godotenv/autoload"
)

func main() {
    // Auto-detect from environment (new!)
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Direct provider usage (simpler than Agent)
    messages := []types.Message{
        {Role: types.RoleUser, Content: "Hello!"},
    }
    
    response, err := llm.Chat(ctx, messages, nil)
    fmt.Println(response.Content)
}
```

**When to use:** New features, multi-environment deployments, want flexibility.

#### Option C: Full Migration 🚀

Rewrite to take full advantage of v0.2.0 features:

1. Replace agent pattern with direct provider usage
2. Add .env configuration
3. Use factory pattern throughout
4. Add support for multiple providers

See [Code Examples](#code-examples) below for complete examples.

### Step 3: Add Environment Configuration (Optional but Recommended)

Create a `.env` file:

```bash
# Choose your provider
LLM_PROVIDER=ollama
LLM_MODEL=qwen3:1.7b

# Ollama configuration (if using Ollama)
OLLAMA_BASE_URL=http://localhost:11434

# OpenAI configuration (if using OpenAI)
# OPENAI_API_KEY=sk-...

# Gemini configuration (if using Gemini)
# GEMINI_API_KEY=...
```

### Step 4: Test Your Code

```bash
# Run existing tests (should still pass)
go test ./...

# Test with different providers
LLM_PROVIDER=ollama go run main.go
LLM_PROVIDER=openai go run main.go
LLM_PROVIDER=gemini go run main.go
```

## 💡 Code Examples

### Example 1: Basic Chat Migration

#### Before (v0.1.0)

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
    provider := ollama.New("http://localhost:11434", "llama3.2")
    ag := agent.New(provider)
    
    ctx := context.Background()
    response, err := ag.Chat(ctx, "What is 2+2?")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(response)
}
```

#### After (v0.2.0 - Recommended Pattern)

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/types"
    _ "github.com/joho/godotenv/autoload"
)

func main() {
    // Auto-detect provider from .env
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    messages := []types.Message{
        {Role: types.RoleUser, Content: "What is 2+2?"},
    }
    
    response, err := llm.Chat(ctx, messages, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Content)
}
```

**Benefits:**
- ✅ Works with Ollama, OpenAI, or Gemini (same code!)
- ✅ Environment-based configuration
- ✅ No hardcoded URLs or models
- ✅ Easier testing with different providers

### Example 2: Tool Calling Migration

#### Before (v0.1.0)

```go
// Tool calling was limited to Ollama with Agent pattern
import (
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

type WeatherTool struct{}

func (w *WeatherTool) Name() string { return "get_weather" }
func (w *WeatherTool) Description() string { return "Get weather" }
// ... implement Tool interface

func main() {
    provider := ollama.New("http://localhost:11434", "llama3.2")
    ag := agent.New(provider)
    ag.AddTool(&WeatherTool{})
    
    response, _ := ag.Chat(ctx, "Weather in Tokyo?")
    fmt.Println(response)
}
```

#### After (v0.2.0 - Direct Tool Calling)

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/types"
    _ "github.com/joho/godotenv/autoload"
)

func main() {
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Define tool using standard types
    tools := []types.ToolDefinition{
        {
            Type: "function",
            Function: types.FunctionDefinition{
                Name:        "get_weather",
                Description: "Get current weather for a location",
                Parameters: &types.JSONSchema{
                    Type: "object",
                    Properties: map[string]*types.JSONSchema{
                        "location": {
                            Type:        "string",
                            Description: "City name",
                        },
                    },
                    Required: []string{"location"},
                },
            },
        },
    }
    
    ctx := context.Background()
    messages := []types.Message{
        {Role: types.RoleUser, Content: "What's the weather in Tokyo?"},
    }
    
    options := &types.ChatOptions{Tools: tools}
    response, err := llm.Chat(ctx, messages, options)
    if err != nil {
        log.Fatal(err)
    }
    
    // Handle tool calls
    if len(response.ToolCalls) > 0 {
        for _, tc := range response.ToolCalls {
            fmt.Printf("Tool: %s\n", tc.Function.Name)
            fmt.Printf("Args: %v\n", tc.Function.Arguments)
            
            // Execute your tool here
            result := executeWeatherTool(tc.Function.Arguments)
            
            // Return result to LLM
            messages = append(messages, types.Message{
                Role:      types.RoleAssistant,
                ToolCalls: response.ToolCalls,
            })
            messages = append(messages, types.Message{
                Role:    types.RoleTool,
                Content: result,
                ToolID:  tc.ID,
            })
            
            // Get final response
            finalResponse, _ := llm.Chat(ctx, messages, nil)
            fmt.Println(finalResponse.Content)
        }
    }
}

func executeWeatherTool(args map[string]interface{}) string {
    location := args["location"].(string)
    return fmt.Sprintf(`{"location": "%s", "temp": 22, "condition": "Sunny"}`, location)
}
```

**Benefits:**
- ✅ Works with OpenAI, Gemini, and select Ollama models
- ✅ Standard tool definition format
- ✅ More control over tool execution
- ✅ Better error handling

### Example 3: Streaming Migration

#### Before (v0.1.0)

```go
// Streaming not supported in v0.1.0
// Had to use blocking Chat() calls
```

#### After (v0.2.0 - Streaming Support)

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/types"
    _ "github.com/joho/godotenv/autoload"
)

func main() {
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    messages := []types.Message{
        {Role: types.RoleUser, Content: "Tell me a story"},
    }
    
    // Stream handler
    handler := func(chunk types.StreamChunk) error {
        fmt.Print(chunk.Content)
        
        if chunk.Done {
            fmt.Println("\n[Stream complete]")
        }
        
        return nil
    }
    
    // Stream the response
    err = llm.Stream(ctx, messages, nil, handler)
    if err != nil {
        log.Fatal(err)
    }
}
```

**Benefits:**
- ✅ Real-time output
- ✅ Better UX for long responses
- ✅ Works with all providers

### Example 4: Multi-Provider Support

**New in v0.2.0** - Same code, multiple providers:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/types"
)

func chatWithProvider(providerName string) {
    // Configure via environment
    os.Setenv("LLM_PROVIDER", providerName)
    
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatalf("Failed to initialize %s: %v", providerName, err)
    }
    
    ctx := context.Background()
    messages := []types.Message{
        {Role: types.RoleUser, Content: "What is AI?"},
    }
    
    response, err := llm.Chat(ctx, messages, nil)
    if err != nil {
        log.Fatalf("Chat failed with %s: %v", providerName, err)
    }
    
    fmt.Printf("[%s] %s\n\n", providerName, response.Content)
}

func main() {
    // Test with all providers
    providers := []string{"ollama", "openai", "gemini"}
    
    for _, p := range providers {
        chatWithProvider(p)
    }
}
```

## 🎯 Provider Selection Guide

### When to Use Each Provider

| Scenario | Recommended Provider | Why |
|----------|---------------------|-----|
| **Development/Testing** | Ollama | Free, local, private |
| **Production (Best Quality)** | OpenAI | Industry-leading quality |
| **Large Documents** | Gemini | 1M+ token context window |
| **Cost-Sensitive** | Ollama or Gemini | Free tier available |
| **Privacy-Critical** | Ollama | 100% local execution |
| **Tool Calling** | OpenAI or Gemini | Reliable tool support |
| **Rapid Prototyping** | Ollama | No API keys, instant setup |

### Migration Decision Tree

```
Do you need cloud-based AI?
├─ No → Use Ollama (local, free, private)
└─ Yes
   ├─ Need best quality? → Use OpenAI
   ├─ Need large context? → Use Gemini (1M tokens)
   ├─ Have budget constraints? → Use Gemini (free tier)
   └─ Need reliability? → Use OpenAI (production-proven)
```

## 📚 Best Practices

### 1. Use Environment Variables

**Don't hardcode:**
```go
// ❌ Bad - hardcoded provider
provider := ollama.New("http://localhost:11434", "llama3.2")
```

**Do configure via environment:**
```go
// ✅ Good - environment-based
llm, err := provider.FromEnv()
```

**Benefits:**
- Easy provider switching
- Better security (no API keys in code)
- Environment-specific configuration
- Easier testing

### 2. Handle Provider-Specific Features

```go
// Check if provider supports tool calling
if llm.GetProviderName() == "ollama" {
    fmt.Println("Note: Tool calling limited on Ollama")
}

// Use tools with all providers
response, err := llm.Chat(ctx, messages, &types.ChatOptions{
    Tools: tools,
})

if len(response.ToolCalls) == 0 && len(tools) > 0 {
    log.Println("Warning: Tools not supported or model doesn't support tools")
}
```

### 3. Graceful Fallbacks

```go
func getProvider() (provider.Provider, error) {
    // Try FromEnv first
    llm, err := provider.FromEnv()
    if err != nil {
        // Fallback to Ollama
        log.Println("Using Ollama fallback")
        cfg := provider.Config{
            Provider: "ollama",
            Model:    "qwen3:1.7b",
            BaseURL:  "http://localhost:11434",
        }
        return provider.New("ollama", cfg)
    }
    return llm, nil
}
```

### 4. Test with Multiple Providers

```go
func TestChatAllProviders(t *testing.T) {
    providers := []string{"ollama", "openai", "gemini"}
    
    for _, p := range providers {
        t.Run(p, func(t *testing.T) {
            os.Setenv("LLM_PROVIDER", p)
            llm, err := provider.FromEnv()
            if err != nil {
                t.Skip("Provider not configured")
            }
            
            // Your test logic
        })
    }
}
```

## ❓ FAQ

### Q: Do I need to update my v0.1.0 code immediately?

**A:** No! v0.2.0 is 100% backward compatible. Your existing code will continue to work.

### Q: Should I migrate to the new pattern?

**A:** Recommended but not required. The new pattern offers:
- Multi-provider support
- Better testing (can switch providers)
- More flexible configuration
- Future-proof architecture

**Migrate when:** You start a new feature, need a new provider, or have time for refactoring.

### Q: Can I use both patterns in the same project?

**A:** Yes! You can mix v0.1.0 Agent pattern with v0.2.0 direct provider usage:

```go
// v0.1.0 pattern (still works)
ollamaProvider := ollama.New("http://localhost:11434", "llama3.2")
ag := agent.New(ollamaProvider)

// v0.2.0 pattern (new code)
llm, _ := provider.FromEnv()
response, _ := llm.Chat(ctx, messages, nil)
```

### Q: What if I only use Ollama?

**A:** You can continue using Ollama exactly as before. v0.2.0 adds options but doesn't force changes.

### Q: How do I switch providers without code changes?

**A:** Just change your `.env` file:

```bash
# Switch from Ollama to OpenAI
LLM_PROVIDER=openai  # was: ollama
LLM_MODEL=gpt-4o-mini  # was: qwen3:1.7b
OPENAI_API_KEY=sk-...  # add this

# Run same code - works with OpenAI now!
go run main.go
```

### Q: Are there any performance differences?

**A:** Yes, but minimal for the library itself:
- **Ollama**: Local, fastest for small models, no network latency
- **OpenAI**: Cloud, fast with global CDN
- **Gemini**: Cloud, very fast with optimized infrastructure

The provider choice impacts **model speed**, not library overhead.

### Q: What about tool calling support?

**A:** Tool calling works best with:
- **OpenAI**: ✅ All models support it
- **Gemini**: ✅ All models support it
- **Ollama**: ⚠️ Model-dependent (qwen3:1.7b, llama3.1+)

Use `PROVIDER_COMPARISON.md` for detailed tool support.

### Q: How do I handle API keys securely?

**A:** Use environment variables and `.env` files:

```bash
# .env file (add to .gitignore!)
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=...
```

```go
// Load .env automatically
import _ "github.com/joho/godotenv/autoload"

// Keys loaded from environment
llm, err := provider.FromEnv()
```

**Never commit API keys to Git!**

### Q: Can I customize provider behavior?

**A:** Yes, use `provider.New()` with custom config:

```go
cfg := provider.Config{
    Provider:    "ollama",
    Model:       "custom-model",
    BaseURL:     "http://custom-server:11434",
    Temperature: 0.7,
    MaxTokens:   2000,
}

llm, err := provider.New("ollama", cfg)
```

### Q: What's next after v0.2.0?

**A:** See roadmap in README.md. v0.3.0 will include:
- Agent builder pattern
- Persistent memory (SQLite, PostgreSQL)
- Vector database integration
- Multi-agent coordination

## 📞 Support

- **Documentation**: [README.md](README.md), [PROVIDER_COMPARISON.md](PROVIDER_COMPARISON.md)
- **Examples**: [examples/](examples/)
- **Issues**: [GitHub Issues](https://github.com/taipm/go-llm-agent/issues)

---

**Ready to migrate?** Start with Option B (gradual migration) and gradually adopt v0.2.0 patterns! 🚀
