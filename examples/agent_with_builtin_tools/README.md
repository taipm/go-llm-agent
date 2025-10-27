# Agent with All Built-in Tools

Comprehensive example demonstrating an AI agent with all 20 built-in tools and multi-provider support.

## Features

- **Multi-provider support**: Auto-detect Ollama, OpenAI, or Gemini from environment
- **20 built-in tools** across 6 categories:
  - File Operations (4 tools)
  - Web Operations (3 tools)
  - DateTime (3 tools)
  - System (3 tools)
  - Math (2 tools)
  - Database (5 tools)
- **Conversation memory**: Maintains context across multiple turns
- **Automatic tool selection**: Agent chooses appropriate tools automatically
- **Real-world testing**: Demonstrates actual tool usage patterns

## Prerequisites

Choose one provider:

### Option 1: Gemini (Recommended - Free tier available)
```bash
# Get API key from https://ai.google.dev
export GEMINI_API_KEY="your_api_key"
```

### Option 2: Ollama (Local, 100% Free)
```bash
# Install and run Ollama
brew install ollama
ollama pull qwen3:1.7b
```

### Option 3: OpenAI (Paid)
```bash
# Get API key from https://platform.openai.com
export OPENAI_API_KEY="sk-..."
```

## Setup

1. Copy environment template:
```bash
cp .env.example .env
```

2. Edit `.env` and configure your provider:
```bash
# For Gemini
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.5-flash
GEMINI_API_KEY=your_key_here

# OR for Ollama
LLM_PROVIDER=ollama
LLM_MODEL=qwen3:1.7b

# OR for OpenAI
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o-mini
OPENAI_API_KEY=sk-...
```

## Run

```bash
go run .
```

## What It Does

The example demonstrates:

1. **Tool Loading**: Lists all 20 built-in tools by category
2. **Math Operations**: Calculates expressions using `math_calculate`
3. **DateTime**: Gets current date/time with `datetime_now`
4. **System Info**: Retrieves OS details with `system_info`
5. **File Operations**: Lists directory with `file_list`
6. **Memory Testing**: Multi-turn conversation maintaining context

## Sample Output

```
=== Agent with All Built-in Tools Demo ===

✓ Using provider: gemini
✓ Loaded 20 built-in tools

Available tools by category:
File Operations:
  - file_read (safe)
  - file_list (safe)
  - file_write (unsafe)
  - file_delete (unsafe)
Web Operations:
  - web_fetch (safe)
  - web_post (unsafe)
  - web_scrape (safe)
DateTime:
  - datetime_now (safe)
  - datetime_format (safe)
  - datetime_calc (safe)
System:
  - system_info (safe)
  - system_processes (safe)
  - system_apps (safe)
Math:
  - math_calculate (safe)
  - math_stats (safe)
Database:
  - mongodb_connect (safe)
  - mongodb_find (safe)
  - mongodb_insert (unsafe)
  - mongodb_update (unsafe)
  - mongodb_delete (unsafe)

✓ Agent created with 20 tools

=== Testing Agent Capabilities ===

Test 1: Math calculation
Query: What is the result of (100 + 50) * 2 - 30?
Response: The result is 270.
Tools called: math_calculate
✓ Success

Test 2: DateTime operation
Query: What is the current date and time?
Response: The current date and time is October 27, 2025 at 10:30:45 AM.
Tools called: datetime_now
✓ Success

Test 3: System information
Query: What operating system am I running?
Response: You are running macOS 14.5.
Tools called: system_info
✓ Success

Test 4: File operation
Query: List files in the current directory
Response: The current directory contains: main.go, .env, .env.example, README.md
Tools called: file_list
✓ Success

=== Testing Conversation Memory ===

Turn 1: My favorite number is 42
Response: Got it, I'll remember that your favorite number is 42.
✓ Success

Turn 2: What is my favorite number?
Response: Your favorite number is 42.
✓ Success

Turn 3: Calculate my favorite number multiplied by 2
Response: 42 multiplied by 2 equals 84.
Tools called: math_calculate
✓ Success

=== Memory Statistics ===
Total messages in memory: 18
Memory capacity: 50 messages
  - User messages: 7
  - Assistant messages: 7
  - Tool messages: 4

=== Demo Complete ===

Key Features Demonstrated:
✓ Multi-provider support (auto-detect from environment)
✓ 20 built-in tools across 6 categories
✓ Automatic tool selection and execution
✓ Conversation memory across multiple turns
✓ Math, DateTime, System, File operations

The agent successfully:
  - Performed calculations using math_calculate tool
  - Retrieved system information
  - Managed file operations
  - Maintained conversation context
  - Remembered user preferences across turns
```

## Switch Providers

Simply update your `.env` file - **no code changes needed**:

```bash
# Test with Gemini (fast, large context)
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.5-flash
go run .

# Test with Ollama (local, private)
LLM_PROVIDER=ollama
LLM_MODEL=qwen3:1.7b
go run .

# Test with OpenAI (best quality)
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o-mini
go run .
```

## Code Structure

```go
// 1. Auto-detect provider from environment
llm, err := provider.FromEnv()

// 2. Load all built-in tools
registry := builtin.GetRegistry()
tools := registry.ListTools() // 20 tools

// 3. Create agent with memory
mem := memory.NewBuffer(50)
ag := agent.New(llm, agent.WithMemory(mem))

// 4. Register all tools
for _, tool := range tools {
    ag.AddTool(tool)
}

// 5. Execute queries - agent auto-calls tools
response, err := ag.Execute(ctx, messages)
```

## Next Steps

- Modify test queries to try different tools
- Add custom tools alongside built-in ones
- Test with MongoDB tools (requires MongoDB running)
- Experiment with different providers
- Build your own agent-based application

## Troubleshooting

**Provider not found**:
- Check `.env` file is in the same directory
- Verify API keys are set correctly
- Make sure Ollama is running if using local provider

**Tool execution fails**:
- Some tools require specific setup (e.g., MongoDB for database tools)
- Check tool requirements in main documentation

**Out of memory errors**:
- Reduce memory buffer size (currently 50 messages)
- Clear conversation history between tests

## Learn More

- [Main README](../../README.md) - Library overview
- [Provider Comparison](../../PROVIDER_COMPARISON.md) - Choose the right provider
- [Built-in Tools](../../TODO.md) - Complete tools reference
- [API Documentation](https://pkg.go.dev/github.com/taipm/go-llm-agent)
