# Learning System Guide

## Quick Start

### Zero-Config Setup (Simplest Way)

```go
package main

import (
    "context"
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider/ollama"
)

func main() {
    // 1. (Optional) Start Qdrant for full learning features
    // docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant
    
    // 2. Create LLM provider
    llm := ollama.New("http://localhost:11434", "qwen2.5:7b")
    
    // 3. Create agent - THAT'S IT! Learning auto-enabled!
    ag := agent.New(llm)
    
    // Behind the scenes:
    // - Learning is enabled by default ✅
    // - Auto-tries VectorMemory if Qdrant running ✅
    // - Gracefully falls back to BufferMemory if Qdrant unavailable ✅
    // - No errors, no panics - just works! ✅
    
    // 4. Use agent normally - it learns automatically!
    response, _ := ag.Chat(context.Background(), "Calculate 15 * 23")
    // With Qdrant running: Full learning with semantic search
    // Without Qdrant: Limited learning mode (still works!)
}
```

### Advanced Setup (Custom Configuration)

If you need custom settings:

```go
// Create custom vector memory
vectorMem, _ := memory.NewVectorMemory(context.Background(), 
    memory.VectorMemoryConfig{
        QdrantURL:      "localhost:6334",
        CollectionName: "my_memory",
        Embedder:       memory.NewOllamaEmbedder("localhost:11434", "nomic-embed-text"),
        CacheSize:      200,
    },
)

// Create agent with custom memory
ag := agent.New(llm,
    agent.WithMemory(vectorMem),  // Custom memory
    // Learning still auto-enabled!
)
```

## What You Get

### Automatic Learning Features

1. **Experience Tracking**
   - Every Chat() call is automatically recorded
   - Captures: query, intent, reasoning mode, tools used, success/failure, latency
   - Stored in vector memory for semantic search (if Qdrant available)

2. **Tool Selection Learning (ε-greedy)**
   - Agent learns which tools work best for different queries
   - 90% exploitation: Uses best tool based on past success
   - 10% exploration: Tries random tools to discover new patterns
   - Automatically improves success rate over time

3. **Performance Optimization**
   - Tracks tool success rates and latencies
   - Prioritizes fast, reliable tools
   - Learns from mistakes

## API

### Query Tool Recommendations

```go
// Get learned recommendation for a query
rec, err := ag.GetToolRecommendation(ctx, 
    "Calculate the area of a circle with radius 5",
    "calculation",
)

fmt.Printf("Recommended tool: %s\n", rec.ToolName)
fmt.Printf("Confidence: %.2f\n", rec.Confidence)
fmt.Printf("Success rate: %.0f%%\n", rec.SuccessRate * 100)
fmt.Printf("Sample size: %d\n", rec.SampleSize)
fmt.Printf("Reasoning: %s\n", rec.Reasoning)
```

### View Tool Statistics

```go
// Get performance stats for a specific tool
stats, err := ag.GetToolStats(ctx, "math_calculate", "calculation")

fmt.Printf("Tool: %s\n", stats.ToolName)
fmt.Printf("Total calls: %d\n", stats.TotalCalls)
fmt.Printf("Success rate: %.0f%%\n", stats.SuccessRate * 100)
fmt.Printf("Average latency: %dms\n", stats.AvgLatency)
```

### Check Learning Status

```go
// Check if learning is active
status := ag.Status()

fmt.Printf("Learning enabled: %v\n", status.Learning.Enabled)
fmt.Printf("Experience store ready: %v\n", status.Learning.ExperienceStoreReady)
fmt.Printf("Tool selector ready: %v\n", status.Learning.ToolSelectorReady)
fmt.Printf("Session ID: %s\n", status.Learning.ConversationID)
```

## How It Works

### 1. Experience Collection

Every `agent.Chat()` call:
```
User: "What is 15 * 23?"
   ↓
Agent selects: math_calculate tool
   ↓
Executes: 15 * 23 = 345
   ↓
Records experience:
  - Query: "What is 15 * 23?"
  - Intent: "calculation"
  - Tool: "math_calculate"
  - Success: true
  - Latency: 150ms
  - Reasoning: "simple"
   ↓
Stores in vector memory (semantic search enabled)
```

