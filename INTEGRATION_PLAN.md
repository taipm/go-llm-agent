# Phương Án Tích Hợp OpenAI & Google Gemini

## 📊 Phân Tích So Sánh

### 1. OpenAI Go SDK (`github.com/openai/openai-go`)

**Ưu điểm:**
- ✅ SDK chính thức, bảo trì tốt (2.6k stars)
- ✅ API design rất clean với functional options pattern
- ✅ Hỗ trợ đầy đủ: Chat, Streaming, Function calling, Azure OpenAI
- ✅ Error handling tốt với typed errors
- ✅ Auto-retry, pagination, middleware support
- ✅ Webhooks verification

**Nhược điểm:**
- ⚠️ API khá verbose (nhiều struct nested)
- ⚠️ Phụ thuộc Go 1.22+ (omitzero semantics)
- ⚠️ Complex type system (unions, param.Opt[T])

**Ví dụ API:**
```go
client := openai.NewClient(option.WithAPIKey("..."))
completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Hello"),
    },
    Model: shared.ChatModelGPT4o,
})
```

---

### 2. Google Gen AI Go SDK (`google.golang.org/genai`)

**Ưu điểm:**
- ✅ SDK chính thức từ Google (825 stars)
- ✅ API đơn giản, dễ dùng
- ✅ Hỗ trợ cả Gemini API & Vertex AI
- ✅ Multimodal Live support (video/audio streaming)
- ✅ Function calling, caching, batches
- ✅ Lightweight, ít boilerplate

**Nhược điểm:**
- ⚠️ Ít mature hơn OpenAI SDK
- ⚠️ Documentation chưa đầy đủ
- ⚠️ Streaming API khác biệt

**Ví dụ API:**
```go
client, _ := genai.NewClient(ctx, &genai.ClientConfig{
    APIKey: apiKey,
    Backend: genai.BackendGeminiAPI,
})
result, _ := client.Models.GenerateContent(ctx, "gemini-2.0-flash", 
    []*genai.Content{{Parts: parts}}, nil)
```

---

## 🎯 Chiến Lược Tích Hợp

### Nguyên Tắc Thiết Kế

1. **Unified API First** - Tất cả providers dùng CHUNG 1 interface, khác biệt chỉ ở constructor
2. **Zero breaking changes** - Backward compatible hoàn toàn với v0.1.x
3. **Provider pattern** - Mỗi LLM là một provider implementation của `types.LLMProvider`
4. **Abstraction over vendor-specific features** - Ẩn complexity của từng SDK phía sau
5. **Ease of Use** - Switch provider = đổi 1 dòng khởi tạo, code còn lại giữ nguyên

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    User Application                         │
│  Code giống hệt nhau, CHỈ đổi provider constructor          │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│          go-llm-agent Core (UNIFIED API)                    │
│  types.LLMProvider interface (Chat, Stream methods)         │
│  Agent.Chat(), Agent.ChatStream() - SAME for ALL providers  │
└───────────────────────────┬─────────────────────────────────┘
                            │
              ┌─────────────┼─────────────┐
              │             │             │
         ┌────▼────┐   ┌────▼────┐   ┌───▼─────┐
         │ Ollama  │   │ OpenAI  │   │ Gemini  │
         │ Provider│   │ Provider│   │ Provider│
         │ (v0.1)  │   │ (NEW)   │   │ (NEW)   │
         └────┬────┘   └────┬────┘   └───┬─────┘
              │             │            │
         ┌────▼────┐   ┌────▼────┐   ┌──▼──────┐
         │  HTTP   │   │ openai- │   │go-genai │
         │ Client  │   │ go v3.6.1│   │v1.32.0  │
         └─────────┘   └─────────┘   └─────────┘
                       (PINNED)      (PINNED)

Key: Tất cả providers implement types.LLMProvider → Code dùng Agent KHÔNG biết provider nào
```

---

## 📦 Implementation Plan

### Phase 1: OpenAI Provider (Ưu tiên cao)

**Lý do ưu tiên:**
- OpenAI phổ biến nhất (GPT-4, GPT-4o)
- SDK mature, well-documented
- Production-ready

**Tasks:**

#### 1.1. Create OpenAI Provider Structure
```go
// pkg/provider/openai/openai.go
package openai

import (
    "context"
    "github.com/openai/openai-go/v3"
    "github.com/openai/openai-go/v3/option"
    "github.com/taipm/go-llm-agent/pkg/types"
)

