# Agent Status Example

This example demonstrates how to inspect agent configuration and runtime state using the `Status()` method.

## What It Shows

The `agent.Status()` method provides comprehensive information about:

1. **Configuration**: System prompt, temperature, max tokens, iterations, confidence threshold
2. **Reasoning Capabilities**: Auto-reasoning, CoT, ReAct, Reflection availability
3. **Tools**: Total count and list of available tools
4. **Memory**: Type, message count, search/vector support
5. **Provider**: LLM provider type

## Use Cases

- 🐛 **Debugging**: Verify agent is configured correctly
- 📊 **Monitoring**: Track runtime state and capabilities
- ✅ **Validation**: Check capabilities before execution
- 📝 **Reporting**: Generate system configuration reports
- 🔍 **Inspection**: Understand agent state at any point

## Run the Example

```bash
# Copy environment config
cp .env.example .env

# Edit .env with your LLM provider settings
# LLM_PROVIDER=ollama
# LLM_MODEL=qwen3:1.7b

# Run
go run main.go
```

## Example Output

```
📊 Scenario 1: Default Agent Configuration

🎛️  Configuration:
   System Prompt: You are a helpful AI assistant.
   Temperature: 0.70
   Max Tokens: 2000
   Max Iterations: 10
   Min Confidence: 0.70
   Reflection Enabled: true

🧠 Reasoning Capabilities:
   Auto-Reasoning: true
   CoT Available: false
   ReAct Available: false
   Reflection Available: false

🔧 Tools:
   Total Count: 25
   Available: [datetime_format system_apps mongodb_update ...]

💾 Memory:
   Type: buffer
   Message Count: 0
   Supports Search: false
   Supports Vectors: false

🤖 Provider:
   Type: ollama
```

## Scenarios Demonstrated

### 1. Default Agent
Shows agent with default configuration and all builtin tools.

### 2. Customized Agent
Agent with custom prompt, temperature, disabled reflection, and no auto-reasoning.

### 3. Minimal Tools Agent
Agent with no builtin tools (useful for specialized use cases).

## API Usage

```go
// Get agent status
status := agent.Status()

// Access configuration
fmt.Printf("Temperature: %.2f\n", status.Configuration.Temperature)
fmt.Printf("Reflection: %v\n", status.Configuration.EnableReflection)

// Check capabilities
if status.Reasoning.CoTAvailable {
    fmt.Println("Chain-of-Thought reasoning is available")
}

// List tools
fmt.Printf("Tools: %d\n", status.Tools.TotalCount)
for _, tool := range status.Tools.ToolNames {
    fmt.Println("  -", tool)
}

// Export as JSON
jsonData, _ := json.MarshalIndent(status, "", "  ")
fmt.Println(string(jsonData))
```

## JSON Export

The status can be exported as JSON for:
- Configuration backups
- System documentation
- Monitoring dashboards
- CI/CD validation

```json
{
  "configuration": {
    "system_prompt": "You are a helpful AI assistant.",
    "temperature": 0.7,
    "max_tokens": 2000,
    "max_iterations": 10,
    "min_confidence": 0.7,
    "enable_reflection": true
  },
  "reasoning": {
    "auto_reasoning_enabled": true,
    "cot_available": false,
    "react_available": false,
    "reflection_available": false
  },
  "tools": {
    "total_count": 25,
    "tool_names": ["datetime_format", "system_apps", ...]
  },
  "memory": {
    "type": "buffer",
    "supports_search": false,
    "supports_vectors": false
  },
  "provider": {
    "type": "ollama"
  }
}
```

## Notes

- **Lazy Initialization**: Reasoning engines (CoT, ReAct, Reflection) are lazy-initialized
  - They show as "available: false" until first use
  - After first Chat() call, they may become available
- **Memory Count**: Only available for BufferMemory and AdvancedMemory types
- **Provider Type**: Auto-detected from provider implementation

## Related Examples

- `examples/simple_agent/` - Basic agent usage
- `examples/reflection_agent/` - Self-reflection demonstration
- `examples/cot_reasoning/` - Chain-of-Thought reasoning
- `examples/react_with_tools/` - ReAct pattern with tools