### 2. Learning Algorithm (ε-greedy)

When agent needs to select a tool:

```
90% of the time (Exploitation):
  1. Query similar past experiences from vector memory
  2. Calculate tool statistics (success rate, latency)
  3. Score each tool: 70% success rate + 30% speed
  4. Select highest scoring tool
  
10% of the time (Exploration):
  1. Pick random tool
  2. Try it to discover new patterns
  3. Learn from the result
```

### 3. Continuous Improvement

Over time:
- Success rate increases (learns which tools work)
- Latency decreases (prefers faster tools)
- Error rate drops (avoids tools that fail often)
- Query understanding improves (semantic memory)

## Example: Learning in Action

```go
// Day 1: First time asking math question
response1, _ := ag.Chat(ctx, "Calculate 100 * 200")
// - Might try wrong tool (exploration)
// - Records result
// - Learns: "math questions → math_calculate works well"

// Day 1: Second math question
response2, _ := ag.Chat(ctx, "What is 50 + 75?")
// - Checks past experiences
// - Sees math_calculate worked before
// - Uses it again (exploitation)
// - Success rate improves!

// Day 2: After 20 math queries
rec, _ := ag.GetToolRecommendation(ctx, "Compute 456 / 12", "calculation")
// Recommendation:
//   Tool: "math_calculate"
//   Confidence: 0.95
//   Success rate: 100%
//   Sample size: 20
//   Reasoning: "Used successfully 20/20 times (100%) with avg latency 120ms"
```

## Requirements

### Must Have
1. **Qdrant Vector Database**
   ```bash
   docker run -p 6333:6333 qdrant/qdrant
   ```

2. **Vector Memory**
   ```go
   vectorMem, _ := memory.NewVectorMemory(
       "http://localhost:6333",
       "agent_memory",
       memory.WithOllamaEmbedder("http://localhost:11434", "nomic-embed-text"),
   )
   ```

3. **Learning Enabled**
   ```go
   agent.WithLearning(true)
   ```

### Optional Configuration

```go
ag := agent.New(llm,
    agent.WithLearning(true),
    agent.WithMemory(vectorMem),
    
    // Optional: Adjust learning parameters (advanced)
    // These are applied via the agent instance after creation
)

// Configure tool selector after agent creation (optional)
// Note: This requires accessing internal components (advanced usage)
```

## Troubleshooting

### "Learning disabled: VectorMemory required"

**Problem:** Using BufferMemory instead of VectorMemory

**Solution:**
```go
// ❌ Wrong - BufferMemory doesn't support learning
agent.New(llm, agent.WithLearning(true))

// ✅ Correct - Use VectorMemory
vectorMem, _ := memory.NewVectorMemory(...)
agent.New(llm, 
    agent.WithLearning(true),
    agent.WithMemory(vectorMem),
)
```

### "Failed to connect to Qdrant"

**Problem:** Qdrant is not running

**Solution:**
```bash
# Start Qdrant
docker run -p 6333:6333 qdrant/qdrant

# Verify it's running
curl http://localhost:6333
```

### No recommendations available

**Problem:** Not enough experience data

**Solution:**
- Learning requires at least 3 experiences per tool/intent combination
- Use the agent normally, it will learn over time
- Check `GetToolStats()` to see sample sizes

## Best Practices

1. **Use Consistent Intents**
   - Agent auto-detects: calculation, information_retrieval, file_operation, coding, conversation
   - Consistent queries help learning

2. **Monitor Learning Progress**
   ```go
   // Periodically check stats
   stats, _ := ag.GetToolStats(ctx, "math_calculate", "calculation")
   if stats.SampleSize >= 10 {
       fmt.Printf("Learned! Success rate: %.0f%%\n", stats.SuccessRate*100)
   }
   ```

3. **Keep Qdrant Running**
   - Experiences stored in Qdrant persist across sessions
   - Agent remembers what it learned even after restart

4. **Start Simple**
   - Enable learning with defaults
   - Let it run for a while
   - Advanced tuning comes later

## Next Steps

- See `examples/learning_agent/` for complete working example
- Check `pkg/learning/` for implementation details
- Read `TODO.md` for upcoming learning features (error patterns, improvement metrics)