type Provider struct {
    client *openai.Client
    model  string
}

func New(apiKey, model string) *Provider {
    client := openai.NewClient(option.WithAPIKey(apiKey))
    return &Provider{
        client: client,
        model:  model,
    }
}

func (p *Provider) Chat(ctx context.Context, messages []types.Message, options *types.ChatOptions) (*types.Response, error) {
    // Convert types.Message -> openai.ChatCompletionMessageParamUnion
    // Call client.Chat.Completions.New()
    // Convert openai.ChatCompletion -> types.Response
}

func (p *Provider) Stream(ctx context.Context, messages []types.Message, options *types.ChatOptions, handler types.StreamHandler) error {
    // Stream with client.Chat.Completions.NewStreaming()
    // Convert chunks and call handler
}
```

#### 1.2. Message Conversion Layer
```go
// Converter từ types.Message sang OpenAI format
func toOpenAIMessages(msgs []types.Message) []openai.ChatCompletionMessageParamUnion {
    result := make([]openai.ChatCompletionMessageParamUnion, len(msgs))
    for i, msg := range msgs {
        switch msg.Role {
        case types.RoleUser:
            result[i] = openai.UserMessage(msg.Content)
        case types.RoleAssistant:
            result[i] = openai.AssistantMessage(msg.Content)
        case types.RoleSystem:
            result[i] = openai.SystemMessage(msg.Content)
        }
        
        // Handle tool calls if present
        if len(msg.ToolCalls) > 0 {
            // Convert tool calls
        }
    }
    return result
}
```

#### 1.3. Tool Calling Support
```go
// Convert types.ToolDefinition -> openai.ChatCompletionToolParam
func toOpenAITools(tools []types.ToolDefinition) []openai.ChatCompletionToolParam {
    result := make([]openai.ChatCompletionToolParam, len(tools))
    for i, tool := range tools {
        result[i] = openai.ChatCompletionToolParam{
            Type: openai.ChatCompletionToolTypeFunction,
            Function: openai.FunctionDefinitionParam{
                Name:        openai.String(tool.Function.Name),
                Description: openai.String(tool.Function.Description),
                Parameters:  tool.Function.Parameters,
            },
        }
    }
    return result
}
```

**Estimated Effort:** 2-3 days

**Files to create:**
- `pkg/provider/openai/openai.go` (~300 lines)
- `pkg/provider/openai/converter.go` (~200 lines)
- `pkg/provider/openai/openai_test.go` (~400 lines)
- `examples/openai_chat/main.go` (~100 lines)

**Critical: API Uniformity Checklist**
- [ ] Constructor signature: `New(apiKey, model string) *Provider`
- [ ] Implements `types.LLMProvider` interface exactly
- [ ] `Chat(ctx, messages, options) (*Response, error)` - same signature
- [ ] `Stream(ctx, messages, options, handler) error` - same signature
- [ ] Error handling returns `types.ProviderError` for consistency
- [ ] Example code structure matches `examples/simple_chat/main.go`

---

### Phase 2: Google Gemini Provider

**Lý do thứ hai:**
- Gemini 2.0 Flash rất nhanh & miễn phí
- Multimodal capabilities (text, image, video, audio)
- Vertex AI support cho enterprise

**Tasks:**

#### 2.1. Gemini Provider Structure
```go
// pkg/provider/gemini/gemini.go
package gemini

import (
    "context"
    "google.golang.org/genai"
    "github.com/taipm/go-llm-agent/pkg/types"
)

type Provider struct {
    client *genai.Client
    model  string
}

func New(apiKey, model string) (*Provider, error) {
    client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
        APIKey:  apiKey,
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        return nil, err
    }
    
    return &Provider{
        client: client,
        model:  model,
    }, nil
}

func (p *Provider) Chat(ctx context.Context, messages []types.Message, options *types.ChatOptions) (*types.Response, error) {
    // Convert to genai.Content
    // Call client.Models.GenerateContent()
    // Convert response
}
```

#### 2.2. Multimodal Support Extension
```go
// Extend types.Message để support multimodal
type Part struct {
    Text      string          `json:"text,omitempty"`
    InlineData *Blob          `json:"inline_data,omitempty"`
}

type Blob struct {
    MIMEType string `json:"mime_type"`
    Data     []byte `json:"data"`
}

