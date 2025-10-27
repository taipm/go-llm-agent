package types

import "time"

// ===========================
// ReAct Pattern Types
// ===========================

// ReActStep represents one iteration in the ReAct (Reasoning + Acting) loop
type ReActStep struct {
	Iteration   int       `json:"iteration"`
	Thought     string    `json:"thought"`     // What am I thinking?
	Action      string    `json:"action"`      // What tool should I call?
	Observation string    `json:"observation"` // What did I observe from the action?
	Reflection  string    `json:"reflection"`  // What did I learn?
	Timestamp   time.Time `json:"timestamp"`
}

// ===========================
// Chain-of-Thought Types
// ===========================

// CoTStep represents one step in a chain-of-thought reasoning process
type CoTStep struct {
	StepNumber  int         `json:"step_number"`
	Description string      `json:"description"` // "Calculate speed in km/h"
	Reasoning   string      `json:"reasoning"`   // "Using formula: distance / time"
	Result      interface{} `json:"result"`      // Intermediate result
}

// CoTChain represents a complete chain-of-thought reasoning process
type CoTChain struct {
	Query      string    `json:"query"`       // Original question
	Steps      []CoTStep `json:"steps"`       // Step-by-step reasoning
	Answer     string    `json:"answer"`      // Final answer
	Confidence float64   `json:"confidence"`  // 0.0 to 1.0
	CreatedAt  time.Time `json:"created_at"`
}

// ===========================
// Task Planning Types
// ===========================

// PlanStep represents a single step in a task plan
type PlanStep struct {
	ID           string        `json:"id"`
	Description  string        `json:"description"`
	Dependencies []string      `json:"dependencies"` // IDs of steps that must complete first
	Status       PlanStatus    `json:"status"`
	Result       interface{}   `json:"result,omitempty"`
	Error        error         `json:"error,omitempty"`
	StartedAt    time.Time     `json:"started_at,omitempty"`
	CompletedAt  time.Time     `json:"completed_at,omitempty"`
}

// PlanStatus represents the status of a plan step
type PlanStatus string

const (
	PlanStatusPending    PlanStatus = "pending"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusFailed     PlanStatus = "failed"
	PlanStatusSkipped    PlanStatus = "skipped"
)

// Plan represents a complete task decomposition plan
type Plan struct {
	ID          string     `json:"id"`
	Goal        string     `json:"goal"`        // High-level goal
	Steps       []PlanStep `json:"steps"`       // Ordered list of steps
	Status      PlanStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   time.Time  `json:"started_at,omitempty"`
	CompletedAt time.Time  `json:"completed_at,omitempty"`
}

// PlanProgress tracks execution progress of a plan
type PlanProgress struct {
	TotalSteps     int       `json:"total_steps"`
	CompletedSteps int       `json:"completed_steps"`
	FailedSteps    int       `json:"failed_steps"`
	CurrentStep    *PlanStep `json:"current_step,omitempty"`
	Progress       float64   `json:"progress"` // 0.0 to 1.0
}

// ===========================
// Self-Reflection Types
// ===========================

// ReflectionCheck represents a verification/reflection process
type ReflectionCheck struct {
	Question       string              `json:"question"`
	InitialAnswer  string              `json:"initial_answer"`
	Concerns       []string            `json:"concerns"`      // What might be wrong?
	Verifications  []VerificationStep  `json:"verifications"` // Verification attempts
	FinalAnswer    string              `json:"final_answer"`
	Confidence     float64             `json:"confidence"` // 0.0 to 1.0
	WasCorrected   bool                `json:"was_corrected"`
	CreatedAt      time.Time           `json:"created_at"`
}

// VerificationStep represents one verification attempt
type VerificationStep struct {
	Method   string      `json:"method"` // "fact_check", "calculation_verify", "consistency_check"
	Query    string      `json:"query"`
	Result   interface{} `json:"result"`
	Passed   bool        `json:"passed"`
	Error    error       `json:"error,omitempty"`
}

// ===========================
// Memory Category Types
// ===========================

// MessageCategory categorizes memory entries for better organization
type MessageCategory string

const (
	CategoryFactual    MessageCategory = "factual"    // Facts, definitions, data
	CategoryProcedural MessageCategory = "procedural" // How-to, tutorials, processes
	CategoryReasoning  MessageCategory = "reasoning"  // ReAct steps, thought processes
	CategoryPlanning   MessageCategory = "planning"   // Plans, strategies, goals
	CategoryReflection MessageCategory = "reflection" // Verifications, corrections, learning
	CategoryTool       MessageCategory = "tool"       // Tool calls and results
	CategoryUser       MessageCategory = "user"       // User preferences, context
	CategoryGeneral    MessageCategory = "general"    // Uncategorized
)

// MemoryPriority determines how long memories should be retained
type MemoryPriority int

const (
	PriorityTransient MemoryPriority = 0 // Delete after session
	PriorityLow       MemoryPriority = 1 // Keep for hours
	PriorityMedium    MemoryPriority = 2 // Keep for days
	PriorityHigh      MemoryPriority = 3 // Keep for weeks
	PriorityCritical  MemoryPriority = 4 // Never delete (user preferences, important facts)
)

// MessageMetadata extends Message with reasoning-specific metadata
type MessageMetadata struct {
	Category       MessageCategory `json:"category,omitempty"`
	Priority       MemoryPriority  `json:"priority,omitempty"`
	Importance     float64         `json:"importance,omitempty"`    // 0.0 to 1.0
	ReActStep      *ReActStep      `json:"react_step,omitempty"`
	CoTChain       *CoTChain       `json:"cot_chain,omitempty"`
	Plan           *Plan           `json:"plan,omitempty"`
	Reflection     *ReflectionCheck `json:"reflection,omitempty"`
	Embedding      []float32       `json:"embedding,omitempty"`     // Vector embedding
	RelatedIDs     []string        `json:"related_ids,omitempty"`   // Related message IDs
}

// ===========================
// Memory Statistics
// ===========================

// MemoryStats provides statistics about memory usage
type MemoryStats struct {
	TotalMessages    int       `json:"total_messages"`
	TotalReActSteps  int       `json:"total_react_steps"`
	TotalCoTChains   int       `json:"total_cot_chains"`
	TotalPlans       int       `json:"total_plans"`
	TotalReflections int       `json:"total_reflections"`
	OldestMessage    time.Time `json:"oldest_message"`
	NewestMessage    time.Time `json:"newest_message"`
	DatabaseSize     int64     `json:"database_size"` // bytes
	VectorCount      int       `json:"vector_count"`
}
