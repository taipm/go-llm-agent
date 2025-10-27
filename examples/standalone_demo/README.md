# Standalone Demo - Go LLM Agent

This is a complete, standalone example that you can copy to any project to test the `go-llm-agent` library.

## Features

This demo includes:

- ✅ **20 Built-in Tools** across 6 categories
  - DateTime (3): Get time, calculate dates, format
  - Math (2): Evaluate expressions, statistics
  - File (3): Read, write, list
  - Web (2): Fetch pages, search
  - System (3): Execute commands, env vars, system info
  - MongoDB (6): Complete database operations

- ✅ **Professional Logging** with colored output and visual indicators
- ✅ **Memory Management** - Agent remembers conversation history
- ✅ **Multi-Provider Support** - Works with Ollama, OpenAI, Gemini
- ✅ **Interactive Chat** - REPL-style conversation interface
- ✅ **Vietnamese Language** - Full Unicode support

## Quick Start

### 1. Copy to Your Project

```bash
# Create a new directory
mkdir test-llm-agent
cd test-llm-agent

# Initialize Go module
go mod init test-llm-agent

# Copy this example
cp -r /path/to/go-llm-agent/examples/standalone_demo/* .
```

### 2. Create `.env` File

Choose one of the following configurations:

#### Option A: Ollama (Local, Free)

```bash
cat > .env << 'EOF'
LLM_PROVIDER=ollama
LLM_MODEL=qwen2.5:7b
OLLAMA_BASE_URL=http://localhost:11434
EOF
```

#### Option B: OpenAI (Cloud, API Key Required)

```bash
cat > .env << 'EOF'
LLM_PROVIDER=openai
LLM_MODEL=gpt-4
OPENAI_API_KEY=your-api-key-here
EOF
```

#### Option C: Google Gemini (Cloud, API Key Required)

```bash
cat > .env << 'EOF'
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.0-flash-exp
GEMINI_API_KEY=your-api-key-here
EOF
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run

```bash
go run main.go
```

## Example Conversations

### Basic Questions

```
👤 You: What time is it now?
🤖 Agent: The current date and time is October 27, 2025, at 11:30:45 AM.

👤 You: Calculate 15% of 350
🤖 Agent: 15% of 350 is 52.5.

👤 You: What's the mean of [10, 20, 30, 40, 50]?
🤖 Agent: The mean (average) is 30.
```

### Date Calculations

```
👤 You: How many days until Christmas 2025?
🤖 Agent: There are 59 days until Christmas 2025.

👤 You: I was born on March 15, 1990. How old am I?
🤖 Agent: You are 35 years old.
```

### Vietnamese Examples

```
👤 You: Tôi sinh ngày 15/03/1990, năm nay tôi bao nhiêu tuổi?
🤖 Agent: Bạn sinh ngày 15/03/1990, vậy năm nay bạn 35 tuổi.

👤 You: Tính 25 * 4 + 100
🤖 Agent: Kết quả là 200.

👤 You: Còn bao nhiêu ngày nữa đến Tết Nguyên Đán 2026?
🤖 Agent: Còn 95 ngày nữa đến Tết Nguyên Đán 2026 (29/01/2026).
```

### File Operations

```
👤 You: List files in the current directory
🤖 Agent: Here are the files:
- main.go
- go.mod
- go.sum
- .env

👤 You: Create a file called hello.txt with content "Hello, World!"
🤖 Agent: File created successfully at hello.txt.
```

### System Information

```
👤 You: What's my operating system?
🤖 Agent: You're running macOS Darwin 23.0.0 on arm64 architecture.

