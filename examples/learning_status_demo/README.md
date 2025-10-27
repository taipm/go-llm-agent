# Learning Status Demo

Demonstrates the enhanced `agent.Status()` API that shows detailed learning progress and intelligence metrics.

## What This Demo Shows

This example demonstrates how to track an agent's learning progress using the enhanced `Status()` method:

1. **Learning Stages**: `exploring` → `learning` → `expert`
2. **Performance Metrics**: Success rate, production readiness
3. **Tool Analysis**: Top performing vs problematic tools
4. **Knowledge Areas**: What the agent has learned (by intent)
5. **Recent Improvements**: Actual progress tracking

## Features Demonstrated

### Agent Status Information

```go
status := ag.Status()

// Learning progress
status.Learning.LearningStage        // "exploring", "learning", "expert"
status.Learning.TotalExperiences     // Number of interactions
status.Learning.OverallSuccessRate   // Success percentage
status.Learning.ReadyForProduction   // true if success >= 85% and exp >= 10

// Tool performance analysis
status.Learning.TopPerformingTools   // Tools with >= 90% success
status.Learning.ProblematicTools     // Tools with < 50% success

// Knowledge tracking
status.Learning.KnowledgeAreas       // Intent → experience count
status.Learning.RecentImprovements   // Recent progress indicators
```

### Learning Stages

- **Exploring** (< 5 experiences): Agent is discovering capabilities
- **Learning** (5-19 experiences): Active learning phase
- **Expert** (>= 20 experiences): Proficient with accumulated knowledge

### Production Readiness

Agent is considered production-ready when:
- Overall success rate >= 85%
- Total experiences >= 10

## Running the Demo

```bash
# Start Qdrant for full learning
docker run -p 6334:6334 -p 6333:6333 qdrant/qdrant

# Run demo
go run examples/learning_status_demo/main.go
```

## What to Observe

The demo runs 10 tasks and shows status updates after every 3 tasks:

1. **Initial State**: No experiences, empty knowledge
2. **After 3 Tasks**: Exploring stage, initial patterns
3. **After 6 Tasks**: Learning stage, tool performance emerging
4. **After 9 Tasks**: More knowledge areas identified
5. **Final State**: Expert stage (if >= 20 total experiences), production ready

## Example Output

```
📊 AGENT STATUS - After 6 Tasks
======================================================================
🧠 Learning Stage:      learning
📚 Total Experiences:   6
✅ Success Rate:        83.3%
🚀 Production Ready:    false

⭐ Top Performing Tools:
   math_calculate (100%)
   file_write (100%)

📖 Knowledge Areas:
   calculation: 3 experiences
   file_operation: 2 experiences
   system: 1 experiences

📈 Recent Improvements:
   Recent success rate (100%) is higher than overall (83%)
```

## Key Insights

1. **Self-Monitoring**: Agent tracks its own learning automatically
2. **Transparency**: Clear visibility into what agent knows
3. **Intelligence Growth**: Observable improvement over time
4. **Tool Mastery**: Identifies which tools it's good/bad at
5. **Knowledge Map**: Shows expertise areas

## Integration

Use this in production to:
- Monitor agent learning progress
- Decide when agent is ready for production
- Identify areas needing improvement
- Track knowledge accumulation
- Debug tool selection issues

```go
// Check if agent is production ready
status := agent.Status()
if status.Learning.ReadyForProduction {
    log.Println("Agent is production ready!")
    log.Printf("Success rate: %.1f%%", status.Learning.OverallSuccessRate)
} else {
    log.Printf("Agent still learning (%s stage)", status.Learning.LearningStage)
}
```
