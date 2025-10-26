# Quick Start Guide

This guide will help you get started with go-llm-agent in 5 minutes.

## Prerequisites

1. **Install Go 1.21+**
   ```bash
   # Check version
   go version
   ```

2. **Install Ollama**
   ```bash
   # macOS
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Or download from https://ollama.ai
   ```

3. **Pull a model**
   ```bash
   ollama pull llama3.2
   
   # Verify Ollama is running
   curl http://localhost:11434
   ```

## Installation

```bash
go get github.com/taipm/go-llm-agent
```

## Your First Agent

### 1. Simple Chat

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
    // Create provider
    provider := ollama.New("http://localhost:11434", "llama3.2")
    
    // Create agent
    ag := agent.New(provider)
    
    // Chat!
    response, err := ag.Chat(context.Background(), "Hello!")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
}
```

Run it:
```bash
go run main.go
```

### 2. Agent with Tools

Add a calculator tool:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
    "github.com/taipm/go-llm-agent/pkg/types"
)

// Simple calculator tool
type CalcTool struct{}

func (c *CalcTool) Name() string {
    return "calculator"
}

func (c *CalcTool) Description() string {
    return "Calculate square root"
}

func (c *CalcTool) Parameters() *types.JSONSchema {
    return &types.JSONSchema{
        Type: "object",
        Properties: map[string]*types.JSONSchema{
            "number": {
                Type: "number",
                Description: "Number to calculate",
            },
        },
        Required: []string{"number"},
    }
}

func (c *CalcTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    num := params["number"].(float64)
    return math.Sqrt(num), nil
}

func main() {
    provider := ollama.New("http://localhost:11434", "llama3.2")
    ag := agent.New(provider)
    
    // Add tool
    ag.AddTool(&CalcTool{})
    
    // Agent will use the tool when needed
    response, err := ag.Chat(context.Background(), "What is the square root of 144?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
}
```

### 3. Conversation with Memory

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/memory"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
    provider := ollama.New("http://localhost:11434", "llama3.2")
    mem := memory.NewBuffer(50)
    
    ag := agent.New(provider, agent.WithMemory(mem))
    
    ctx := context.Background()
    
    // First message
    resp1, _ := ag.Chat(ctx, "My name is Alice")
    fmt.Println(resp1)
    
    // Agent remembers!
    resp2, _ := ag.Chat(ctx, "What's my name?")
    fmt.Println(resp2)
}
```

## Next Steps

1. **Explore Examples**: Check `examples/` directory
2. **Read Documentation**: See `SPEC.md` for details
3. **Build Custom Tools**: Create tools for your use case
4. **Contribute**: See `CONTRIBUTING.md`

## Common Issues

### Ollama not running
```bash
# Start Ollama
ollama serve
```

### Model not found
```bash
# List available models
ollama list

# Pull a model
ollama pull llama3.2
```

### Connection refused
Check Ollama is running on `http://localhost:11434`

## Resources

- [Examples](./examples)
- [API Documentation](https://pkg.go.dev/github.com/taipm/go-llm-agent)
- [Roadmap](./ROADMAP.md)
- [Spec](./SPEC.md)

Happy coding! 🚀
