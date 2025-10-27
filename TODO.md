# TODO List - Active Tasks

> **Note**: Completed tasks have been moved to [DONE.md](DONE.md)

## Current Status

**Project**: go-llm-agent  
**Version**: v0.4.0-alpha+vector (Auto-Reasoning + Vector Memory COMPLETED)  
**Last Updated**: October 27, 2025  
**Progress**: Phase 1.1 + 2.1 ✅ COMPLETE (55% of v0.4.0)

---

## 🎯 INTELLIGENCE REASSESSMENT - v0.4.0-alpha+vector

> **Critical Update**: Phase 1.1 (ReAct) + Phase 2.1 (Vector Memory) discovered to be 90% complete!  
> **Discovery Date**: October 27, 2025  
> **Assessment**: Infrastructure exists, self-learning capabilities needed

### Current Intelligence Score: 6.8/10 (↑0.8 from baseline)

| Dimension | Before | Current | Target | Status | Priority |
|-----------|--------|---------|--------|--------|----------|
| **Reasoning** | 5.0 | 7.0 | 8.0 | ✅ ReAct + CoT working | **HIGH** (Self-improvement needed) |
| **Memory** | 6.0 | 7.5 | 8.5 | ✅ Vector search working | **HIGH** (Learning persistence needed) |
| **Tools** | 8.5 | 8.5 | 9.0 | 28 tools complete | **MEDIUM** (Need parallel execution) |
| **Learning** | 2.0 | 3.0 | 9.0 | ❌ **CRITICAL GAP** | **🔥 URGENT** |
| **Architecture** | 7.0 | 7.5 | 8.0 | Clean, extensible | **LOW** |
| **Scalability** | 5.5 | 6.5 | 7.5 | Vector DB integrated | **MEDIUM** |
| **OVERALL IQ** | **6.0** | **6.8** | **8.5** | +0.8 improvement | **Gap: -1.7** |

---

## 🚨 CRITICAL FINDING: Self-Learning Missing!

### The Problem

**Current State**: Agent has memory but **DOESN'T LEARN FROM IT**
```go
// What happens now:
1. Agent uses tool incorrectly → Gets error
2. Error saved to memory
3. Next time: Agent makes SAME mistake!
4. No learning loop, no improvement

// Example:
Day 1: agent.Chat("Calculate 2+2")
  → Calls web_fetch instead of math_calculate ❌
  → Error saved to vector memory
  
Day 2: agent.Chat("Calculate 3+3")  
  → Calls web_fetch AGAIN! ❌❌
  → Same mistake repeated!
```

**Root Cause**: No feedback loop connecting memory → reasoning → behavior change

### What's Missing: Self-Learning System

```
┌─────────────────────────────────────────────────────────────┐
│                    SELF-LEARNING LOOP                        │
│                      (NOT IMPLEMENTED)                       │
└─────────────────────────────────────────────────────────────┘

1. EXPERIENCE COLLECTION ❌
   └─ Store: successful actions, failed actions, corrections
   
2. PATTERN RECOGNITION ❌
   └─ Analyze: "When I use tool X for task Y, success rate is Z%"
   
3. STRATEGY ADJUSTMENT ❌
   └─ Update: tool selection priorities based on past success
   
4. VERIFICATION ❌
   └─ Check: "Did my strategy change improve outcomes?"
   
5. CONTINUOUS IMPROVEMENT ❌
   └─ Iterate: Keep learning from new experiences
```

### Intelligence Without Learning = Stagnation

**Analogy**: Current agent is like a student who:
- ✅ Has a textbook (vector memory)
- ✅ Can read and understand (reasoning)
- ✅ Has tools to solve problems (28 tools)
- ❌ **NEVER learns from mistakes**
- ❌ **Makes same errors repeatedly**
- ❌ **Doesn't improve over time**

---

## 🎓 SELF-LEARNING ARCHITECTURE (NEW PRIORITY #1)

### Phase 3: Learning & Adaptation System (v0.4.0-beta)

**Goal**: Agent learns from experience, improves tool selection, reduces errors over time

#### 3.1 Experience Tracking System

**Purpose**: Record outcomes of every action for later analysis

```go
// pkg/learning/experience.go (NEW)

type Experience struct {
    ID          string
    Timestamp   time.Time
    
    // Context
    Query       string
    Intent      string              // What user wanted
    Reasoning   ReasoningPattern    // CoT, ReAct, Simple
    
    // Action
    ToolCalled  string
    Arguments   map[string]interface{}
    
    // Outcome
    Success     bool
    Result      interface{}
    Error       error
    
    // Feedback
    UserFeedback *Feedback         // Optional: "This answer was good/bad"
    Correction   *string           // What should have been done
    
    // Metrics
    LatencyMs    int64
    TokensUsed   int
    Confidence   float64           // Agent's self-assessed confidence
}

type ExperienceStore struct {
    vectorMemory *memory.VectorMemory
    sqliteDB     *sql.DB
}

func (e *ExperienceStore) Record(ctx context.Context, exp Experience) error
func (e *ExperienceStore) Query(ctx context.Context, filters ExperienceFilters) ([]Experience, error)
func (e *ExperienceStore) GetToolSuccessRate(toolName string, intentPattern string) float64
```

**Implementation Tasks**:
- [ ] **Task 3.1.1**: Create experience data structures
- [ ] **Task 3.1.2**: Integrate experience recording into agent.Chat()
- [ ] **Task 3.1.3**: Store experiences in vector memory (semantic search)
- [ ] **Task 3.1.4**: SQLite storage for structured queries
- [ ] **Task 3.1.5**: Add feedback collection API

