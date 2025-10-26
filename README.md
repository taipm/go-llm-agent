# 🤖 go-llm-agent

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-alpha-orange.svg)](ROADMAP.md)

**go-llm-agent** là thư viện Go đơn giản, mạnh mẽ để xây dựng AI agents thông minh với khả năng sử dụng tools và duy trì context. Bắt đầu với Ollama support, tiến hóa dần để support nhiều LLM providers.

## ✨ Tính năng chính

- 🚀 **Simple & Intuitive API** - Dễ học, dễ dùng trong vài phút
- 🔧 **Tool System** - Cho phép agent thực hiện các tác vụ thực tế
- 💬 **Conversation Memory** - Duy trì context qua nhiều lượt hội thoại
- 📡 **Streaming Responses** - Real-time output cho UX tốt hơn
- 🦙 **Ollama First** - Chạy local models miễn phí với Ollama
- 📦 **Zero Dependencies** - Chỉ dùng Go standard library
- 🧪 **Production Ready** - Test coverage cao, error handling tốt

## 🎯 Use Cases

- **Chatbots thông minh** với khả năng truy cập dữ liệu thực
- **CLI assistants** tự động hóa workflows
- **Data analysis agents** xử lý và phân tích dữ liệu
- **Code assistants** hỗ trợ development tasks
- **Customer service bots** với domain knowledge

## 📦 Installation

```bash
go get github.com/taipm/go-llm-agent
```

**Prerequisites:**
- Go 1.21 trở lên
- [Ollama](https://ollama.ai/) đã cài đặt và đang chạy

```bash
# Cài đặt Ollama (nếu chưa có)
curl -fsSL https://ollama.ai/install.sh | sh

# Pull model
ollama pull llama3.2
```

## 🚀 Quick Start

### 1. Simple Chat

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/agent"
    "github.com/taipm/go-llm-agent/provider/ollama"
)

func main() {
    ctx := context.Background()
    
    // Tạo Ollama provider
    provider := ollama.New("http://localhost:11434", "llama3.2")
    
    // Tạo agent
    agent := agent.New(provider)
    
    // Chat
    response, err := agent.Chat(ctx, "What is the capital of France?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
    // Output: The capital of France is Paris.
}
```

### 2. Agent with Tools

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/taipm/go-llm-agent/agent"
    "github.com/taipm/go-llm-agent/provider/ollama"
    "github.com/taipm/go-llm-agent/tool"
)

// Định nghĩa một tool đơn giản
type WeatherTool struct{}

func (w *WeatherTool) Name() string {
    return "get_weather"
}

func (w *WeatherTool) Description() string {
    return "Get current weather for a location"
}

func (w *WeatherTool) Parameters() *tool.JSONSchema {
    return &tool.JSONSchema{
        Type: "object",
        Properties: map[string]*tool.JSONSchema{
            "location": {
                Type:        "string",
                Description: "City name",
            },
        },
        Required: []string{"location"},
    }
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    location := params["location"].(string)
    
    // Mock weather data
    return map[string]interface{}{
        "location":    location,
        "temperature": 22,
        "condition":   "Sunny",
        "timestamp":   time.Now().Format(time.RFC3339),
    }, nil
}

func main() {
    ctx := context.Background()
    
    // Setup agent với tool
    provider := ollama.New("http://localhost:11434", "llama3.2")
    agent := agent.New(provider)
    
    // Đăng ký tool
    weatherTool := &WeatherTool{}
    agent.AddTool(weatherTool)
    
    // Agent sẽ tự động gọi tool khi cần
    response, err := agent.Chat(ctx, "What's the weather like in Tokyo?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
    // Output: Based on current data, it's sunny in Tokyo with temperature of 22°C.
}
```

### 3. Streaming Responses (New!)

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/agent"
    "github.com/taipm/go-llm-agent/provider/ollama"
    "github.com/taipm/go-llm-agent/types"
)