// User có thể gửi:
msg := types.Message{
    Role: types.RoleUser,
    Parts: []types.Part{
        {Text: "What's in this image?"},
        {InlineData: &types.Blob{
            MIMEType: "image/jpeg",
            Data: imageBytes,
        }},
    },
}
```

**Estimated Effort:** 2-3 days

**Files to create:**
- `pkg/provider/gemini/gemini.go` (~250 lines)
- `pkg/provider/gemini/converter.go` (~150 lines)
- `pkg/provider/gemini/gemini_test.go` (~300 lines)
- `examples/gemini_chat/main.go` (~100 lines)
- `examples/gemini_multimodal/main.go` (~150 lines)

**Critical: API Uniformity Checklist**
- [ ] Constructor signature: `New(apiKey, model string) (*Provider, error)`
- [ ] Implements `types.LLMProvider` interface exactly
- [ ] Chat/Stream signatures match OpenAI and Ollama providers
- [ ] Basic text chat works identically to other providers
- [ ] Multimodal = optional extension, doesn't break base API
- [ ] Example `gemini_chat` mirrors `openai_chat` and `simple_chat`

---

### Phase 3: Unified Configuration & Factory

#### 3.1. Provider Factory Pattern
```go
// pkg/provider/factory.go
package provider

import (
    "fmt"
    "github.com/taipm/go-llm-agent/pkg/types"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
    "github.com/taipm/go-llm-agent/pkg/provider/openai"
    "github.com/taipm/go-llm-agent/pkg/provider/gemini"
)

type ProviderType string

const (
    ProviderOllama  ProviderType = "ollama"
    ProviderOpenAI  ProviderType = "openai"
    ProviderGemini  ProviderType = "gemini"
)

type Config struct {
    Type     ProviderType
    APIKey   string
    BaseURL  string
    Model    string
}

func New(config Config) (types.LLMProvider, error) {
    switch config.Type {
    case ProviderOllama:
        return ollama.New(config.BaseURL, config.Model), nil
    case ProviderOpenAI:
        return openai.New(config.APIKey, config.Model), nil
    case ProviderGemini:
        return gemini.New(config.APIKey, config.Model)
    default:
        return nil, fmt.Errorf("unknown provider type: %s", config.Type)
    }
}
```

#### 3.2. Easy Setup cho User
```go
// examples/multi_provider/main.go
func main() {
    // Cách 1: Direct provider creation
    ollamaProvider := ollama.New("http://localhost:11434", "llama3.2")
    agent1 := agent.New(ollamaProvider)
    
    // Cách 2: Factory pattern
    provider, _ := provider.New(provider.Config{
        Type:   provider.ProviderOpenAI,
        APIKey: os.Getenv("OPENAI_API_KEY"),
        Model:  "gpt-4o",
    })
    agent2 := agent.New(provider)
    
    // Cách 3: From environment variables
    provider, _ := provider.FromEnv() // Auto-detect từ env vars
    agent3 := agent.New(provider)
}
```

**Estimated Effort:** 1 day
**Files to create:**
- `pkg/provider/factory.go` (~150 lines)
- `pkg/provider/factory_test.go` (~100 lines)
- `examples/multi_provider/main.go` (~200 lines)

---

## 🔧 Technical Decisions

### 1. Dependencies Management

**Strategy: Pinned Versions for Stability**

```go
// go.mod
module github.com/taipm/go-llm-agent

go 1.22 // Required by both openai-go and go-genai