**Example Usage**:
```go
// Automatic experience recording
exp := learning.Experience{
    Query: "Calculate 15 * 23",
    Intent: "mathematical_calculation",
    ToolCalled: "web_fetch",      // WRONG TOOL
    Success: false,
    Error: errors.New("404 not found"),
}
experienceStore.Record(ctx, exp)

// Later: Query similar experiences
similar := experienceStore.Query(ctx, ExperienceFilters{
    IntentPattern: "mathematical_calculation",
    MinSimilarity: 0.8,
})
// Finds: "Last 5 times for math, web_fetch failed 100%, math_calculate succeeded 100%"
```

---

#### 3.2 Tool Selection Learning

**Purpose**: Learn which tools work best for which queries

```go
// pkg/learning/tool_selector.go (NEW)

type ToolSelector struct {
    experiences  *ExperienceStore
    toolRegistry *tools.Registry
    
    // Learning parameters
    explorationRate float64  // Probability of trying new tools (default: 0.1)
    minConfidence   float64  // Minimum confidence to use learned strategy
}

type ToolRecommendation struct {
    ToolName   string
    Confidence float64  // 0.0 to 1.0
    Reasoning  string   // "Used successfully 15 times for similar queries"
    
    // Supporting evidence
    SuccessRate     float64
    SampleSize      int
    AvgLatency      int64
    AlternativeTools []string  // Other tools that might work
}

func (t *ToolSelector) RecommendTool(ctx context.Context, query string) (*ToolRecommendation, error) {
    // 1. Analyze query intent
    intent := t.analyzeIntent(query)
    
    // 2. Search for similar past experiences
    similar := t.experiences.Query(ctx, ExperienceFilters{
        IntentPattern: intent,
        MinSimilarity: 0.75,
        Limit: 100,
    })
    
    // 3. Calculate success rates per tool
    toolStats := t.calculateToolStats(similar)
    
    // 4. Select best tool based on:
    //    - Success rate (primary)
    //    - Latency (secondary)
    //    - Recency (tertiary)
    best := t.selectBestTool(toolStats)
    
    // 5. Apply exploration: occasionally try different tools
    if rand.Float64() < t.explorationRate {
        return t.explorativeSelection(toolStats, best)
    }
    
    return best, nil
}

func (t *ToolSelector) UpdateFromFeedback(experience Experience, feedback Feedback) {
    // Reinforcement learning: adjust tool selection based on outcomes
    // Increase confidence in successful tools
    // Decrease confidence in failed tools
}
```

**Implementation Tasks**:
- [ ] **Task 3.2.1**: Implement intent analysis (keywords + embeddings)
- [ ] **Task 3.2.2**: Build tool statistics calculator
- [ ] **Task 3.2.3**: Create exploration-exploitation balance (ε-greedy)
- [ ] **Task 3.2.4**: Integrate with agent's tool selection logic
- [ ] **Task 3.2.5**: Add visualization for tool performance over time

**Learning Algorithm**:
```
Initial state: All tools have equal probability (uniform distribution)

After each experience:
  If success:
    confidence[tool][intent] += α * (1 - confidence[tool][intent])
  If failure:
    confidence[tool][intent] -= β * confidence[tool][intent]
    
Where:
  α = learning rate for positive outcomes (default: 0.1)
  β = learning rate for negative outcomes (default: 0.15)
  
Over time:
  Good tools → confidence approaches 1.0
  Bad tools → confidence approaches 0.0
```

---

#### 3.3 Error Pattern Recognition

**Purpose**: Detect recurring errors and prevent them

```go
// pkg/learning/error_patterns.go (NEW)

type ErrorPattern struct {
    ID          string
    Pattern     string              // Regex or semantic pattern
    Description string              // Human-readable description
    
    // Detection
    Occurrences []Experience        // Experiences matching this pattern
    Frequency   int                 // How many times seen
    FirstSeen   time.Time
    LastSeen    time.Time
    
    // Solution
    Correction  string              // How to fix this error
    Prevention  string              // How to avoid this error
    Confidence  float64             // How sure we are about the solution
}

type ErrorAnalyzer struct {
    patterns    []ErrorPattern
    experiences *ExperienceStore
}

func (e *ErrorAnalyzer) DetectPatterns(ctx context.Context) ([]ErrorPattern, error) {
    // 1. Fetch recent failed experiences
    failures := e.experiences.Query(ctx, ExperienceFilters{
        Success: false,
        Limit: 1000,
    })
    
    // 2. Cluster similar errors using vector embeddings
    clusters := e.clusterByError(failures)
    
    // 3. For each cluster, extract common pattern
    patterns := make([]ErrorPattern, 0)
    for _, cluster := range clusters {
        if len(cluster) >= 3 {  // At least 3 occurrences
            pattern := e.extractPattern(cluster)
            patterns = append(patterns, pattern)
        }
    }
    
    return patterns, nil
}

func (e *ErrorAnalyzer) SuggestCorrection(ctx context.Context, error error) (*ErrorPattern, error) {
    // Search for known error patterns matching this error
    // Return correction if found
}
```

**Implementation Tasks**:
- [ ] **Task 3.3.1**: Create error clustering algorithm (vector similarity)
- [ ] **Task 3.3.2**: Build pattern extraction from clusters
- [ ] **Task 3.3.3**: Implement correction suggestion logic
- [ ] **Task 3.3.4**: Add prevention rules to agent's decision making
- [ ] **Task 3.3.5**: Create error pattern dashboard