👤 You: What's the value of my HOME environment variable?
🤖 Agent: Your HOME directory is /Users/yourusername.
```

## Commands

While chatting, you can use these commands:

- `quit` or `exit` - Exit the program
- `clear` - Clear conversation history (fresh start)
- `help` - Show help menu with all available tools

## Logging Output

The agent shows detailed logs of what it's doing:

```
👤 User: Calculate 25 * 4 + 100
11:30:45 [INFO] 👤 User: Calculate 25 * 4 + 100
11:30:45 [INFO] 🤔 Agent thinking...
11:30:46 [INFO] 🔧 Agent wants to call 1 tool(s): math_calculate
11:30:46 [INFO] 🔧 Calling tool: math_calculate
11:30:46 [INFO]    Parameters: {"expression":"25 * 4 + 100"}
11:30:46 [INFO] ✓ Tool math_calculate completed successfully
11:30:46 [INFO]    Result: 200
11:30:46 [INFO] 🤔 Agent thinking...
11:30:47 [INFO] 💬 Agent response:
11:30:47 [INFO]    The result is 200.
```

## Customization

## Customization

### Change Log Level

In `main.go`, find this line:

```go
a := agent.New(llm, agent.WithLogLevel(agent.LogLevelInfo))
```

Change to:
- `agent.LogLevelDebug` - See everything (including LLM responses)
- `agent.LogLevelWarn` - Only warnings and errors
- Or use `agent.DisableLogging()` - No logs at all

### Add Custom Memory Size

By default, the agent uses 100 messages memory. To customize:

```go
import "github.com/taipm/go-llm-agent/pkg/memory"

mem := memory.NewBuffer(200) // Remember last 200 messages
a := agent.New(llm, agent.WithMemory(mem))
```

You can add your own tools:

```go
customTool := agent.Tool{
    Name:        "my_custom_tool",
    Description: "Does something useful",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type":        "string",
                "description": "Input parameter",
            },
        },
        "required": []string{"input"},
    },
    Func: func(params map[string]interface{}) (string, error) {
        input := params["input"].(string)
        return fmt.Sprintf("Processed: %s", input), nil
    },
}

allTools = append(allTools, customTool)
```

### Adjust Memory Size

Memory is initialized automatically with 100 messages. To customize:

```go
import "github.com/taipm/go-llm-agent/pkg/memory"

mem := memory.NewBuffer(200) // Remember last 200 messages
a := agent.New(llm, agent.WithMemory(mem))
```

## Troubleshooting

### "Failed to initialize LLM provider"

Make sure your `.env` file has the correct settings:
- `LLM_PROVIDER` is set (ollama, openai, or gemini)
- API keys are set for cloud providers
- Ollama is running for local provider

### Ollama Connection Error

Start Ollama:
```bash
ollama serve
```

Pull the model:
```bash
ollama pull qwen2.5:7b
```

### Rate Limits (OpenAI/Gemini)

If you hit rate limits, try:
- Using a smaller model
- Adding delays between requests
- Switching to Ollama (local, no limits)

### Tool Not Working

Check the logs to see what the agent is trying to do. Enable debug logging:

```go
agent.WithLogLevel(agent.LogLevelDebug),
```

## What's Happening Under the Hood

1. **User Input** → Agent receives your message
2. **Memory** → Agent retrieves conversation history
3. **LLM Call** → Agent asks LLM what to do
4. **Tool Selection** → LLM decides which tools to use
5. **Tool Execution** → Agent runs the tools
6. **Tool Results** → Agent sends results back to LLM
7. **Response** → LLM generates final answer
8. **Memory Update** → Agent saves the conversation

All of this is visible in the logs!

## Production Usage

For production applications:

1. **Error Handling** - Add proper error handling for your use case
2. **Rate Limiting** - Implement rate limiting for API calls
3. **Logging** - Configure logging level based on environment
4. **Security** - Validate user input, sanitize file paths, restrict commands
5. **Monitoring** - Track tool usage, errors, response times
6. **Testing** - Write tests for your custom tools

## Learn More

- GitHub: https://github.com/taipm/go-llm-agent
- Documentation: See README.md in the repository
- Examples: Check the `examples/` directory for more use cases

## License

This example is part of the go-llm-agent library (MIT License).
