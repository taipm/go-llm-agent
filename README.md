# 🤖 go-llm-agent

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/taipm/go-llm-agent)](https://goreportcard.com/report/github.com/taipm/go-llm-agent)
[![Coverage](https://img.shields.io/badge/coverage-71.8%25-brightgreen.svg)](https://github.com/taipm/go-llm-agent)

**go-llm-agent** is a simple yet powerful Go library for building intelligent AI agents with tool usage and conversation context. **Now supports multiple LLM providers** with a unified API!

## 🌟 Multi-Provider Support (v0.2.0)

Switch between providers with **zero code changes** - just update your environment variables!

| Provider | Status | Use Case | Free Tier |
|----------|--------|----------|-----------|
| 🦙 **Ollama** | ✅ Ready | Local development, privacy | ✅ 100% Free |
| 🤖 **OpenAI** | ✅ Ready | Production, best quality | ❌ Paid API |
| ✨ **Gemini** | ✅ Ready | Large context, fast | ✅ Free tier available |

**One API, Three Providers!** See [PROVIDER_COMPARISON.md](PROVIDER_COMPARISON.md) for detailed comparison.

## ✨ Key Features

- 🚀 **Simple & Intuitive API** - Start building in minutes
- 🔄 **Multi-Provider Support** - Ollama, OpenAI, Gemini with unified interface
- 🏭 **Factory Pattern** - Auto-detect provider from environment
- 🔧 **28 Built-in Tools** - File ops, web, database, network, email automation (8 categories)
- 💬 **Conversation Memory** - Maintain context across conversations
- 📡 **Streaming Responses** - Real-time output for better UX
- 🎯 **Provider Flexibility** - Switch providers without code changes
- 📦 **Production Ready** - 71.8% test coverage, comprehensive error handling
- 🧪 **Fully Tested** - Compatibility tests across all providers
- 🔒 **Security First** - SSRF prevention, path validation, OAuth2 support

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

### Prerequisites

**Choose your provider:**

#### Option 1: Ollama (Local, Free)

- Go 1.25 or higher
- [Ollama](https://ollama.ai/) installed and running

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull qwen3:1.7b
```

#### Option 2: OpenAI (Cloud, Paid)

- Go 1.25 or higher
- OpenAI API key from [platform.openai.com](https://platform.openai.com/)

```bash
export OPENAI_API_KEY="sk-..."
```

#### Option 3: Gemini (Cloud, Free Tier Available)

- Go 1.25 or higher
- Gemini API key from [ai.google.dev](https://ai.google.dev/)

```bash
export GEMINI_API_KEY="..."
```

## 🚀 Quick Start

### Universal Provider Setup (Recommended)

Use the factory pattern to auto-detect your provider from environment variables:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    _ "github.com/joho/godotenv/autoload" // Optional: load .env file
)

func main() {
    // Auto-detect provider from environment
    // Supports: ollama, openai, gemini
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
    
    fmt.Println(response.Content)
    // Output: The capital of France is Paris.
}
```

**Environment Configuration** (`.env` file):

```bash
# Choose your provider
LLM_PROVIDER=ollama    # or: openai, gemini
LLM_MODEL=qwen3:1.7b   # or: gpt-4o-mini, gemini-2.5-flash

# Provider-specific settings
OLLAMA_BASE_URL=http://localhost:11434  # Ollama
OPENAI_API_KEY=sk-...                   # OpenAI
GEMINI_API_KEY=...                      # Gemini
```

**Switch providers** by changing just 2 lines in `.env` - **no code changes needed!**

### Provider-Specific Examples

<details>
<summary><b>Ollama (Local)</b></summary>

```go
package main

import (
    "context"
    "github.com/taipm/go-llm-agent/pkg/provider"
)

func main() {
    llm, err := provider.New(provider.Config{
        Type:    provider.ProviderOllama,
        BaseURL: "http://localhost:11434",
        Model:   "qwen3:1.7b",
    })
    // ... use llm.Chat()
}
```

</details>

<details>
<summary><b>OpenAI (Cloud)</b></summary>

```go
package main

import (
    "context"
    "os"
    "github.com/taipm/go-llm-agent/pkg/provider"
)

func main() {
    llm, err := provider.New(provider.Config{
        Type:   provider.ProviderOpenAI,
        APIKey: os.Getenv("OPENAI_API_KEY"),
        Model:  "gpt-4o-mini",
    })
    // ... use llm.Chat()
}
```

</details>

<details>
<summary><b>Gemini (Cloud)</b></summary>

```go
package main

import (
    "context"
    "os"
    "github.com/taipm/go-llm-agent/pkg/provider"
)

func main() {
    llm, err := provider.New(provider.Config{
        Type:   provider.ProviderGemini,
        APIKey: os.Getenv("GEMINI_API_KEY"),
        Model:  "gemini-2.5-flash",
    })
    // ... use llm.Chat()
}
```

</details>

### 1. Simple Chat (Legacy - Ollama Only)

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

### 2. Multi-Provider Example

All providers share the same API - switch with just environment variables!

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
    ctx := context.Background()
    
    // Auto-detect from environment
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Same code works with all providers!
    messages := []types.Message{
        {Role: types.RoleUser, Content: "What is 2+2?"},
    }
    
    response, err := llm.Chat(ctx, messages, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Provider: %s\n", llm)
    fmt.Printf("Answer: %s\n", response.Content)
    if response.Metadata != nil {
        fmt.Printf("Tokens: %d\n", response.Metadata.TotalTokens)
    }
}
```

**Test with different providers:**

```bash
# Test with Ollama (local, free)
echo "LLM_PROVIDER=ollama" > .env
echo "LLM_MODEL=qwen3:1.7b" >> .env
go run .

# Test with OpenAI (cloud)
echo "LLM_PROVIDER=openai" > .env
echo "LLM_MODEL=gpt-4o-mini" >> .env
echo "OPENAI_API_KEY=sk-..." >> .env
go run .

# Test with Gemini (cloud, free tier)
echo "LLM_PROVIDER=gemini" > .env
echo "LLM_MODEL=gemini-2.5-flash" >> .env
echo "GEMINI_API_KEY=..." >> .env
go run .
```

All run the **exact same code** - just different `.env` configuration!

### 3. Streaming Responses

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
    ctx := context.Background()
    
    // Works with any provider!
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    messages := []types.Message{
        {Role: types.RoleUser, Content: "Tell me a short story"},
    }
    
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
    
    err = llm.Stream(ctx, messages, nil, handler)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 4. Tool Calling (OpenAI & Gemini)

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
    ctx := context.Background()
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Define tools
    tools := []types.ToolDefinition{
        {
            Type: "function",
            Function: types.FunctionDefinition{
                Name:        "get_weather",
                Description: "Get weather for a location",
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
    
    messages := []types.Message{
        {Role: types.RoleUser, Content: "What's the weather in Tokyo?"},
    }
    
    options := &types.ChatOptions{Tools: tools}
    response, err := llm.Chat(ctx, messages, options)
    if err != nil {
        log.Fatal(err)
    }
    
    if len(response.ToolCalls) > 0 {
        for _, tc := range response.ToolCalls {
            fmt.Printf("Tool called: %s\n", tc.Function.Name)
            fmt.Printf("Arguments: %v\n", tc.Function.Arguments)
        }
    }
}
```

### 5. Multi-turn Conversation (Legacy - Agent)

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
    
    // Tạo agent (memory tự động khởi tạo với 100 messages)
    provider := ollama.New("http://localhost:11434", "llama3.2")
    agent := agent.New(provider)
    
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

**Tùy chỉnh kích thước memory:**

```go
import "github.com/taipm/go-llm-agent/memory"

mem := memory.NewBuffer(200) // Tùy chỉnh 200 messages
agent := agent.New(provider, agent.WithMemory(mem))
```

## � Built-in Tools (28 Tools, 8 Categories)

The library includes **28 production-ready tools** across 8 categories. Most tools are **auto-loaded by default** for instant use!

### Tool Categories

| Category | Tools | Auto-loaded | Description |
|----------|-------|-------------|-------------|
| 📁 **File** | 4 | ✅ Yes | Read, write, list, delete files with security |
| 🌐 **Web** | 3 | ✅ Yes | HTTP GET/POST, web scraping with SSRF prevention |
| 📅 **DateTime** | 3 | ✅ Yes | Current time, formatting, date calculations |
| 💻 **System** | 3 | ✅ Yes | System info, process list, installed apps |
| 🧮 **Math** | 2 | ✅ Yes | Expression evaluation, statistics |
| 🗄️ **Database** | 5 | ✅ Yes | MongoDB operations (connect, CRUD) |
| 🌍 **Network** | 5 | ✅ Yes | DNS, ping, WHOIS, SSL certs, IP geolocation |
| 📧 **Email** | 4 | ⚠️ Opt-in | Gmail integration (requires OAuth2) |

**Total: 24 auto-loaded + 4 Gmail (opt-in) = 28 tools**

### Quick Start with Built-in Tools

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/builtin"
    "github.com/taipm/go-llm-agent/pkg/provider"
)

func main() {
    ctx := context.Background()
    
    // 1. Auto-detect LLM provider
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Load all 24 built-in tools (one line!)
    registry := builtin.GetRegistry()
    tools := registry.ListTools()
    
    // 3. Create agent with tools
    agent := agent.New(llm, agent.WithTools(tools))
    
    // 4. Agent can now use all tools automatically!
    response, err := agent.Chat(ctx, "What files are in the current directory?")
    fmt.Println(response)
    // Agent will automatically call file_list tool
}
```

### Featured Tools

<details>
<summary><b>📁 File Tools</b> - Secure file operations</summary>

- `file_read` - Read file content with size limits
- `file_write` - Write/append content with backups
- `file_list` - List directory with pattern matching
- `file_delete` - Safe deletion with protected paths

**Security**: Path validation, size limits (10MB), directory traversal prevention

</details>

<details>
<summary><b>🌐 Web Tools</b> - HTTP requests and web scraping</summary>

- `web_fetch` - HTTP GET with SSRF prevention
- `web_post` - HTTP POST (JSON/form data)
- `web_scrape` - Extract data with CSS selectors

**Security**: Private IP blocking, domain whitelisting, timeout protection

</details>

<details>
<summary><b>🌍 Network Tools</b> - Professional network diagnostics</summary>

- `network_dns_lookup` - DNS queries (A, AAAA, MX, TXT, NS, CNAME, SOA, PTR)
- `network_ping` - ICMP ping & TCP connectivity checks
- `network_whois_lookup` - Domain registration info
- `network_ssl_cert_check` - SSL/TLS certificate validation
- `network_ip_info` - IP geolocation with GeoIP2 database

**Libraries**: miekg/dns, go-ping, likexian/whois, oschwald/geoip2

</details>

<details>
<summary><b>📧 Gmail Tools</b> - Email automation (OAuth2 required)</summary>

- `gmail_send` - Send emails (to, cc, bcc, HTML)
- `gmail_read` - Read messages by ID (full/metadata/minimal)
- `gmail_list` - List with filters & pagination
- `gmail_search` - Advanced search (Gmail query syntax)

**Setup**: Requires Google Cloud credentials ([Setup Guide](pkg/tools/gmail/README.md))

**Enable Gmail tools:**
```go
config := builtin.Config{NoGmail: false}
registry := builtin.GetRegistryWithConfig(config)
```

</details>

<details>
<summary><b>🗄️ MongoDB Tools</b> - Database operations</summary>

- `mongodb_connect` - Connection pooling (max 10)
- `mongodb_find` - Query with filtering/sorting
- `mongodb_insert` - Insert documents (batch up to 100)
- `mongodb_update` - UpdateOne/UpdateMany
- `mongodb_delete` - DeleteOne/DeleteMany with safety

**Safety**: Empty filter prevention, connection pool limits

</details>

<details>
<summary><b>🧮 Math & DateTime Tools</b> - Calculations and time operations</summary>

**Math**:
- `math_calculate` - Safe expression evaluation (govaluate)
- `math_stats` - Statistics (mean, median, mode, stddev) using gonum

**DateTime**:
- `datetime_now` - Current time with formats/timezones
- `datetime_format` - Format & timezone conversion
- `datetime_calc` - Date arithmetic (add/subtract/diff)

</details>

### Tool Configuration

```go
// Customize tool behavior
config := builtin.Config{
    // File tools
    NoFile: false,
    File: builtin.FileConfig{
        AllowedPaths:   []string{"/tmp", "/data"},
        ProtectedPaths: []string{"/etc", "/sys"},
        MaxFileSize:    5 * 1024 * 1024, // 5MB
    },
    
    // Web tools
    NoWeb: false,
    Web: builtin.WebConfig{
        AllowPrivateIPs: false,
        AllowedDomains:  []string{"api.example.com"},
        Timeout:         30 * time.Second,
    },
    
    // Gmail tools (disabled by default)
    NoGmail: false,
    Gmail: builtin.GmailConfig{
        Config: gmail.GmailConfig{
            CredentialsFile: "credentials.json",
            TokenFile:       "token.json",
        },
    },
}

registry := builtin.GetRegistryWithConfig(config)
```

See [pkg/builtin/README.md](pkg/builtin/README.md) for complete documentation.

## �📖 Documentation

- [🚀 QUICKSTART.md](QUICKSTART.md) - Get started in 5 minutes
- [� SPEC.md](SPEC.md) - Technical specification and architecture
- [🔀 PROVIDER_COMPARISON.md](PROVIDER_COMPARISON.md) - Provider comparison guide
- [📚 Examples](examples/) - Complete code examples for all providers
- [🔧 API Reference](https://pkg.go.dev/github.com/taipm/go-llm-agent) - Go package docs
- [📝 TODO.md](TODO.md) - Development progress and roadmap

## 🏗️ Architecture

```text
┌──────────────────────────────────────────────┐
│           Your Application                   │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│              Agent                           │
│  - Chat(), ChatStream()                      │
│  - Orchestrates LLM, Tools, Memory           │
└──────────┬───────────────────────────────────┘
           │
    ┌──────┼──────────┐
    │      │          │
┌───▼──┐ ┌─▼──────┐ ┌▼────────┐
│ LLM  │ │ Tools  │ │ Memory  │
│Multi-│ │ System │ │ Manager │
│Prov- │ └────────┘ └─────────┘
│ider  │
└──┬───┘
   │
   ├─ Ollama (Local)
   ├─ OpenAI (Cloud)
   └─ Gemini (Cloud)
```

### Core Components

1. **Agent** - Orchestrates workflow, manages context
2. **Multi-Provider System** - Unified interface for Ollama, OpenAI, Gemini
3. **Factory Pattern** - Auto-detect provider from environment
4. **Tool System** - Enable agents to perform real actions
5. **Memory** - Store and manage conversation context
6. **Streaming** - Real-time response streaming

## 🛣️ Roadmap

### ✅ v0.1.0 - Foundation (Oct 26, 2025)

- ✅ Basic agent with Ollama
- ✅ Simple tool system
- ✅ In-memory conversation history
- ✅ Working examples
- ✅ **Streaming responses** (bonus!)

### ✅ v0.2.0 - Multi-Provider (Oct 27-31, 2025)

- ✅ **OpenAI provider** (gpt-4o-mini, gpt-4o)
- ✅ **Gemini provider** (gemini-2.5-flash, gemini-2.0-pro)
- ✅ **Factory pattern** for provider auto-detection
- ✅ **Unified API** - switch providers with zero code changes
- ✅ **Comprehensive testing** - 71.8% coverage, compatibility tests
- ✅ **Provider comparison guide**
- ⏸️ Documentation updates (in progress)
- ⏸️ Migration guide (pending)

### 🔮 v0.3.0 - Advanced Features (Future)

- ✅ **28 built-in tools** across 8 categories (file, web, datetime, system, math, database, network, email)
  - ✅ File operations (read, write, list, delete)
  - ✅ Web tools (HTTP GET/POST, web scraping)
  - ✅ DateTime tools (now, format, calculate)
  - ✅ System tools (info, processes, apps)
  - ✅ Math tools (calculate, statistics)
  - ✅ MongoDB database tools (connect, find, insert, update, delete)
  - ✅ Network diagnostic tools (DNS, ping, WHOIS, SSL, IP info)
  - ✅ Gmail tools (send, read, list, search) - OAuth2 required
- [ ] Persistent storage (SQLite, PostgreSQL)
- [ ] Vector database integration (Qdrant)
- [ ] ReAct pattern implementation
- [ ] Multi-agent collaboration
- [ ] Advanced configuration system
- [ ] Performance optimizations
- [ ] Azure OpenAI dedicated provider
- [ ] Anthropic Claude support
- [ ] Production monitoring & metrics

[See TODO.md for detailed development progress](TODO.md)

## 🎯 Which Provider Should I Use?

| Scenario | Recommended Provider | Why |
|----------|---------------------|-----|
| **Local development** | 🦙 Ollama | Free, fast, no internet needed |
| **Production apps** | 🤖 OpenAI | Best quality, reliable, proven |
| **Large context tasks** | ✨ Gemini | 1M+ token context window |
| **Cost-sensitive** | 🦙 Ollama or ✨ Gemini | Free (Ollama) or free tier (Gemini) |
| **Privacy-critical** | 🦙 Ollama | 100% local, nothing leaves your machine |
| **Tool calling** | 🤖 OpenAI or ✨ Gemini | Better tool support than local models |

See [PROVIDER_COMPARISON.md](PROVIDER_COMPARISON.md) for detailed comparison.

## 🤝 Contributing

This project is in active development. Contributions are welcome!

```bash
# Clone repository
git clone https://github.com/taipm/go-llm-agent.git
cd go-llm-agent

# Install dependencies
go mod download

# Run tests
go test ./...

# Run compatibility tests (requires providers)
export OPENAI_API_KEY="sk-..."  # Optional
export GEMINI_API_KEY="..."     # Optional
go test ./pkg/provider -run=TestCompatibility -v

# Run examples
cd examples/simple_chat && go run .
cd examples/multi_provider && go run .
```

**Development Guidelines:**
- Write tests for new features
- Follow Go best practices
- Update documentation
- Test with multiple providers

## 🌟 Star History

If you find this project useful, please consider giving it a star ⭐

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Ollama](https://ollama.ai/) - Excellent local LLM runtime
- [OpenAI](https://openai.com/) - Leading AI research and APIs
- [Google AI](https://ai.google.dev/) - Gemini API and documentation
- [LangChain](https://github.com/langchain-ai/langchain) - Architecture inspiration
- Go Community - For an amazing language and ecosystem

## 📧 Contact

- **Author**: taipm
- **GitHub**: [@taipm](https://github.com/taipm)
- **Repository**: [go-llm-agent](https://github.com/taipm/go-llm-agent)
- **Issues**: [GitHub Issues](https://github.com/taipm/go-llm-agent/issues)
- **Discussions**: [GitHub Discussions](https://github.com/taipm/go-llm-agent/discussions)

---

**Built with ❤️ using Go** | **Multi-Provider Support since v0.2.0**

---

**⚠️ Status**: Alpha - API có thể thay đổi. Không khuyến khích dùng trong production.

**🌟 Star this repo** nếu bạn thấy project hữu ích!