**Example Error Patterns**:
```
Pattern 1: "Math calculation → web_fetch (100% failure rate)"
  Correction: "Use math_calculate instead"
  Prevention: "For queries with numbers and math operators, prefer math tools"

Pattern 2: "Empty filter → mongodb_delete (safety violation)"
  Correction: "Add explicit filter like {_id: ObjectId('...')}"
  Prevention: "Never allow empty filter in delete operations"

Pattern 3: "Large dataset → math_stats timeout"
  Correction: "Use sampling for datasets > 1000 elements"
  Prevention: "Check array length before calling math_stats"
```

---

#### 3.4 Continuous Improvement Loop

**Purpose**: Close the learning loop, measure improvement over time

```go
// pkg/learning/improvement.go (NEW)

type ImprovementMetrics struct {
    Period         time.Duration
    
    // Performance metrics
    TotalQueries   int
    SuccessRate    float64
    AvgLatency     time.Duration
    ErrorRate      float64
    
    // Learning metrics
    NewPatterns    int          // New error patterns discovered
    ToolChanges    int          // Times tool selection changed based on learning
    Corrections    int          // Times error was auto-corrected
    
    // Comparison with baseline
    BaselineSuccess float64
    Improvement     float64     // % improvement in success rate
}

type LearningMonitor struct {
    experiences  *ExperienceStore
    toolSelector *ToolSelector
    errorAnalyzer *ErrorAnalyzer
}

func (l *LearningMonitor) MeasureImprovement(ctx context.Context, period time.Duration) (*ImprovementMetrics, error) {
    now := time.Now()
    start := now.Add(-period)
    
    // Get experiences from period
    recent := l.experiences.Query(ctx, ExperienceFilters{
        StartTime: start,
        EndTime: now,
    })
    
    // Calculate metrics
    metrics := &ImprovementMetrics{
        Period: period,
        TotalQueries: len(recent),
    }
    
    // Calculate success rate
    successes := 0
    for _, exp := range recent {
        if exp.Success {
            successes++
        }
    }
    metrics.SuccessRate = float64(successes) / float64(len(recent))
    
    // Compare with baseline (first 100 queries)
    baseline := l.calculateBaseline(ctx)
    metrics.BaselineSuccess = baseline
    metrics.Improvement = ((metrics.SuccessRate - baseline) / baseline) * 100
    
    return metrics, nil
}

func (l *LearningMonitor) ReportProgress() string {
    // Generate human-readable progress report
    // "Last week: 500 queries, 92% success (+12% vs baseline)"
    // "Learned 3 new patterns, prevented 15 errors"
    // "Top improvement: math queries (65% → 95%)"
}
```

**Implementation Tasks**:
- [ ] **Task 3.4.1**: Create metrics calculation system
- [ ] **Task 3.4.2**: Build baseline establishment logic
- [ ] **Task 3.4.3**: Implement improvement tracking over time
- [ ] **Task 3.4.4**: Add progress reporting (daily/weekly/monthly)
- [ ] **Task 3.4.5**: Create learning visualization dashboard

**Expected Results After 1 Week**:
```
Before Learning (Day 1):
  Queries: 100
  Success rate: 75%
  Avg latency: 2.5s
  Common errors: Wrong tool selection (15%), Timeout (10%)

After Learning (Day 7):
  Queries: 700
  Success rate: 88% (+17% improvement!)
  Avg latency: 1.8s (-28% faster)
  Common errors: Timeout (5%), Rare edge cases (3%)
  
Learning achievements:
  ✅ Learned to prefer math_calculate for calculations (95% → 100%)
  ✅ Detected 3 error patterns and auto-corrected them
  ✅ Improved tool selection latency by caching recommendations
  ✅ Reduced redundant web_fetch calls by 40%
```

---

#### 3.5 Feedback Integration

**Purpose**: Allow users to teach the agent

```go
// pkg/learning/feedback.go (NEW)

type Feedback struct {
    ExperienceID string
    Rating       FeedbackRating  // Positive, Negative, Neutral
    Comment      string          // Optional explanation
    Correction   *Correction     // What should have been done
    Timestamp    time.Time
}

type FeedbackRating int
const (
    Negative FeedbackRating = -1
    Neutral  FeedbackRating = 0
    Positive FeedbackRating = 1
)

type Correction struct {
    ShouldUseTool   string
    ShouldNotUseTool string
    BetterApproach   string
}

// Agent API extension
func (a *Agent) ChatWithFeedback(ctx context.Context, message string) (*Response, error) {
    resp, err := a.Chat(ctx, message)
    
    // Prompt user for feedback
    fmt.Println("\nWas this answer helpful? (y/n/skip): ")
    // Store feedback and learn from it
    
    return resp, err
}

func (a *Agent) SubmitFeedback(experienceID string, feedback Feedback) error {
    // Update tool selector confidence
    // Update error patterns
    // Trigger immediate learning update
}
```

**Implementation Tasks**:
- [ ] **Task 3.5.1**: Create feedback data structures
- [ ] **Task 3.5.2**: Add feedback API to agent
- [ ] **Task 3.5.3**: Implement feedback processing logic
- [ ] **Task 3.5.4**: Update learning models based on feedback
- [ ] **Task 3.5.5**: Add interactive feedback examples

---

### Self-Learning Timeline & Milestones

**Week 1-2: Experience Tracking** (Tasks 3.1.x)
- [ ] Implement experience data structures
- [ ] Integrate recording into agent
- [ ] Set up dual storage (vector + SQLite)
- **Milestone 1**: Agent records all actions with outcomes

