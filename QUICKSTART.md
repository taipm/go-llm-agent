# 🚀 Quick Start Guide

Get started with **go-llm-agent** in 5 minutes! This guide covers multi-provider setup and basic usage.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Provider Setup](#provider-setup)
- [Your First Chat](#your-first-chat)
- [Streaming Responses](#streaming-responses)
- [Tool Calling](#tool-calling)
- [Next Steps](#next-steps)

## Prerequisites

**Required:**
- Go 1.25 or higher

**Choose ONE provider** (or use all three):

### Option 1: Ollama (Recommended for Beginners)

✅ **Best for:** Local development, learning, privacy  
✅ **Cost:** 100% Free  
✅ **Setup time:** 5 minutes

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull qwen3:1.7b

# Verify it's running
curl http://localhost:11434
# Should return: Ollama is running
```

### Option 2: OpenAI

✅ **Best for:** Production, best quality  
✅ **Cost:** Pay-per-use ($0.15/1M tokens for gpt-4o-mini)  
✅ **Setup time:** 2 minutes

1. Get API key from [platform.openai.com](https://platform.openai.com/api-keys)
2. Set environment variable:

```bash
export OPENAI_API_KEY="sk-..."
```

### Option 3: Gemini

✅ **Best for:** Large context, free tier  
✅ **Cost:** Free tier available  
✅ **Setup time:** 2 minutes

1. Get API key from [ai.google.dev](https://aistudio.google.com/app/apikey)
2. Set environment variable:

```bash
export GEMINI_API_KEY="..."
```

## 📦 Installation

```bash
go get github.com/taipm/go-llm-agent
```

## 🎯 Provider Setup

### Create `.env` File (Recommended)

Create a `.env` file in your project root:

```bash
# Choose your provider (ollama, openai, or gemini)
LLM_PROVIDER=ollama
LLM_MODEL=qwen3:1.7b

# Ollama configuration (if using Ollama)
OLLAMA_BASE_URL=http://localhost:11434

# OpenAI configuration (if using OpenAI)
# OPENAI_API_KEY=sk-...

# Gemini configuration (if using Gemini)
# GEMINI_API_KEY=...
```

### Install godotenv (Optional but Recommended)

```bash
go get github.com/joho/godotenv
```

## 🤖 Your First Chat

### Step 1: Create `main.go`

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/types"
    _ "github.com/joho/godotenv/autoload" // Load .env file
)

func main() {
    // Auto-detect provider from environment
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create conversation
    ctx := context.Background()
    messages := []types.Message{
        {Role: types.RoleUser, Content: "What is the capital of France?"},
    }
    
    // Get response
    response, err := llm.Chat(ctx, messages, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Content)
    // Output: The capital of France is Paris.
}
```

### Step 2: Run It

```bash
go run main.go
```

**That's it!** 🎉 You just built your first AI chat application.

### Switch Providers (Zero Code Changes!)

Just update your `.env` file:

```bash
# Switch to OpenAI
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o-mini
OPENAI_API_KEY=sk-...

# Switch to Gemini  
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.5-flash
GEMINI_API_KEY=...
```

Run the **exact same code** - it works with all providers!

## 📡 Streaming Responses

Get real-time output as the model generates text:

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
    
    // Stream handler
    handler := func(chunk types.StreamChunk) error {
        // Print each token as it arrives
        fmt.Print(chunk.Content)
        
        if chunk.Done {
            fmt.Println("\n✓ Complete")
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

**Output:**
```text
1
2
3
...
10
✓ Complete
```

## 🔧 Tool Calling

Let your AI agent use external functions:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/taipm/go-llm-agent/pkg/provider"
    "github.com/taipm/go-llm-agent/pkg/types"
    _ "github.com/joho/godotenv/autoload"
)

func main() {
    llm, err := provider.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Define a weather tool
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
                            Description: "City name (e.g., Tokyo, Paris)",
                        },
                        "unit": {
                            Type:        "string",
                            Description: "Temperature unit",
                            Enum:        []interface{}{"celsius", "fahrenheit"},
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
    
    // Chat with tools
    options := &types.ChatOptions{
        Tools: tools,
    }
    
    response, err := llm.Chat(ctx, messages, options)
    if err != nil {
        log.Fatal(err)
    }
    
    // Check if the model wants to call a tool
    if len(response.ToolCalls) > 0 {
        for _, tc := range response.ToolCalls {
            fmt.Printf("Tool: %s\n", tc.Function.Name)
            fmt.Printf("Arguments: %v\n", tc.Function.Arguments)
            
            // Execute the tool (your implementation)
            result := executeWeatherTool(tc.Function.Arguments)
            
            // Return result to model
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
    } else {
        fmt.Println(response.Content)
    }
}

func executeWeatherTool(args map[string]interface{}) string {
    location := args["location"].(string)
    // Mock implementation
    return fmt.Sprintf(`{"location": "%s", "temperature": 22, "condition": "Sunny", "timestamp": "%s"}`, 
        location, time.Now().Format(time.RFC3339))
}
```

**Note:** Tool calling works best with OpenAI and Gemini. Some Ollama models support it (like qwen3:1.7b).

## 📚 Next Steps

### Explore Examples

Check out complete working examples:

```bash
# Multi-provider example (recommended)
cd examples/multi_provider
go run .

# Provider-specific examples
cd examples/openai_chat && go run .
cd examples/gemini_chat && go run .
cd examples/simple_chat && go run .
```

### Learn More

- [📋 SPEC.md](SPEC.md) - Technical specification
- [🔀 PROVIDER_COMPARISON.md](PROVIDER_COMPARISON.md) - Detailed provider comparison
- [📖 README.md](README.md) - Full documentation
- [📝 TODO.md](TODO.md) - Development roadmap

### Advanced Topics

1. **Multi-turn Conversations**
   - Build conversation history
   - Maintain context across messages

2. **Custom Tools**
   - Implement your own tool interfaces
   - Connect to databases, APIs, files

3. **Production Deployment**
   - Error handling best practices
   - Rate limiting and retry logic
   - Cost monitoring (for cloud providers)

4. **Performance Optimization**
   - Batch requests
   - Caching strategies
   - Model selection

## 🆘 Troubleshooting

### Ollama Connection Error

```bash
# Check if Ollama is running
curl http://localhost:11434

# If not, start it
ollama serve

# Pull model if not exists
ollama pull qwen3:1.7b
```

### OpenAI API Key Error

```bash
# Verify API key is set
echo $OPENAI_API_KEY

# Test with curl
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

### Gemini API Key Error

```bash
# Verify API key is set
echo $GEMINI_API_KEY

# Make sure API is enabled in Google Cloud Console
```

### Provider Not Detected

```bash
# Check .env file exists and has correct format
cat .env

# Make sure LLM_PROVIDER is set
# Valid values: ollama, openai, gemini
```

## 💡 Tips

1. **Start with Ollama** - It's free and runs locally, perfect for learning
2. **Use `.env` files** - Makes switching providers effortless  
3. **Test with all providers** - Ensure your app works everywhere
4. **Monitor costs** - OpenAI and Gemini charge per token
5. **Check model availability** - Not all models support all features (especially tool calling)

## 🎯 Quick Provider Comparison

| Feature | Ollama | OpenAI | Gemini |
|---------|--------|--------|--------|
| **Cost** | Free | Paid | Free tier |
| **Speed** | Fast (local) | Fast | Very fast |
| **Quality** | Good | Excellent | Excellent |
| **Privacy** | 100% local | Cloud | Cloud |
| **Tool calling** | Some models | All models | All models |
| **Context size** | Varies | 128K+ | 1M+ |

See [PROVIDER_COMPARISON.md](PROVIDER_COMPARISON.md) for detailed comparison.

---

**Ready to build?** Start with the simplest example above and gradually add features! 🚀