func main() {
    ctx := context.Background()
    
    provider := ollama.New("http://localhost:11434", "llama3.2")
    agent := agent.New(provider)
    
    // Stream response in real-time
    handler := func(chunk types.StreamChunk) error {
        if chunk.Content != "" {
            fmt.Print(chunk.Content) // Print as tokens arrive
        }
        if chunk.Done {
            fmt.Println("\n✓ Done")
        }
        return nil
    }
    
    err := agent.ChatStream(ctx, "Tell me a short story", handler)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 4. Multi-turn Conversation

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/agent"
    "github.com/taipm/go-llm-agent/memory"
    "github.com/taipm/go-llm-agent/provider/ollama"
)

func main() {
    ctx := context.Background()
    
    // Tạo agent với memory
    provider := ollama.New("http://localhost:11434", "llama3.2")
    mem := memory.NewBuffer(100) // Lưu 100 messages
    
    agent := agent.New(provider, agent.WithMemory(mem))
    
    // Conversation
    resp1, _ := agent.Chat(ctx, "My name is John and I love programming")
    fmt.Println(resp1)
    
    resp2, _ := agent.Chat(ctx, "What's my name?")
    fmt.Println(resp2)
    // Output: Your name is John.
    
    resp3, _ := agent.Chat(ctx, "What do I love?")
    fmt.Println(resp3)
    // Output: You love programming.
}
```

## 📖 Documentation

- [📋 SPEC.md](SPEC.md) - Đặc tả kỹ thuật chi tiết
- [🗺️ ROADMAP.md](ROADMAP.md) - Kế hoạch phát triển
- [📚 Examples](examples/) - Code examples đầy đủ
- [🔧 API Reference](https://pkg.go.dev/github.com/taipm/go-llm-agent) - Go package docs

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│           Your Application                  │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│              Agent                          │
│  - Chat(), Run(), Execute()                 │
│  - Orchestrates LLM, Tools, Memory          │
└─────────────┬───────────────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
┌───▼────┐ ┌─▼──────┐ ┌▼────────┐
│  LLM   │ │ Tools  │ │ Memory  │
│Provider│ │ System │ │ Manager │
└────────┘ └────────┘ └─────────┘
```

### Core Components

1. **Agent** - Trung tâm điều phối, quản lý workflow
2. **LLM Provider** - Interface với các LLM (Ollama, OpenAI, v.v.)
3. **Tool System** - Cho phép agent thực hiện actions
4. **Memory** - Lưu trữ và quản lý conversation context

## 🛣️ Roadmap

### ✅ v0.1 - Foundation (Current)

- Basic agent với Ollama
- Simple tool system
- In-memory conversation history
- Working examples
- **Streaming responses** ✨ (New)

### 🔄 v0.2 - Enhanced (Next)

- 10+ built-in tools
- Advanced configuration
- Performance optimizations
- Better error handling

### 🔮 v0.3 - Multi-Provider

- OpenAI/Azure OpenAI support
- Anthropic Claude support
- Persistent storage
- Production features

[Chi tiết đầy đủ tại ROADMAP.md](ROADMAP.md)

## 🤝 Contributing

Dự án đang trong giai đoạn đầu phát triển. Mọi đóng góp đều được hoan nghênh!

```bash
# Clone repository
git clone https://github.com/taipm/go-llm-agent.git
cd go-llm-agent

# Run tests
go test ./...

# Run examples
go run examples/simple_chat/main.go
```

## 📝 License

MIT License - xem [LICENSE](LICENSE) để biết chi tiết.

## 🙏 Acknowledgments

- [Ollama](https://ollama.ai/) - Local LLM runtime tuyệt vời
- [LangChain](https://github.com/langchain-ai/langchain) - Inspiration cho architecture
- Go Community - Vì một ngôn ngữ tuyệt vời

## 📧 Contact

- Author: taipm
- GitHub: [@taipm](https://github.com/taipm)
- Issues: [GitHub Issues](https://github.com/taipm/go-llm-agent/issues)

---

**⚠️ Status**: Alpha - API có thể thay đổi. Không khuyến khích dùng trong production.

**🌟 Star this repo** nếu bạn thấy project hữu ích!