**Week 3-4: Tool Selection Learning** (Tasks 3.2.x)
- [ ] Build intent analysis
- [ ] Implement tool statistics
- [ ] Create exploration-exploitation logic
- **Milestone 2**: Agent learns which tools work best

**Week 5-6: Error Recognition & Prevention** (Tasks 3.3.x)
- [ ] Implement error clustering
- [ ] Build pattern extraction
- [ ] Add prevention rules
- **Milestone 3**: Agent avoids known mistakes

**Week 7-8: Continuous Improvement** (Tasks 3.4.x + 3.5.x)
- [ ] Create metrics tracking
- [ ] Build progress reporting
- [ ] Add feedback integration
- **Milestone 4**: Agent improves measurably over time

**Success Criteria**:
- [ ] Success rate improves by ≥10% after 1 week of usage
- [ ] Tool selection errors decrease by ≥50%
- [ ] Agent detects and prevents ≥3 common error patterns
- [ ] Learning dashboard shows clear improvement trends
- [ ] User feedback integration works in examples

---

## 🧠 v0.4.0 Planning - Intelligence Upgrade (REVISED PRIORITIES)

> **Strategic Update**: Phase 1.1 + 2.1 discovered **90% complete**! Infrastructure exists.  
> **New Critical Gap**: **SELF-LEARNING** (2.0/10) - Agent doesn't improve from experience  
> **Revised Objective**: Add self-learning (Phase 3) → Target IQ **8.5/10**  
> **Target**: January-February 2026

### Revised Priority Order

**URGENT** (Phase 3): Self-Learning System
- Agent currently repeats mistakes
- No improvement over time despite having memory
- Critical for production use

**HIGH** (Phase 1.3-1.4): Planning & Reflection
- Builds on existing ReAct/CoT
- Enhances reasoning quality

**MEDIUM** (Phase 2.2-2.4): Memory Persistence & Management
- Vector memory works but not persistent
- Needs importance scoring and cleanup

---

## ✅ Phase 1.1 + 2.1: COMPLETED (Discovery)

### What Was Discovered (October 27, 2025)

**Expected**: Need to implement from scratch  
**Reality**: Infrastructure 90% complete! Just needed examples and cleanup

**pkg/reasoning/react.go** (426 lines) - ✅ COMPLETE
- ReActStep structures with Thought/Action/Observation/Reflection
- SaveToMemory() integration
- Structured logging
- Used by agent's auto-reasoning system

**pkg/reasoning/cot.go** (344 lines) - ✅ COMPLETE  
- Chain-of-Thought step structures
- SaveToMemory() integration
- Step-by-step reasoning logging
- Auto-selected for math/logic queries

**pkg/memory/vector.go** (471 lines) - ✅ COMPLETE
- Full Qdrant integration
- SearchSemantic() - cosine similarity search
- HybridSearch() - keyword + vector
- GetByCategory() - category filtering
- GetMostImportant() - importance-based retrieval
- Archive(), Export(), GetStats() - management functions

**pkg/memory/embedder.go** (172 lines) - ✅ COMPLETE
- Embedder interface abstraction
- OllamaEmbedder (nomic-embed-text, mxbai-embed-large)
- OpenAIEmbedder (text-embedding-3-small/large)
- Automatic dimensionality detection

**pkg/types/types.go** - ✅ COMPLETE
- AdvancedMemory interface defined
- MessageCategory enum (factual, procedural, reasoning, etc.)
- Clean interface design

**What Was Created**:
- [x] `examples/vector_memory_agent/main.go` (219 lines)
  - 3-phase demo: teach, semantic search, recall
  - Graceful degradation to BufferMemory
  - Tested with Qdrant - **works perfectly!**

**What Was Cleaned Up**:
- [x] Deleted pkg/embedding duplicate
- [x] Removed 3 outdated examples
- [x] Fixed react_with_tools
- [x] All 17 examples build successfully

**Results**:
- ✅ Semantic search working (finds similar conversations by meaning)
- ✅ ReAct steps saved to memory
- ✅ CoT chains stored with embeddings
- ✅ Vector similarity search operational
- ✅ All 28 tools integrated

**Gap**: These features exist but **agent doesn't learn from them**!

---

## 🎯 Phase 1.3-1.4: Task Planning & Self-Reflection (NEXT PRIORITY)

> **Status**: Not started (depends on Phase 3 self-learning)  
> **Rationale**: Planning and reflection are more valuable when agent can learn from them

### 1.3 Task Planning & Decomposition

### 1.3 Task Planning & Decomposition

**What is Planning?**
```
Decompose complex goals into sub-tasks before execution.

Example:
Goal: "Create a comprehensive report on Go performance in 2025"

Without Planning:
  - Agent tries to answer in one shot
  - Misses important aspects
  - Unstructured output

With Planning:
  Plan:
    1. Search for Go 1.25 release notes
    2. Find performance benchmarks
    3. Compare with previous versions
    4. Identify key improvements
    5. Structure as markdown report
    6. Verify all claims with sources
  
  Execution: Execute sub-tasks sequentially
  Result: Comprehensive, well-structured report
```

**Implementation Tasks**:

- [ ] **Task 1.3.1**: Create `pkg/reasoning/planner.go`
  ```go
  type PlanStep struct {
      ID           string
      Description  string
      Dependencies []string  // Which steps must complete first
      Status       string    // pending, in_progress, completed, failed
      Result       interface{}
      Error        error
  }
  
  type Plan struct {
      Goal        string
      Steps       []PlanStep
      CreatedAt   time.Time
      CompletedAt time.Time
  }
  
  type Planner struct {
      agent *agent.Agent
  }
  ```

