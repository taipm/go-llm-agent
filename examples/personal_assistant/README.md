# Personal Assistant Example

A practical demonstration of the go-llm-agent library showcasing a real-world use case: an intelligent personal assistant that can help with daily tasks.

## Features

### 🎯 Core Capabilities
- **Web Research**: Search and fetch information from the internet
- **Calculations**: Perform math operations, statistics, compound interest
- **File Management**: List, read, and manage files
- **System Monitoring**: Check CPU, memory, disk usage
- **Network Tools**: DNS lookups, ping, SSL certificate checks
- **Time Management**: Current time, date calculations, countdowns

### 🧠 Self-Learning
- Records all interactions and learns from experience
- Improves tool selection based on past success
- Detects and learns from error patterns
- Provides better responses over time

### 💬 Two Modes
1. **Demo Mode**: Pre-built scenarios showcasing various capabilities
2. **Interactive Mode**: Free-form chat with the assistant

## Prerequisites

### Required
- Go 1.25+
- Ollama running locally with a model installed

### Recommended (for learning features)
- Qdrant vector database for experience storage
```bash
docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant
```

## Quick Start

### 1. Install Ollama and Download a Model
```bash
# Install Ollama (macOS/Linux)
curl -fsSL https://ollama.ai/install.sh | sh

# Download a recommended model (choose one)
ollama pull qwen2.5:3b    # Recommended - good balance
ollama pull llama3.2:3b   # Alternative
ollama pull gemma2:2b     # Lighter option
```

### 2. (Optional) Start Qdrant for Learning
```bash
docker run -d -p 6334:6334 -p 6333:6333 --name qdrant qdrant/qdrant
```

### 3. Run the Assistant
```bash
cd examples/personal_assistant
go run main.go
```

## Usage Examples

### Demo Mode
Runs 7 practical scenarios:
1. **Daily Information**: Get current date and time
2. **Web Research**: Search and summarize AI news
3. **Calculations**: Calculate compound interest
4. **File Management**: List directory contents
5. **System Information**: Check CPU/memory usage
6. **Network Tools**: DNS lookup
7. **Time Calculations**: Days until Christmas

### Interactive Mode
Free-form chat examples:
```
You: What's the current time and date?
🤖 Assistant: Today is January 27, 2025, Monday. The current time is 3:45 PM.

You: Calculate 15% of 250
🤖 Assistant: 15% of 250 is 37.5

You: Search for Golang best practices
🤖 Assistant: [Searches web and provides summary]

You: status
📊 Shows agent status with learning metrics

You: help
📖 Displays example questions
```

## Special Commands

- `status` - View agent status and learning metrics
- `help` - Show example questions
- `exit` or `quit` - Exit the program

## Learning Features

When Qdrant is available, the assistant will:
- ✅ Remember all past interactions
- ✅ Learn which tools work best for different tasks
- ✅ Detect recurring error patterns
- ✅ Improve success rate over time

Status display shows:
```
🧠 LEARNING SYSTEM:
   ✅ Enabled: true
   📚 Total Experiences: 42
   ✨ Success Rate: 87.5%
   🎯 Tool Selector: Active
   🔍 Error Analyzer: Active
```

## Available Tool Categories

The assistant has access to:
- **DateTime**: Current time, date calculations, formatting
- **File**: Read, write, list files
- **Web**: Fetch pages, search (requires API keys)
- **System**: Execute commands, environment variables, system info
- **Math**: Calculations, statistics
- **Network**: DNS, ping, WHOIS, SSL checks, IP info

## Configuration

### Change LLM Model
Edit `main.go`:
```go
llm := ollama.New("your-model-name")
```

### Disable Tool Categories
```go
registry := builtin.GetRegistry(builtin.Config{
    NoGmail:   true,  // Gmail requires OAuth
    NoNetwork: true,  // Disable network tools
    NoMongoDB: true,  // MongoDB requires server
})
```

### Adjust Log Level
```go
agent.WithLogLevel(logger.LogLevelDebug)  // More verbose
agent.WithLogLevel(logger.LogLevelWarn)   // Less verbose
```

## Troubleshooting

### "VectorMemory not available"
- Learning features disabled but assistant still works
- To enable: Start Qdrant (`docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant`)

### "No tools available"
- Check that builtin tools are loaded
- Verify Ollama is running

### Slow responses
- Try a smaller model (`qwen2.5:1.5b`)
- Reduce number of tools loaded

### Model not found
```bash
ollama list                    # Check installed models
ollama pull qwen2.5:3b        # Download model
```

## What Makes This Example Useful?

1. **Real-World Application**: Not just a toy demo, actually useful for daily tasks
2. **Self-Improving**: Gets better over time with learning system
3. **User-Friendly**: Interactive menus and helpful prompts
4. **Educational**: Shows all major library features in action
5. **Customizable**: Easy to modify for specific use cases

## Next Steps

After trying this example, explore:
- Add custom tools for your specific needs
- Integrate with your own APIs
- Build domain-specific assistants (DevOps, Data Analysis, etc.)
- Add voice input/output
- Create a web interface

## Related Examples

- `examples/learning_agent` - Focus on learning system internals
- `examples/zero_config_agent` - Minimal setup example
- `examples/agent_with_builtin_tools` - Complete tool showcase

## License

This example is part of the go-llm-agent library.