require (
    github.com/openai/openai-go/v3 v3.6.1    // Pinned: Latest stable (Nov 2024)
    google.golang.org/genai v1.32.0          // Pinned: Latest stable (Oct 2025)
)
```

**Why Pinned Versions:**
- ✅ **Stability**: Avoid breaking changes from SDK updates
- ✅ **Reproducible builds**: Same behavior across environments
- ✅ **Controlled upgrades**: Update when we validate compatibility
- ✅ **Security**: Known versions, easier to audit

**Version Update Policy:**
1. Review SDK changelogs quarterly
2. Test in dev branch first
3. Update pin after validation
4. Document changes in CHANGELOG.md

**Trade-offs:**
- Users get all SDKs (larger package) → OK for production use
- Single go.mod → easier to maintain
- Consistent experience → prioritizes ease of use

---

### 2. API Compatibility Matrix

| Feature | Ollama | OpenAI | Gemini | Support Strategy |
|---------|--------|--------|--------|------------------|
| Chat | ✅ | ✅ | ✅ | Core feature |
| Streaming | ✅ | ✅ | ✅ | Core feature |
| Function calling | ✅ | ✅ | ✅ | Core feature |
| System prompts | ✅ | ✅ | ✅ | Core feature |
| Temperature | ✅ | ✅ | ✅ | Core feature |
| Max tokens | ✅ | ✅ | ✅ | Core feature |
| Vision (images) | ⚠️ | ✅ | ✅ | Extended feature |
| Audio | ❌ | ✅ | ✅ | Extended feature |
| Video | ❌ | ❌ | ✅ | Extended feature |
| JSON mode | ⚠️ | ✅ | ✅ | Extended feature |
| Caching | ❌ | ❌ | ✅ | Provider-specific |

**Strategy:**
- Core features: Implement cho tất cả providers
- Extended features: Optional, graceful degradation
- Provider-specific: Document clearly, có thể access raw client

---

### 3. Error Handling Strategy

```go
// pkg/types/errors.go
package types

type ProviderError struct {
    Provider   string
    StatusCode int
    Message    string
    Original   error
}

func (e *ProviderError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Provider, e.Message)
}

// Usage
_, err := provider.Chat(ctx, messages, nil)
if err != nil {
    var provErr *types.ProviderError
    if errors.As(err, &provErr) {
        if provErr.StatusCode == 429 {
            // Rate limit
        }
    }
}
```

---

## 📝 Documentation Updates

### README.md additions:

```markdown
## 🌐 Supported Providers

- **Ollama** - Run models locally (llama3, qwen, etc.)
- **OpenAI** - GPT-4, GPT-4o, GPT-3.5
- **Google Gemini** - Gemini 2.0 Flash, Pro, Ultra

### Quick Start with Different Providers

#### Ollama (Local)
\`\`\`go
provider := ollama.New("http://localhost:11434", "llama3.2")
agent := agent.New(provider)
\`\`\`

#### OpenAI
\`\`\`go
provider := openai.New(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
agent := agent.New(provider)
\`\`\`

#### Google Gemini
\`\`\`go
provider, _ := gemini.New(os.Getenv("GEMINI_API_KEY"), "gemini-2.0-flash")
agent := agent.New(provider)
\`\`\`

### Provider Comparison

| Provider | Speed | Cost | Local | Multimodal |
|----------|-------|------|-------|------------|
| Ollama   | Fast  | Free | ✅    | ⚠️         |
| OpenAI   | Fast  | $$   | ❌    | ✅         |
| Gemini   | Fastest| $   | ❌    | ✅         |
```

---

## 📅 Implementation Timeline

### Sprint 1: OpenAI Provider (Week 1)
- [x] Research & design (1 day) - **Current**
- [ ] OpenAI provider implementation (2 days)
- [ ] Message/tool conversion layer (1 day)
- [ ] Tests & examples (1 day)
- [ ] Documentation (0.5 day)

### Sprint 2: Gemini Provider (Week 2)
- [ ] Gemini provider implementation (2 days)
- [ ] Multimodal support (1 day)
- [ ] Tests & examples (1 day)
- [ ] Documentation (0.5 day)

### Sprint 3: Integration & Polish (Week 3)
- [ ] Factory pattern (1 day)
- [ ] Unified configuration (1 day)
- [ ] Comprehensive tests (1 day)
- [ ] Examples for all providers (1 day)
- [ ] Final documentation (1 day)

**Total: 3 weeks to complete**

---

## 🎯 Success Metrics

1. **Backward Compatibility**: 100% - existing code works unchanged
2. **Test Coverage**: Maintain >70% for all providers
3. **Documentation**: Clear examples for each provider
4. **Performance**: No degradation from direct SDK usage
5. **User Experience**: Simple 3-line provider switch

---

## 🚀 Example Usage (Final State) - API Đồng Nhất

### Cách 1: Direct Provider (Rõ ràng, dễ debug)

