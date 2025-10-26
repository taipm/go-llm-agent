# go-llm-agent Project Structure

```
go-llm-agent/
├── README.md              # Project overview and quick start
├── SPEC.md               # Technical specification
├── ROADMAP.md            # Development roadmap (v0.1, v0.2, v0.3)
├── QUICKSTART.md         # 5-minute quick start guide
├── CONTRIBUTING.md       # Contribution guidelines
├── LICENSE               # MIT License
├── Makefile             # Build automation
├── go.mod               # Go module definition
│
├── pkg/                 # Main library code
│   ├── types/          # Core types and interfaces
│   │   └── types.go    # Message, Tool, Provider, Memory interfaces
│   │
│   ├── provider/       # LLM provider implementations
│   │   └── ollama/     # Ollama provider
│   │       └── ollama.go
│   │
│   ├── agent/          # Agent orchestration
│   │   └── agent.go    # Main agent implementation
│   │
│   ├── tool/           # Tool system
│   │   ├── registry.go      # Tool registry
│   │   └── registry_test.go # Tests
│   │
│   └── memory/         # Memory management
│       ├── buffer.go        # Buffer memory implementation
│       └── buffer_test.go   # Tests
│
└── examples/           # Example programs
    ├── simple_chat/    # Basic chat example
    │   └── main.go
    │
    ├── tool_usage/     # Tool usage example
    │   └── main.go
    │
    ├── conversation/   # Multi-turn conversation
    │   └── main.go
    │
    └── tools/          # Reusable tool implementations
        ├── calculator.go
        └── weather.go
```

## Key Files

### Core Library (`pkg/`)

1. **types/types.go** - Foundation
   - `Message` - Chat message structure
   - `Tool` interface - Tool definition
   - `LLMProvider` interface - LLM abstraction
   - `Memory` interface - History management
   - `JSONSchema` - Parameter schemas

2. **provider/ollama/ollama.go** - Ollama Integration
   - HTTP client for Ollama API
   - Message format conversion
   - Tool calling support

3. **agent/agent.go** - Agent Core
   - Orchestrates LLM, Tools, Memory
   - Handles tool calling loop
   - Manages conversation flow

4. **tool/registry.go** - Tool Management
   - Thread-safe tool registry
   - Tool execution
   - Schema validation (planned)

5. **memory/buffer.go** - Memory
   - In-memory FIFO buffer
   - Configurable size
   - Thread-safe operations

### Examples

1. **simple_chat/** - Basic Usage
   - Single-turn conversations
   - No tools, no memory
   - ~40 lines of code

2. **tool_usage/** - Tools Demo
   - Calculator and Weather tools
   - Automatic tool calling
   - ~60 lines of code

3. **conversation/** - Memory Demo
   - Multi-turn conversation
   - Context preservation
   - ~65 lines of code

4. **tools/** - Reusable Tools
   - `CalculatorTool` - Math operations
   - `WeatherTool` - Mock weather data

## Usage Patterns

### Import Paths

```go
import (
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/memory"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
    "github.com/taipm/go-llm-agent/pkg/types"
)
```

### Create Agent

```go
// Simple agent
provider := ollama.New("http://localhost:11434", "llama3.2")
ag := agent.New(provider)

// With memory
mem := memory.NewBuffer(50)
ag := agent.New(provider, agent.WithMemory(mem))

// With custom options
ag := agent.New(provider, 
    agent.WithMemory(mem),
    agent.WithTemperature(0.8),
    agent.WithSystemPrompt("You are a helpful assistant"),
)
```

### Add Tools

```go
type MyTool struct{}

func (t *MyTool) Name() string { return "my_tool" }
func (t *MyTool) Description() string { return "Does something" }
func (t *MyTool) Parameters() *types.JSONSchema { /* ... */ }
func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Tool logic
    return result, nil
}

// Register
ag.AddTool(&MyTool{})
```

## Build & Test

```bash
# Run tests
make test

# Test with coverage
make test-coverage

# Build examples
make build

# Run examples
make run-simple
make run-tools
make run-conv

# All checks
make check
```

## Version 0.1 Features ✅

- ✅ Core agent implementation
- ✅ Ollama provider
- ✅ Tool system with registry
- ✅ Buffer memory
- ✅ 2 example tools (calculator, weather)
- ✅ 3 working examples
- ✅ Unit tests (70%+ coverage)
- ✅ Documentation

## Next Steps (v0.2)

See [ROADMAP.md](ROADMAP.md) for planned features:
- Streaming responses
- 10+ built-in tools
- Advanced configuration
- Performance optimizations
