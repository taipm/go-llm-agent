# Technical Specification - go-llm-agent

## 1. Project Goals

**go-llm-agent** is a lightweight, easy-to-use Go library for building intelligent AI agents with:
- Multi-provider LLM support (Ollama, OpenAI, Gemini)
- Tools/functions for specific tasks
- Context/memory management
- Multi-turn conversation handling

### Development Principles
- **LEAN**: Focus on essential features that work immediately
- **80/20**: Prioritize 20% of features that create 80% of value
- **Iterative**: Evolve through small versions
- **Simple API**: Easy to learn, use, and extend
- **Provider Agnostic**: Write once, run with any LLM provider

## 2. Overall Architecture

```
┌─────────────────────────────────────────────┐
│              Application Code               │
└─────────────────┬───────────────────────────┘
                  │
        ┌─────────▼──────────┐
        │  Factory Pattern   │
        │  provider.FromEnv()│
        │  provider.New()    │
        └─────────┬──────────┘
                  │
┌─────────────────▼───────────────────────────┐
│          Unified Provider Interface         │
│  Chat(), Stream(), GetModel(), etc.         │
└─────┬───────────┬───────────┬───────────────┘
      │           │           │
┌─────▼────┐ ┌───▼──────┐ ┌─▼────────┐
│  Ollama  │ │  OpenAI  │ │  Gemini  │
│ Provider │ │ Provider │ │ Provider │
└──────────┘ └──────────┘ └──────────┘
      │           │           │
┌─────▼───────────▼───────────▼───────────────┐
│         External LLM Services               │
│  Ollama API, OpenAI API, Gemini API         │
└─────────────────────────────────────────────┘
```

## 3. Core Components

### 3.1. Provider System (v0.2.0+)

**Purpose**: Abstract LLM communication across multiple providers

**Factory Pattern**:
- `provider.FromEnv()`: Auto-detect provider from environment variables
- `provider.New(name, config)`: Manual provider configuration

**Supported Providers**:

| Provider | Status | Features | Use Case |
|----------|--------|----------|----------|
| **Ollama** | ✅ Production | Local, Free, Privacy | Development, Learning |
| **OpenAI** | ✅ Production | GPT-4o, Tools, Streaming | Production, Best Quality |
| **Gemini** | ✅ Production | Large Context, Free Tier | Large Docs, Cost-Effective |

**Unified Interface**:
```go
type Provider interface {
    // Core methods
    Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error)
    Stream(ctx context.Context, messages []Message, options *ChatOptions, handler StreamHandler) error
    
    // Provider info
    GetModel() string
    GetProviderName() string
}
```

**Provider Configuration**:
```go
// Auto-detection from environment
provider, err := provider.FromEnv()

// Manual configuration
cfg := provider.Config{
    Provider: "openai",
    Model:    "gpt-4o-mini",
    APIKey:   "sk-...",
}
provider, err := provider.New("openai", cfg)
```

**Environment Variables**:

| Variable | Description | Example |
|----------|-------------|---------|
| `LLM_PROVIDER` | Provider name | `ollama`, `openai`, `gemini` |
| `LLM_MODEL` | Model name | `qwen3:1.7b`, `gpt-4o-mini`, `gemini-2.5-flash` |
| `OLLAMA_BASE_URL` | Ollama server URL | `http://localhost:11434` |
| `OPENAI_API_KEY` | OpenAI API key | `sk-...` |
| `GEMINI_API_KEY` | Gemini API key | `AI...` |

**Provider-Specific Behaviors**:

```go
// Ollama
- Local execution (no API key required)
- Supports custom models via `ollama pull`
- Tool calling: Limited model support (qwen3, llama3.1)
- Streaming: Full support
- Base URL: Configurable (default: http://localhost:11434)

// OpenAI
- Cloud-based (API key required)
- Models: gpt-4o, gpt-4o-mini, gpt-3.5-turbo
- Tool calling: Full support on all models
- Streaming: Full support
- Rate limits: Based on tier

// Gemini
- Cloud-based (API key required)
- Models: gemini-2.5-flash, gemini-2.0-pro
- Tool calling: Full support
- Streaming: Full support
- Context window: Up to 1M tokens
```