- [ ] **Task 1.3.2**: Implement goal decomposition
  ```go
  func (p *Planner) DecomposeGoal(ctx context.Context, goal string) (*Plan, error) {
      // Use LLM to break goal into sub-tasks
      // Identify dependencies between tasks
      // Return structured plan
  }
  ```

- [ ] **Task 1.3.3**: Build task executor with dependency tracking
  ```go
  func (p *Planner) ExecutePlan(ctx context.Context, plan *Plan) error {
      // Execute tasks in order respecting dependencies
      // Support parallel execution of independent tasks (future)
      // Handle failures and retries
  }
  ```

- [ ] **Task 1.3.4**: Create progress monitoring
  ```go
  type PlanProgress struct {
      TotalSteps     int
      CompletedSteps int
      CurrentStep    *PlanStep
      Progress       float64  // 0.0 to 1.0
  }
  
  func (p *Planner) GetProgress(plan *Plan) *PlanProgress
  ```

- [ ] **Task 1.3.5**: Integration and examples
  - Add `agent.Plan(ctx, goal)` method
  - Store plans in memory (Phase 2)
  - Create `examples/planning_agent/main.go`

**Expected Improvement**: Reasoning score **7.5/10 → 8/10** (+0.5 points)

**Benefits**:
- ✅ Handle complex multi-step tasks reliably
- ✅ Clear progress tracking
- ✅ Better failure recovery

---

### 1.4 Self-Reflection & Verification

**What is Self-Reflection?**
```
Agent verifies its own answers before returning to user.

Example:
User: "What's the capital of Australia?"

Without Reflection:
  Agent: "Sydney" (WRONG - common mistake)

With Reflection:
  Initial Answer: "Sydney"
  Reflection: "Wait, let me verify. Sydney is the largest city, 
               but I should double-check the capital..."
  Verification: [Calls fact_check or web_search]
  Observation: "Canberra is the capital of Australia"
  Corrected Answer: "The capital of Australia is Canberra, 
                     not Sydney (which is the largest city)"
```

**Implementation Tasks**:

- [ ] **Task 1.4.1**: Create `pkg/reasoning/reflection.go`
  ```go
  type ReflectionCheck struct {
      Question       string
      InitialAnswer  string
      Concerns       []string  // What might be wrong?
      Verifications  []VerificationStep
      FinalAnswer    string
      Confidence     float64   // 0.0 to 1.0
  }
  
  type VerificationStep struct {
      Method    string    // "fact_check", "calculation_verify", etc.
      Query     string
      Result    interface{}
      Passed    bool
  }
  ```

- [ ] **Task 1.4.2**: Implement verification strategies
  ```go
  // Strategy 1: Factual verification
  func (r *Reflector) VerifyFacts(answer string) (bool, error)
  
  // Strategy 2: Calculation re-check
  func (r *Reflector) VerifyCalculation(expression string, result interface{}) (bool, error)
  
  // Strategy 3: Consistency check
  func (r *Reflector) CheckConsistency(answer string, context []Message) (bool, error)
  ```

- [ ] **Task 1.4.3**: Build confidence scoring
  ```go
  func (r *Reflector) CalculateConfidence(answer string, verifications []VerificationStep) float64 {
      // Score based on:
      // - Number of verification methods passed
      // - Consistency with known facts
      // - Complexity of question
  }
  ```

- [ ] **Task 1.4.4**: Integration with Agent
  - Add `agent.ChatWithReflection(ctx, message, minConfidence)` method
  - Automatically verify answers for critical domains (facts, calculations)
  - Store reflection results in memory (Phase 2)

**Expected Improvement**: Overall Quality +15%, Accuracy +20%

**Benefits**:
- ✅ Fewer factual errors
- ✅ Higher answer quality
- ✅ User trust improvement

---

## 🧠 Phase 2: Memory Persistence & Management (v0.4.0-gamma)

> **Status**: Core vector memory ✅ COMPLETE, persistence pending  
> **What exists**: VectorMemory with Qdrant, semantic search, embeddings  
> **What's missing**: Persistence (SQLite), importance scoring, cleanup strategies

### Current Memory State

**What works** (pkg/memory/vector.go - 471 lines):
- ✅ Qdrant vector database integration
- ✅ SearchSemantic() - find similar conversations
- ✅ HybridSearch() - keyword + vector combined
- ✅ GetByCategory() - filter by message type
- ✅ GetMostImportant() - importance-based retrieval
- ✅ Archive(), Export(), GetStats() - management
- ✅ OllamaEmbedder + OpenAIEmbedder working

**What's missing**:
- ❌ Memory not persisted (lost on restart)
- ❌ No SQLite for structured queries
- ❌ No importance scoring algorithm
- ❌ No automatic cleanup strategies
- ❌ No memory size management

### 2.2 Persistent Memory (SQLite + Qdrant Sync)

**Why Persistence Matters**:
```
Current: Agent forgets everything on restart
Problem: 
  - Can't build long-term knowledge
  - User must re-teach context every session
  - No learning from past interactions

With Persistence:
  - Agent remembers user preferences
  - Accumulated knowledge over weeks/months
  - Continuous improvement
```

**Implementation Tasks**:

- [ ] **Task 2.2.1**: Create `pkg/memory/persistent.go`
  ```go
  type PersistentMemory struct {
      db          *sql.DB        // SQLite for structured data
      vector      *VectorMemory  // Qdrant for semantic search
      syncManager *SyncManager   // Keep DB and vector in sync
  }
  ```

