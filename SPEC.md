# Đặc tả Thư viện go-llm-agent

## 1. Mục tiêu dự án

**go-llm-agent** là một thư viện Go lightweight, dễ sử dụng để xây dựng các AI agent thông minh có khả năng:
- Tương tác với LLM (bắt đầu với Ollama)
- Sử dụng tools/functions để thực hiện các tác vụ cụ thể
- Lưu trữ và sử dụng context/memory
- Xử lý conversation có nhiều vòng tương tác

### Nguyên tắc phát triển
- **LEAN**: Tập trung vào tính năng thiết yếu, chạy được ngay
- **80/20**: Ưu tiên 20% tính năng tạo ra 80% giá trị
- **Iterative**: Tiến hóa qua nhiều phiên bản nhỏ
- **Simple API**: Dễ học, dễ dùng, dễ mở rộng

## 2. Kiến trúc tổng quan

```
┌─────────────────────────────────────────────┐
│              Application Code               │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│              Agent Interface                │
│  - Run(), Chat(), Execute()                 │
└─────────────┬───────────────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
┌───▼────┐ ┌─▼──────┐ ┌▼────────┐
│  LLM   │ │ Tools  │ │ Memory  │
│Provider│ │ System │ │ Manager │
└────────┘ └────────┘ └─────────┘
```

## 3. Các thành phần chính

### 3.1. Agent
**Trách nhiệm**: Điều phối giữa LLM, Tools và Memory

**Chức năng cốt lõi**:
- Nhận input từ người dùng
- Gửi request đến LLM
- Xử lý tool calls từ LLM response
- Quản lý conversation flow
- Lưu trữ lịch sử hội thoại

**Interface cơ bản**:
```go
type Agent interface {
    Chat(ctx context.Context, message string) (string, error)
    Run(ctx context.Context, task string) (*Result, error)
    AddTool(tool Tool) error
    Reset() error
}
```

### 3.2. LLM Provider
**Trách nhiệm**: Trừu tượng hóa việc giao tiếp với các LLM khác nhau

**Phase 1 - Ollama Support**:
- Gửi messages đến Ollama API
- Xử lý streaming responses
- Hỗ trợ function calling (tool use)
- Quản lý system prompts

**Interface**:
```go
type LLMProvider interface {
    Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error)
    Stream(ctx context.Context, messages []Message, options *ChatOptions) (<-chan StreamChunk, error)
}

type OllamaProvider struct {
    baseURL string
    model   string
    client  *http.Client
}
```

### 3.3. Tool System
**Trách nhiệm**: Cho phép agent thực hiện các hành động cụ thể

**Chức năng**:
- Đăng ký tools với tên và schema
- Validate input parameters
- Thực thi tool functions
- Trả về kết quả cho LLM

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

### 3.4. Memory Manager
**Trách nhiệm**: Quản lý context và lịch sử hội thoại

**Phase 1 - Simple Memory**:
- In-memory conversation history
- Message buffer với giới hạn
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

## 5. Workflow cơ bản

### Execution Flow
```
1. User Input → Agent
2. Agent → Add to Memory
3. Agent → Format messages for LLM
4. Agent → Send to LLM Provider
5. LLM → Response with potential tool calls
6. If tool calls:
   a. Agent → Execute tools
   b. Agent → Add results to messages
   c. Go back to step 4
7. Else:
   a. Agent → Return final response
   b. Agent → Save to Memory
```

## 6. Ví dụ sử dụng

### 6.1. Simple Chat
```go
agent := llmagent.NewAgent(
    llmagent.WithOllama("http://localhost:11434", "llama3.2"),
)

response, err := agent.Chat(ctx, "What is the capital of France?")
fmt.Println(response) // "The capital of France is Paris."
```

### 6.2. Agent with Tools
```go
// Define a tool
weatherTool := &WeatherTool{}

agent := llmagent.NewAgent(
    llmagent.WithOllama("http://localhost:11434", "llama3.2"),
)
agent.AddTool(weatherTool)

response, err := agent.Chat(ctx, "What's the weather in Tokyo?")
// Agent will call weatherTool.Execute() and use result
```

### 6.3. Multi-turn Conversation
```go
agent := llmagent.NewAgent(
    llmagent.WithOllama("http://localhost:11434", "llama3.2"),
    llmagent.WithMemory(llmagent.NewBufferMemory(100)),
)

agent.Chat(ctx, "My name is John")
agent.Chat(ctx, "What's my name?") // "Your name is John"
```

## 7. Yêu cầu kỹ thuật

### Dependencies tối thiểu
- Go 1.21+
- Standard library
- HTTP client (net/http)

### Optional dependencies
- JSON schema validation
- Vector database client (future)

### Testing
- Unit tests cho từng component
- Integration tests với Ollama
- Example programs

## 8. Non-functional Requirements

### Performance
- Hỗ trợ streaming responses
- Timeout configuration
- Connection pooling

### Reliability
- Error handling rõ ràng
- Retry logic cho network calls
- Graceful degradation

### Maintainability
- Clean code, well documented
- Examples cho mọi tính năng
- Semantic versioning

## 9. Giới hạn phiên bản đầu (v0.1)

**Out of scope cho v0.1**:
- Multiple LLM providers (chỉ Ollama)
- Vector database integration
- Persistent storage
- Advanced memory strategies
- Streaming support
- Multi-agent systems
- Fine-tuning support

**Focus v0.1**:
- Basic agent với Ollama
- Simple tool system
- In-memory conversation history
- Clear, simple API
- Working examples

## 10. Success Metrics

**v0.1 thành công khi**:
- ✅ Có thể chat với Ollama model
- ✅ Có thể đăng ký và sử dụng ít nhất 2 tools
- ✅ Có thể maintain conversation context
- ✅ Có ít nhất 3 working examples
- ✅ Documentation đầy đủ cho basic usage
- ✅ Code coverage >= 70%
