# Learning Agent Demo

This example demonstrates how the agent learns to improve its performance over time through experience.

## What This Demo Shows

1. **ε-Greedy Learning Algorithm in Action**
   - 90% exploitation: Uses best tool based on past success
   - 10% exploration: Tries random tools to discover new patterns

2. **Performance Improvement Over Time**
   - Phase 1 (Initial): Agent explores different tools
   - Phase 2 (Learning): Agent starts exploiting learned knowledge
   - Phase 3 (Expert): Agent confidently uses best tools

3. **Metrics Tracking**
   - Success rate improvements
   - Latency reduction
   - Tool performance statistics

## Prerequisites

### 1. Qdrant Vector Database
```bash
docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant
```

### 2. Ollama with Required Models
```bash
# Main LLM model
ollama pull qwen2.5:7b

# Embedding model for semantic search
ollama pull nomic-embed-text
```

## Running the Demo

```bash
cd examples/learning_agent
go run main.go
```

## Expected Output

```
🧠 Agent Learning Demo - Watch the Agent Improve Over Time
=============================================================

⚙️  Setting up agent with VectorMemory for learning...
✅ Agent ready: VectorMemory memory, Learning: true

📚 Scenario: Agent learns which tools work best for calculations

Phase 1: Initial Learning (First 5 queries)
-------------------------------------------
[Initial 1/5] Query: Calculate 123 * 456
  ✅ Success: 56088 (latency: 2.3s)
...

Phase 2: Learning in Action (Next 10 queries)
----------------------------------------------
[Learned 1/10] Query: Calculate 234 * 567
  ✅ Success: 132678 (latency: 1.8s)
...

Phase 3: Expert Mode (Final 5 queries)
---------------------------------------
[Expert 1/5] Query: Calculate 111 * 222
  ✅ Success: 24642 (latency: 1.2s)
...

📊 Learning Progress Summary
============================
Phase 1 (Initial):  5 queries, 5 success, avg latency: 2.3s
Phase 2 (Learned):  10 queries, 10 success, avg latency: 1.8s
Phase 3 (Expert):   5 queries, 5 success, avg latency: 1.2s

📈 Latency improvement: 47.8% faster in Expert mode
📈 Overall success rate: 100.0% (20/20)

🔧 Tool Performance Analysis
============================
Calculator (calculation queries):
  Success Rate: 100.0%
  Average Latency: 1.5s
  Total Calls: 20
  Successes: 20, Failures: 0

Recommended tool for 'calculate 100 * 200':
  Tool: calculator
  Confidence: 95.0%
  Reasoning: High success rate from past experiences
  Strategy: learned
  Mode: Exploitation (using learned knowledge)

✨ Demo complete! The agent learned to:
  1. Identify calculation queries faster
  2. Select the best tool (calculator) consistently
  3. Reduce latency through experience
  4. Balance exploration (10%) vs exploitation (90%)
```

## How It Works

### Experience Recording
Every `Chat()` call automatically records:
- Query and detected intent
- Tool(s) used
- Success/failure status
- Latency metrics
- LLM token usage

### Learning Algorithm (ε-Greedy)

```go
// 90% of the time: Use best tool (exploitation)
if random() > 0.1 {
    tool = selectBestTool(experiences)
}
// 10% of the time: Try random tool (exploration)
else {
    tool = selectRandomTool()
}
```

### Tool Scoring
Best tool selected based on composite score:
```
score = 0.7 × success_rate + 0.3 × (1 - normalized_latency)
```

## Customization

### Change Exploration Rate
```go
// In pkg/learning/tool_selector.go
const defaultExplorationRate = 0.15  // 15% exploration
```

### Minimum Sample Size
```go
// Require more experiences before exploiting
const defaultMinSampleSize = 5  // Default: 3
```

### Query Scenarios
Edit the `queries` arrays in `main.go` to test different scenarios:
- File operations
- Web searches
- Code generation
- Mixed query types

## What's Next?

After running this demo, try:

1. **Error Pattern Detection** (Task 7)
   - Cluster similar errors
   - Detect common failure patterns
   - Auto-suggest corrections

2. **Long-term Metrics** (Task 8)
   - Track improvement over days/weeks
   - Export metrics to dashboard
   - A/B test different learning strategies

## Troubleshooting

### "Failed to create VectorMemory"
- Make sure Qdrant is running: `docker ps | grep qdrant`
- Check port 6334 is not in use: `lsof -i :6334`

### Slow Performance
- First few queries are slower (cold start)
- Latency improves as agent learns
- Consider using faster LLM model

### Low Success Rate
- Check Ollama is running: `curl http://localhost:11434/api/tags`
- Verify model is pulled: `ollama list`
- Check logs for specific errors
