# ReAct Agent with Vector Memory Example

This example demonstrates the **ReAct (Reasoning + Acting)** pattern combined with **Vector Memory** for semantic search capabilities.

## Features Demonstrated

### 1. **ReAct Pattern**
- Explicit reasoning with Thought→Action→Observation→Reflection loop
- Transparent thinking process (debuggable AI)
- Step-by-step problem solving
- Automatic reasoning history tracking

### 2. **Vector Memory (Optional)**
- Semantic search using Qdrant vector database
- Automatic embedding generation (Ollama nomic-embed-text)
- Find similar past conversations by meaning
- Memory statistics and analytics

### 3. **Dual Memory Modes**
- **Simple Mode**: BufferMemory (FIFO, no dependencies)
- **Advanced Mode**: VectorMemory with semantic search

## Prerequisites

### Required
- Go 1.25+
- Ollama running locally (`http://localhost:11434`)
- A model installed (e.g., `ollama pull qwen2.5:3b`)

### Optional (for Vector Memory)
- Qdrant running locally (`localhost:6334`)
- Embedding model: `ollama pull nomic-embed-text`

## Quick Start

### 1. Run with Simple Memory (No Qdrant Required)

```bash
cd examples/react_memory

# Set environment variables
export LLM_PROVIDER=ollama
export LLM_MODEL=qwen2.5:3b

# Run
go run main.go
```

### 2. Run with Vector Memory (Qdrant Required)

```bash
# Start Qdrant (if not running)
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant

# Pull embedding model
ollama pull nomic-embed-text

# Set environment variables
export LLM_PROVIDER=ollama
export LLM_MODEL=qwen2.5:3b
export USE_VECTOR_MEMORY=true

# Run
go run main.go
```

## Example Output

```
=== ReAct Agent with Vector Memory Example ===

📡 Connecting to Ollama...
✅ LLM Provider ready

🧠 Setting up Vector Memory with Qdrant...
✅ Vector Memory ready (semantic search enabled)

�� Creating ReAct Agent...
✅ ReAct Agent ready

======================================================================
Question 1: What is 15 * 23 + 47?
======================================================================

=== ReAct Iteration 1 ===
💭 Thought: I need to calculate 15 * 23 first, then add 47
🎯 Action: Calculate multiplication
👁️  Observation: 15 * 23 = 345
🤔 Reflection: Now I need to add 47 to get the final answer

=== ReAct Iteration 2 ===
💭 Thought: Adding 47 to 345
🎯 Action: Answer
👁️  Observation: Executed: Answer
🤔 Reflection: The calculation is complete

----------------------------------------------------------------------
✅ Final Answer: 345 + 47 = 392
----------------------------------------------------------------------

📝 Reasoning History:
=== ReAct Reasoning History ===

Iteration 1 (14:23:15):
  💭 Thought: I need to calculate 15 * 23 first, then add 47
  🎯 Action: Calculate multiplication
  👁️  Observation: 15 * 23 = 345
  🤔 Reflection: Now I need to add 47 to get the final answer

Iteration 2 (14:23:17):
  💭 Thought: Adding 47 to 345
  🎯 Action: Answer
  👁️  Observation: Executed: Answer
  🤔 Reflection: The calculation is complete

======================================================================
🔍 Testing Semantic Search in Vector Memory
======================================================================

Searching for: 'mathematics calculations'
Found 3 related memories:

1. [assistant] Solved: What is 15 * 23 + 47?
   Used 2 reasoning steps
   Answer: 392

2. [assistant] Solved: If I have 100 dollars and spend 35% on food, how much do I have left?
   Used 2 reasoning steps
   Answer: 65 dollars

3. [assistant] ReAct Step 1: I need to calculate 15 * 23 first, then add 47

📊 Memory Statistics:
  Total Messages: 12
  Vector Count: 12

======================================================================
🎉 Demo completed!
======================================================================
```

## What's Happening?

### Step 1: Setup
- Connects to Ollama LLM provider
- Optionally creates Qdrant collection for vector storage
- Initializes embedding generator (nomic-embed-text)

### Step 2: ReAct Reasoning
For each question:
1. **Thought**: Agent thinks about what to do
2. **Action**: Agent decides on an action
3. **Observation**: Result of the action
4. **Reflection**: What was learned

### Step 3: Memory Storage
- Each ReAct step is stored with metadata
- Messages are categorized (reasoning, factual, etc.)
- Embeddings are generated automatically
- Stored in both hot cache and Qdrant

### Step 4: Semantic Search
- Query: "mathematics calculations"
- Vector similarity search finds related memories
- Returns relevant past conversations
- Works across different phrasings (semantic understanding)

## Key Code Snippets

### Creating ReAct Agent
```go
agent := reasoning.NewReActAgent(llm, memory, 5)
agent.SetVerbose(true) // Show thinking process
```

### Solving with ReAct
```go
answer, err := agent.Solve(ctx, "What is 2+2?")
// Agent will think through the problem step-by-step
```

### Vector Memory Setup
```go
vectorMem, err := memory.NewVectorMemory(ctx, memory.VectorMemoryConfig{
    QdrantURL:      "localhost:6334",
    CollectionName: "my_agent",
    Embedder:       memory.NewOllamaEmbedder("", "nomic-embed-text"),
})
```

### Semantic Search
```go
results, err := vectorMem.SearchSemantic(ctx, "find math problems", 5)
// Returns 5 most similar messages by semantic meaning
```

## Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `LLM_PROVIDER` | LLM provider name | `ollama` | `ollama`, `openai` |
| `LLM_MODEL` | Model to use | - | `qwen2.5:3b`, `gpt-4o-mini` |
| `OLLAMA_BASE_URL` | Ollama server URL | `http://localhost:11434` | - |
| `USE_VECTOR_MEMORY` | Enable vector memory | `false` | `true` |

## Troubleshooting

### "Failed to connect to Qdrant"
- Make sure Qdrant is running: `docker ps | grep qdrant`
- Check port 6334 is available
- Or run without vector memory (set `USE_VECTOR_MEMORY=false`)

### "Failed to generate embedding"
- Pull embedding model: `ollama pull nomic-embed-text`
- Check Ollama is running: `ollama list`

### "LLM call failed"
- Verify Ollama is running: `curl http://localhost:11434`
- Check model is installed: `ollama list`

## Next Steps

1. Try different questions (math, logic, explanation)
2. Experiment with `maxSteps` parameter
3. Test semantic search with various queries
4. Explore memory statistics and categorization
5. Combine with tool calling (file_read, web_search, etc.)

## Related Examples

- `examples/simple_agent/` - Basic agent without ReAct
- `examples/tool_agent/` - Agent with tool usage
- `examples/streaming/` - Streaming responses

## Learn More

- [ReAct Paper](https://arxiv.org/abs/2210.03629)
- [Qdrant Documentation](https://qdrant.tech/documentation/)
- [go-llm-agent Documentation](../../README.md)