- [ ] **Task 2.2.2**: Design database schema
  ```sql
  CREATE TABLE messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      role TEXT NOT NULL,
      content TEXT NOT NULL,
      category TEXT,
      importance REAL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      metadata JSON
  );
  
  CREATE TABLE tool_calls (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      message_id INTEGER REFERENCES messages(id),
      tool_name TEXT NOT NULL,
      arguments JSON NOT NULL,
      result JSON,
      success BOOLEAN,
      created_at TIMESTAMP
  );
  
  CREATE TABLE react_steps (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL,
      iteration INTEGER,
      thought TEXT,
      action TEXT,
      observation TEXT,
      reflection TEXT,
      created_at TIMESTAMP
  );
  
  CREATE TABLE plans (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      goal TEXT NOT NULL,
      steps JSON NOT NULL,
      status TEXT,
      created_at TIMESTAMP,
      completed_at TIMESTAMP
  );
  
  CREATE INDEX idx_messages_category ON messages(category);
  CREATE INDEX idx_messages_created_at ON messages(created_at);
  CREATE INDEX idx_react_steps_session ON react_steps(session_id);
  ```

- [ ] **Task 2.2.3**: Implement sync mechanism (DB ↔ Vector DB)
  ```go
  type SyncManager struct {
      db     *sql.DB
      vector *VectorMemory
  }
  
  func (s *SyncManager) SyncMessage(ctx context.Context, message Message) error {
      // 1. Insert to SQLite (structured storage)
      // 2. Generate embedding
      // 3. Insert to Qdrant (semantic search)
      // 4. Transaction: rollback both on failure
  }
  ```

- [ ] **Task 2.2.4**: Add memory statistics and management
  ```go
  type MemoryStats struct {
      TotalMessages    int
      TotalReActSteps  int
      TotalPlans       int
      OldestMessage    time.Time
      DatabaseSize     int64  // bytes
      VectorCount      int
  }
  
  func (p *PersistentMemory) GetStats() (*MemoryStats, error)
  func (p *PersistentMemory) Compact(olderThan time.Duration) error  // Archive old data
  func (p *PersistentMemory) Export(path string) error  // Backup
  ```

**Expected Improvement**: Scalability score **5.5/10 → 7.5/10** (+2 points)

---

### 2.3 Importance-Based Memory Management

**The Problem**:
```
Not all memories are equal:
  High importance: "User's name is John, prefers Python"
  Low importance: "Weather check at 2pm on random Tuesday"

Current FIFO: Deletes important memories first (oldest)
Needed: Keep important, archive unimportant
```

**Implementation Tasks**:

- [ ] **Task 2.3.1**: Create importance scoring system
  ```go
  type ImportanceScorer struct {
      factors map[string]float64  // Factor → weight
  }
  
  func (i *ImportanceScorer) Score(message Message) float64 {
      score := 0.0
      
      // Factor 1: User preferences (+0.3)
      if containsPreference(message.Content) {
          score += 0.3
      }
      
      // Factor 2: Successful tool use (+0.2)
      if hasSuccessfulToolCalls(message) {
          score += 0.2
      }
      
      // Factor 3: ReAct/CoT reasoning (+0.25)
      if isReasoningStep(message) {
          score += 0.25
      }
      
      // Factor 4: User feedback (+0.15)
      if hasPositiveFeedback(message) {
          score += 0.15
      }
      
      // Factor 5: Recency (+0.1)
      score += recencyBonus(message.CreatedAt)
      
      return score  // 0.0 to 1.0
  }
  ```

- [ ] **Task 2.3.2**: Implement smart cleanup strategy
  ```go
  func (p *PersistentMemory) CleanupLowImportance(ctx context.Context, threshold float64) error {
      // 1. Score all messages
      // 2. Archive messages with score < threshold
      // 3. Keep in vector DB for semantic search
      // 4. Remove from hot cache
  }
  ```

- [ ] **Task 2.3.3**: Add memory priorities
  ```go
  type MemoryPriority int
  const (
      PriorityCritical  MemoryPriority = 4  // Never delete (user prefs)
      PriorityHigh      MemoryPriority = 3  // Keep for weeks
      PriorityMedium    MemoryPriority = 2  // Keep for days
      PriorityLow       MemoryPriority = 1  // Keep for hours
      PriorityTransient MemoryPriority = 0  // Delete after session
  )
  ```

**Expected Improvement**: Memory quality +30%, Storage efficiency +50%

---

### 2.4 Memory-Reasoning Integration

**Critical Synergy**:

- [ ] **Task 2.4.1**: Store ReAct steps in vector memory
  ```go
  func (v *VectorMemory) AddReActStep(ctx context.Context, step ReActStep) error {
      // Store step with embedding
      // Tag with category = "reasoning"
      // Index by thought/action/reflection
  }
  
  func (v *VectorMemory) FindSimilarReasoning(ctx context.Context, query string) ([]ReActStep, error) {
      // Semantic search for similar thought processes
  }
  ```

- [ ] **Task 2.4.2**: Index CoT chains for reuse
  ```go
  func (v *VectorMemory) StoreCoTChain(ctx context.Context, steps []CoTStep, result string) error
  
  func (v *VectorMemory) FindSimilarProblems(ctx context.Context, problem string) ([]CoTChain, error) {
      // Find similar step-by-step solutions
  }
  ```

- [ ] **Task 2.4.3**: Plan library with semantic search
  ```go
  func (v *VectorMemory) StorePlan(ctx context.Context, plan Plan) error
  
  func (v *VectorMemory) FindSimilarPlans(ctx context.Context, goal string) ([]Plan, error) {
      // Discover past solutions to similar goals
  }
  ```

