# Chain-of-Thought Reasoning Example

This example demonstrates **Chain-of-Thought (CoT)** reasoning, where the LLM breaks down complex problems into step-by-step reasoning before providing a final answer.

## What is Chain-of-Thought?

Chain-of-Thought prompting encourages the model to show its work by thinking through problems step by step, similar to how humans solve complex problems.

### Benefits:
- ✅ **Fewer errors**: Step-by-step reasoning reduces calculation mistakes
- ✅ **Transparent thinking**: See exactly how the agent arrived at an answer
- ✅ **Better explanations**: Each step is clearly explained
- ✅ **Verifiable logic**: Easy to check if reasoning is correct
- ✅ **Complex problem solving**: Handles multi-step math, logic puzzles, word problems

### Example:

**Without CoT:**
```
Q: If a train travels 120km in 1.5 hours, what's its speed in mph?
A: 50 mph
```

**With CoT:**
```
Q: If a train travels 120km in 1.5 hours, what's its speed in mph?

Step 1: Calculate speed in km/h = 120 / 1.5 = 80 km/h
Step 2: Convert km to miles: 1 km = 0.621371 miles
Step 3: Calculate mph = 80 * 0.621371 = 49.7 mph

Answer: Approximately 50 mph
```

## Prerequisites

- Go 1.21 or later
- Ollama running locally with a model (e.g., `qwen3:1.7b`)
- Environment variables configured (see below)

## Quick Start

### 1. Setup Environment

Create a `.env` file in the parent examples directory:

```bash
LLM_PROVIDER=ollama
LLM_MODEL=qwen3:1.7b
OLLAMA_BASE_URL=http://localhost:11434
```

Or export environment variables:

```bash
export LLM_PROVIDER=ollama
export LLM_MODEL=qwen3:1.7b
export OLLAMA_BASE_URL=http://localhost:11434
```

### 2. Run the Example

```bash
cd examples/cot_reasoning
go run main.go
```

## What This Example Does

The example tests Chain-of-Thought reasoning on 4 different types of problems:

1. **Math Problem**: Unit conversion with calculations
2. **Logic Puzzle**: Age-based algebraic reasoning
3. **Word Problem**: Percentage and reverse calculation
4. **Complex Calculation**: Compound interest formula

For each problem, the agent:
1. Breaks down the problem into steps
2. Shows reasoning for each step
3. Provides a final answer
4. Validates the logical soundness of the reasoning chain

## Expected Output

```
=== Chain-of-Thought Reasoning Example ===

📡 Connecting to LLM...
✅ LLM Provider ready

🧠 Creating Chain-of-Thought Agent...
✅ CoT Agent ready

======================================================================
Question 1: Math Problem
======================================================================

📝 If a train travels 120 kilometers in 1.5 hours, what is its speed in miles per hour?

=== Chain-of-Thought Reasoning ===

Question: If a train travels 120 kilometers in 1.5 hours...

Step 1: Calculate speed in km/h = 120 / 1.5 = 80 km/h
Step 2: Convert to miles: 80 km × 0.621371 = 49.70968 mph
Step 3: Round to reasonable precision = 49.7 mph

✅ Final Answer: Approximately 49.7 mph

⏱️  Time taken: 2.3s

✅ Reasoning is logically sound

📊 Final Answer: Approximately 49.7 mph
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LLM_PROVIDER` | LLM provider (ollama, openai, gemini) | - | Yes |
| `LLM_MODEL` | Model name (e.g., qwen3:1.7b, gpt-4) | - | Yes |
| `OLLAMA_BASE_URL` | Ollama API base URL | http://localhost:11434 | For Ollama |
| `OPENAI_API_KEY` | OpenAI API key | - | For OpenAI |
| `GEMINI_API_KEY` | Google Gemini API key | - | For Gemini |

## How Chain-of-Thought Works

### 1. Prompt Engineering

The CoT agent uses a special prompt template that instructs the LLM to:
- Think step by step
- Show calculations and reasoning
- Use a structured format (Step 1, Step 2, etc.)
- Provide a clear final answer

### 2. Step Parsing

The agent parses the LLM response to extract:
- Individual reasoning steps
- Step numbers
- Descriptions and calculations
- Final answer

### 3. Validation

The agent validates the reasoning chain by checking:
- All steps are numbered correctly
- No steps are empty
- A final answer is provided
- Logical flow is maintained

## Code Structure

```go
// Create CoT agent
agent := reasoning.NewCoTAgent(llm, memory, maxSteps)
agent.SetVerbose(true)

// Think through a problem
answer, err := agent.Think(ctx, "Your question here")

// Get detailed reasoning history
history := agent.GetReasoningHistory()

// Validate reasoning
valid, issues := agent.Validate()
```

## CoT vs ReAct

| Feature | Chain-of-Thought | ReAct |
|---------|-----------------|-------|
| **Purpose** | Step-by-step reasoning | Action-based reasoning |
| **Steps** | Thought → Thought → Answer | Thought → Action → Observation |
| **Tool Use** | No | Yes |
| **Best For** | Math, logic, explanations | Multi-step tasks with tools |
| **Complexity** | Lower | Higher |

## Troubleshooting

### Model Doesn't Follow Format

Some smaller models may not consistently follow the Step 1/Step 2 format. Solutions:
- Use a larger or more capable model (qwen2.5:3b, qwen2.5:7b)
- Check the parsing logic handles numbered lists (1., 2., etc.)
- Enable verbose mode to see raw LLM responses

### Validation Fails

If validation reports issues:
- Check if the model provided a final answer
- Verify step numbering is sequential
- Enable verbose logging to see what was parsed

### Poor Reasoning Quality

To improve reasoning:
- Use a more capable model
- Adjust maxSteps if problems are too complex
- Add domain-specific few-shot examples to the prompt

## Next Steps

- Try **ReAct pattern** (`examples/react_memory/`) for action-based reasoning with tool use
- Explore **Vector Memory** for semantic search over reasoning history
- Combine CoT with persistent memory for learning from past reasoning

## Related Examples

- `examples/react_memory/` - ReAct pattern with memory
- `examples/simple/` - Basic agent usage
- `examples/agent_with_logging/` - Agent with structured logging

## References

- [Chain-of-Thought Prompting Paper](https://arxiv.org/abs/2201.11903)
- [Large Language Models are Zero-Shot Reasoners](https://arxiv.org/abs/2205.11916)