### 3.2. Agent (Legacy Pattern - Still Supported)
**Purpose**: Coordinate LLM, Tools, and Memory

**Core Functions**:
- Receive user input
- Send requests to LLM
- Handle tool calls from LLM response
- Manage conversation flow
- Store conversation history

**Interface**:
```go
type Agent interface {
    Chat(ctx context.Context, message string) (string, error)
    Run(ctx context.Context, task string) (*Result, error)
    AddTool(tool Tool) error
    Reset() error
}
```

**Note**: In v0.2.0+, direct provider usage is recommended over Agent pattern for simpler use cases.

### 3.3. LLM Provider (Implementation Details)

### 3.3. Tool System
**Purpose**: Enable agent to perform specific actions

**Functions**:
- Register tools with name and schema
- Validate input parameters
- Execute tool functions
- Return results to LLM

**Interface**:
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() *JSONSchema
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

type ToolRegistry struct {
    tools map[string]Tool
}
```

**Provider Tool Calling Support**:

| Provider | Tool Support | Notes |
|----------|--------------|-------|
| OpenAI | ✅ Full | All models support function calling |
| Gemini | ✅ Full | Native function calling support |
| Ollama | ⚠️ Limited | Model-dependent (qwen3, llama3.1+) |

### 3.4. Memory Manager
**Purpose**: Manage context and conversation history

**Current (v0.2.0) - Simple Memory**:
- In-memory conversation history
- Message buffer with limits
- Truncation strategies

**Future**:
- Vector database integration
- Semantic search
- Long-term memory

**Interface**:
```go
type Memory interface {
    Add(message Message) error
    GetHistory(limit int) ([]Message, error)
    Clear() error
    Search(query string, limit int) ([]Message, error) // Future
}
```

## 4. Data Models

### 4.1. Message
```go
type Message struct {
    Role      string                 `json:"role"`      // system, user, assistant, tool
    Content   string                 `json:"content"`
    ToolCalls []ToolCall            `json:"tool_calls,omitempty"`
    ToolID    string                 `json:"tool_call_id,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

### 4.2. Tool Call
```go
type ToolCall struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"` // function
    Function FunctionCall           `json:"function"`
}

type FunctionCall struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}
```

### 4.3. Response
```go
type Response struct {
    Content   string     `json:"content"`
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
    Metadata  *Metadata  `json:"metadata,omitempty"`
}

type Metadata struct {
    Model            string `json:"model"`
    PromptTokens     int    `json:"prompt_tokens"`
    CompletionTokens int    `json:"completion_tokens"`
    TotalTokens      int    `json:"total_tokens"`
}
```

## 5. Basic Workflow

### Execution Flow (Multi-Provider)
```
1. Application Code
   ↓
2. provider.FromEnv() or provider.New()
   ↓
3. Auto-detect provider (Ollama/OpenAI/Gemini)
   ↓
4. Create messages []types.Message
   ↓
5. Call provider.Chat(ctx, messages, options)
   ↓
6. Provider-specific API call
   ↓
7. Parse response (unified format)
   ↓
8. If tool calls:
   a. Execute tools
   b. Add results to messages
   c. Call provider.Chat() again
   ↓
9. Return final response to application
```

### Streaming Flow
```
1. Application provides StreamHandler function
   ↓
2. Call provider.Stream(ctx, messages, options, handler)
   ↓
3. Provider opens streaming connection
   ↓
4. For each token/chunk:
   a. Parse chunk
   b. Call handler(chunk)
   c. Handler prints/processes chunk
   ↓
5. Stream complete (chunk.Done = true)
```

## 6. Usage Examples

### 6.1. Simple Chat (Multi-Provider)

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
    // Auto-detect provider from environment
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    messages := []types.Message{
        {Role: types.RoleUser, Content: "What is the capital of France?"},
    }
    
    response, err := llm.Chat(ctx, messages, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Content) // "The capital of France is Paris."
}
```