- [ ] **Task 2.4.4**: Reflection history for learning
  ```go
  type ReflectionHistory struct {
      OriginalAnswer  string
      Correction      string
      Reason          string
      Confidence      float64
  }
  
  func (v *VectorMemory) LearnFromReflections(ctx context.Context) error {
      // Analyze patterns in past corrections
      // Improve future answers
  }
  ```

---

## 📊 Implementation Timeline & Milestones

### Week 1-2: Foundation (Phase 1.1 + 2.1 Core) ✅ COMPLETED
- [x] Intelligence analysis (COMPLETED)
- [x] Design types and interfaces (pkg/types - AdvancedMemory interface EXISTS)
- [x] ReAct pattern prototype (pkg/reasoning/react.go - 426 lines EXISTS)
- [x] Vector memory prototype with Qdrant (pkg/memory/vector.go - 471 lines EXISTS)
- [x] Basic embedding integration (pkg/memory/embedder.go - 172 lines EXISTS)

**Milestone 1**: ✅ ACHIEVED - Can do explicit reasoning with semantic memory search

### Week 3-4: Core Features (Phase 1.2 + 1.3 + 2.2)
- [x] Chain-of-Thought implementation (pkg/reasoning/cot.go - 344 lines EXISTS)
- [ ] Task planning system
- [ ] SQLite persistence layer
- [ ] DB ↔ Vector sync mechanism

**Milestone 2**: CoT complete, planning & persistence pending

### Week 5-6: Quality & Integration (Phase 1.4 + 2.3 + 2.4)
- [ ] Self-reflection system
- [ ] Importance scoring
- [ ] Memory-reasoning integration
- [ ] Comprehensive testing

**Milestone 3**: Full intelligence upgrade with persistent learning

### Week 7-8: Polish & Release (v0.4.0)
- [ ] Performance optimization
- [ ] Documentation (README, guides, examples)
- [ ] Example agents (ReAct, Planning, Learning)
- [ ] Benchmarks and comparisons
- [ ] Release v0.4.0

**Final Milestone**: IQ 8.0/10, Production-ready intelligent agent

---

## 🎯 Success Criteria for v0.4.0

### Quantitative Metrics
- [ ] Overall IQ: 6.0 → 8.0+ (at least +2 points)
- [ ] Reasoning score: 5.0 → 8.0+ (+3 points)
- [ ] Memory score: 6.0 → 8.0+ (+2 points)
- [ ] Test coverage: Maintain 80%+
- [ ] Performance: <10% overhead vs v0.3.0

### Qualitative Metrics
- [ ] Can solve complex 5+ step tasks reliably
- [ ] Explains reasoning clearly in logs
- [ ] Remembers context across sessions
- [ ] Self-corrects obvious errors
- [ ] Finds relevant past solutions

### User Experience
- [ ] Transparent thinking process (users can debug)
- [ ] Better answer quality (fewer errors)
- [ ] Faster responses (memory reuse)
- [ ] Continuous learning (improves over time)

---

## 🎯 v0.3.0 Planning - Advanced Tools

### Phase 2: Vector Database & Data Tools (Target: Nov-Dec 2025)
- [ ] qdrant_connect - Connect to Qdrant vector DB
- [ ] qdrant_create_collection - Create vector collection
- [ ] qdrant_upsert - Insert/update vectors
- [ ] qdrant_search - Semantic vector search
- [ ] qdrant_delete - Delete vectors

**Data Processing Tools (3 tools)** - Priority: MEDIUM
- [ ] data_json - JSON parsing and manipulation
- [ ] data_csv - CSV read/write/transform
- [ ] data_xml - XML parsing

**Status**: MongoDB tools COMPLETED, Qdrant tools in research phase (see RESEARCH_NEW_TOOLS.md)

---

## 📋 Documentation Updates Needed

- ✅ Update README.md with Network & Gmail tools examples (COMPLETED)
- ✅ Update CHANGELOG.md with Network & Gmail tools (COMPLETED)
- ✅ Update DONE.md and TODO.md with current status (COMPLETED Oct 27)
- [ ] Add Network & Gmail tools to BUILTIN_TOOLS_DESIGN.md
- [ ] Create Qdrant design document
- [ ] Add MongoDB connection pooling best practices doc
- [ ] Gmail OAuth2 setup video tutorial (optional)

---

## 🚀 Release Planning

### v0.2.0 Release - ✅ RELEASED
**Status**: Released to GitHub  
**Release Date**: October 26, 2025  
**Features**:
- Core agent framework
- 3 LLM providers (Ollama, OpenAI, Gemini)
- Memory management (Buffer + Vector with Qdrant)
- Auto-reasoning system (CoT/ReAct/Simple)
- 28 built-in tools (8 categories)
- Comprehensive examples (17 examples)

### v0.3.0 Release - IN PROGRESS (85% Complete)
**Target**: December 2025  
**Features**:
- ✅ Math tools (2 tools - COMPLETED Oct 27)
- ✅ MongoDB tools (5 tools - COMPLETED Oct 27)
- ✅ Network tools (5 tools - COMPLETED Oct 27)
- ✅ Gmail tools (4 tools - COMPLETED Oct 27)
- [ ] Qdrant vector search (5 tools - Planned)
- [ ] Data processing tools (3 tools - Planned)
- Current: 28 tools (24 auto-loaded + 4 Gmail opt-in) | Target: 40+ built-in tools total