```go
package main

import (
    "context"
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
    "github.com/taipm/go-llm-agent/pkg/provider/openai"
    "github.com/taipm/go-llm-agent/pkg/provider/gemini"
)

func main() {
    ctx := context.Background()
    
    // CHỈ cần đổi 1 DÒNG này, code còn lại GIỐNG HỆT NHAU!
    
    // Option 1: Ollama (local, free)
    provider := ollama.New("http://localhost:11434", "llama3.2")
    
    // Option 2: OpenAI (GPT-4, cloud)
    // provider := openai.New(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
    
    // Option 3: Gemini (fastest, cheapest)
    // provider, _ := gemini.New(os.Getenv("GEMINI_API_KEY"), "gemini-2.0-flash")
    
    // --- Code dưới đây HOÀN TOÀN GIỐNG NHAU cho tất cả providers ---
    ag := agent.New(provider)
    
    response, err := ag.Chat(ctx, "What is the capital of France?")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(response) // "Paris" - regardless of provider!
}
```

### Cách 2: Factory Pattern (Linh hoạt, config-driven)

```go
func main() {
    ctx := context.Background()
    
    // Load từ env/config
    // PROVIDER=openai OPENAI_API_KEY=xxx go run .
    // PROVIDER=gemini GEMINI_API_KEY=xxx go run .
    // PROVIDER=ollama OLLAMA_BASE_URL=xxx go run .
    
    provider, err := provider.FromEnv()
    if err != nil {
        panic(err)
    }
    
    ag := agent.New(provider)
    response, _ := ag.Chat(ctx, "Hello!")
    fmt.Println(response)
}
```

### Cách 3: Advanced - Provider Fallback

```go
func createProviderWithFallback() types.LLMProvider {
    // Try OpenAI first
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        return openai.New(apiKey, "gpt-4o")
    }
    
    // Fallback to Gemini
    if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
        provider, _ := gemini.New(apiKey, "gemini-2.0-flash")
        return provider
    }
    
    // Default to local Ollama
    return ollama.New("http://localhost:11434", "llama3.2")
}

func main() {
    provider := createProviderWithFallback()
    ag := agent.New(provider)
    // Rest is the same!
}
```

**🎯 Điểm Mạnh:**
- ✅ Switch provider = 1 dòng code
- ✅ API đồng nhất 100% - không cần học lại
- ✅ Type-safe - compile-time check
- ✅ Test dễ - mock provider interface

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| SDK breaking changes | High | Pin versions, vendor if needed |
| API differences | Medium | Abstraction layer, feature matrix |
| Increased complexity | Medium | Clear docs, examples |
| Package size growth | Low | Optional dependencies |
| Maintenance burden | High | Prioritize popular providers |

---

## 💡 Future Enhancements (Post v0.2)

1. **Anthropic Claude** provider
2. **Azure OpenAI** specific provider
3. **Provider fallback/retry** - try OpenAI, fallback to Gemini
4. **Cost tracking** per provider
5. **Provider pooling** - load balance across multiple
6. **Streaming aggregation** - combine multiple provider streams

---

## 📌 Decision Summary

**Recommended Approach:**
1. ✅ **Pin SDK versions** - `openai-go@v3.6.1`, `go-genai@v1.32.0` for stability
2. ✅ **Include all SDKs** - Users get consistent experience, easier maintenance
3. ✅ **Unified API enforced** - All providers implement same interface exactly
4. ✅ **Switch = 1 line change** - Only constructor differs, rest identical
5. ✅ **100% backward compatible** - v0.1 code works unchanged
6. ✅ **Ease of use first** - Hide SDK complexity completely

**Core Principles:**
- **API Uniformity** - Học 1 lần, dùng cho tất cả providers
- **Type Safety** - Interface contract enforced at compile time
- **Stability** - Pinned versions, controlled upgrades
- **Simplicity** - Provider switching transparent to user code

**Implementation Priority:**
1. **Week 1: OpenAI Provider** (v3.6.1 pinned)
   - Most popular, enterprise-ready
   - Validate unified API design
   
2. **Week 2: Gemini Provider** (v1.32.0 pinned)
   - Fast, cost-effective
   - Multimodal capabilities
   
3. **Week 3: Integration & Polish**
   - Factory pattern
   - Comprehensive tests
   - Migration guide

**Quality Gates (Each Provider):**
- [ ] Implements `types.LLMProvider` interface
- [ ] Constructor matches pattern: `New(apiKey, model string)`
- [ ] Chat/Stream work identically to Ollama provider
- [ ] Example code structure mirrors existing examples
- [ ] Tests match coverage level (~70%)
- [ ] Documentation shows provider switch = 1 line change

**Next Step:** 
Start implementing OpenAI provider with v3.6.1 following this plan!