### 6.2. Streaming Response

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
        {Role: types.RoleUser, Content: "Count from 1 to 10"},
    }
    
    handler := func(chunk types.StreamChunk) error {
        fmt.Print(chunk.Content)
        return nil
    }
    
    err = llm.Stream(ctx, messages, nil, handler)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 6.3. Tool Calling

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
    
    // Define tool
    tools := []types.ToolDefinition{
        {
            Type: "function",
            Function: types.FunctionDefinition{
                Name:        "get_weather",
                Description: "Get current weather",
                Parameters: &types.JSONSchema{
                    Type: "object",
                    Properties: map[string]*types.JSONSchema{
                        "location": {
                            Type: "string",
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
    
    if len(response.ToolCalls) > 0 {
        // Execute tool and return result
        fmt.Printf("Tool called: %s\n", response.ToolCalls[0].Function.Name)
    }
}
```

## 7. Technical Requirements

### Minimum Dependencies

- Go 1.25+
- Standard library
- HTTP client (net/http)

### Provider Dependencies

| Provider | Required Package | Version |
|----------|------------------|---------|
| Ollama | None (HTTP only) | - |
| OpenAI | `github.com/sashabaranov/go-openai` | v3.6.1+ |
| Gemini | `google.golang.org/genai` | v1.32.0+ |

### Testing

- Unit tests for each component
- Integration tests with all providers
- Compatibility test suite
- Example programs

## 8. Non-functional Requirements

### Performance

- Streaming response support (all providers)
- Timeout configuration
- Connection pooling
- Concurrent request handling

### Reliability

- Clear error handling
- Retry logic for network calls
- Graceful degradation
- Provider failover (future)

### Maintainability

- Clean code, well documented
- Examples for all features
- Semantic versioning
- Comprehensive test coverage (70%+)

## 9. Version Scope

### v0.1.0 (Released) ✅

- ✅ Basic agent with Ollama
- ✅ Simple tool system
- ✅ In-memory conversation history
- ✅ Clear, simple API
- ✅ Working examples

**Out of scope**:
- Multiple LLM providers
- Vector database integration
- Persistent storage
- Advanced memory strategies
- Streaming support

### v0.2.0 (Current - 60% Complete) 🔄

**Completed**:
- ✅ Multi-provider architecture (Ollama, OpenAI, Gemini)
- ✅ Factory pattern (FromEnv, New)
- ✅ Unified provider interface
- ✅ Streaming support (all providers)
- ✅ Tool calling (provider-dependent)
- ✅ Compatibility test suite
- ✅ Provider comparison documentation

**In Progress**:
- 🔄 Documentation update (README, QUICKSTART, SPEC)
- 🔄 Migration guide

**Remaining**:
- ⏸️ Release preparation
- ⏸️ Final testing
- ⏸️ Tag v0.2.0

### v0.3.0 (Planned)

- Agent builder pattern
- Persistent memory (SQLite, PostgreSQL)
- Vector database integration
- Multi-agent coordination
- Advanced streaming (function calling in stream)
- Cost tracking and monitoring

## 10. Success Metrics

### v0.1.0 Success Criteria ✅

- ✅ Chat with Ollama models
- ✅ Register and use at least 2 tools
- ✅ Maintain conversation context
- ✅ At least 3 working examples
- ✅ Full documentation
- ✅ Code coverage >= 70%

### v0.2.0 Success Criteria

- ✅ Support 3 providers (Ollama, OpenAI, Gemini)
- ✅ Same code works with all providers
- ✅ Streaming support
- ✅ Tool calling (where supported)
- ✅ Comprehensive provider comparison
- 🔄 Updated documentation
- ⏸️ Migration guide for v0.1 users
- ⏸️ Test coverage >= 71.8%

### Performance Targets

| Metric | Target | Actual |
|--------|--------|--------|
| Test Coverage | >= 70% | 71.8% ✅ |
| Compatibility Tests | 100% pass | 100% ✅ |
| Provider Support | 3 providers | 3 ✅ |
| API Breaking Changes | 0 | 0 ✅ |
| Documentation Completeness | 100% | 75% 🔄 |