### v0.4.0 Release - PLANNING (Self-Learning Focus)
**Target**: February 2026  
**Current Progress**: Phase 1.1 + 2.1 ✅ COMPLETED (infrastructure discovered)  
**Critical Gap**: Self-Learning System (Phase 3 - URGENT)  
**Features**:
- Phase 3: Experience tracking, tool selection learning, error patterns
- Phase 1.3-1.4: Task planning, self-reflection
- Phase 2.2-2.4: Memory persistence (SQLite), importance scoring
- Expected IQ improvement: 6.8 → 8.5 (+1.7 points)

---

## 🔄 Ongoing Maintenance

### Testing
- ✅ Maintain 80%+ code coverage
- ✅ All CI/CD pipelines green
- ✅ MongoDB tools: 7 test functions passing
- [ ] Add MongoDB integration tests with testcontainers
- [ ] Add Qdrant integration tests

### Performance
- ✅ Math tools tested with professional libraries
- [ ] Benchmark MongoDB connection pooling
- [ ] Optimize stat calculations for large datasets (>10k elements)
- [ ] Add caching for repeated calculations

### Security
- ✅ Expression evaluation safety (whitelist approach)
- ✅ MongoDB empty filter prevention (delete safety)
- ✅ Connection pool limits (max 10 connections)
- ✅ Network tools: DNS server validation, SSL verification
- ✅ Gmail tools: OAuth2 credential protection, token caching
- [ ] MongoDB connection string sanitization
- [ ] Qdrant API key management
- [ ] Rate limiting for database operations

---

## 📝 Notes

- **Professional Libraries Used**:
  - govaluate v3.0.0 (4.3k stars) - Expression evaluation
  - gonum v0.16.0 (7.2k stars) - Statistical operations
  - mongo-driver v1.17.4 (Official MongoDB Go driver)
  - miekg/dns v1.1.68 (Professional DNS library)
  - go-ping/ping v1.2.0 (ICMP ping)
  - likexian/whois v1.15.6 + whois-parser v1.24.20 (WHOIS queries)
  - oschwald/geoip2-golang v1.13.0 (IP geolocation)
  - google.golang.org/api v0.253.0 (Official Google Gmail API)
- **Current Status**: 28 tools registered in builtin package
- **Tool Categories**: 8 categories (File, Web, DateTime, System, Math, Database, Network, Email)
- **Safety**: 19/28 tools are safe (68% read-only operations)
- **Auto-loaded**: 24 tools (File, Web, DateTime, System, Math, Database, Network)
- **Opt-in**: 4 Gmail tools (requires OAuth2 credentials setup)
- **Examples**: 9 comprehensive demos with real-world use cases
- **Next Focus**: Self-Learning System (Phase 3) - CRITICAL PRIORITY

---

## 📊 SUMMARY: Current State & Next Steps

### What We Have (v0.4.0-alpha+vector)

✅ **Auto-Reasoning** (Phase 1.1)
- ReActAgent with Thought/Action/Observation/Reflection
- CoTAgent with step-by-step reasoning
- Automatic pattern selection (CoT/ReAct/Simple)
- 25 builtin tools auto-loaded

✅ **Vector Memory** (Phase 2.1)
- Qdrant integration with semantic search
- Ollama + OpenAI embedders
- Hybrid search (keyword + vector)
- Category-based filtering
- Importance-based retrieval

✅ **28 Builtin Tools** (v0.3.0)
- File, Web, DateTime, System (13 tools)
- Math, MongoDB, Network, Gmail (15 tools)
- Professional libraries (goquery, govaluate, gonum, etc.)
- Auto-loaded by default (24) + opt-in Gmail (4)

### 🚨 Critical Gap: Self-Learning

❌ **Agent repeats mistakes**
- Uses wrong tools despite past failures
- No learning from experience
- Success rate doesn't improve over time

❌ **Memory without learning**
- Stores conversations but doesn't analyze them
- Can search past experiences but doesn't use insights
- Vector database underutilized

### 🎯 Next Priority: Phase 3 (Self-Learning System)

**Why urgent**: Production agents MUST improve over time

**Timeline**: 8 weeks (Jan-Feb 2026)

**Expected impact**: 
- Success rate: 75% → 88% (+17%)
- Tool selection errors: -50%
- Learning IQ: 2.0 → 8.0 (+6.0 points!)
- Overall IQ: 6.8 → 8.5 (+1.7 points)

**Key components**:
1. Experience tracking (record all actions + outcomes)
2. Tool selection learning (ε-greedy exploration)
3. Error pattern recognition (clustering + prevention)
4. Continuous improvement (metrics + feedback)
5. User feedback integration (human-in-the-loop)

### 🗺️ Revised Roadmap

**Phase 3** (Weeks 1-8): Self-Learning System → IQ 8.0
- Experience tracking infrastructure
- Tool selection learning algorithm
- Error pattern recognition
- Continuous improvement metrics

**Phase 1.3-1.4** (Weeks 9-12): Planning + Reflection → IQ 8.3
- Task decomposition and planning
- Self-verification and correction
- Quality improvements

**Phase 2.2-2.4** (Weeks 13-16): Memory Persistence → IQ 8.5
- SQLite + Qdrant sync
- Importance-based cleanup
- Long-term memory management

**v0.4.0 Release** (Week 17): Full Intelligence Upgrade
- Production-ready self-learning agent
- Comprehensive documentation
- Performance benchmarks

**Success Metric**: Agent demonstrably learns from experience and improves over 1 week of use.

---

**Last Updated**: October 27, 2025  
**Next Milestone**: Phase 3.1 - Experience Tracking System

