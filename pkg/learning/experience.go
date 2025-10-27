package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// Experience represents a single interaction with the agent, including context,
// action taken, and outcome. This data is used for learning and improvement.
type Experience struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`

	// Context: What was the situation?
	Query          string                 `json:"query"`           // Original user query
	Intent         string                 `json:"intent"`          // Detected intent (e.g., "calculation", "web_search")
	ReasoningMode  string                 `json:"reasoning_mode"`  // CoT, ReAct, Simple
	ConversationID string                 `json:"conversation_id"` // Session identifier
	Metadata       map[string]interface{} `json:"metadata"`        // Additional context

	// Action: What did the agent do?
	ToolCalled string                 `json:"tool_called,omitempty"` // Tool name if any
	Arguments  map[string]interface{} `json:"arguments,omitempty"`   // Tool arguments
	Response   string                 `json:"response"`              // Agent's response

	// Outcome: What was the result?
	Success      bool        `json:"success"`              // Did it succeed?
	Error        string      `json:"error,omitempty"`      // Error message if failed
	ErrorType    string      `json:"error_type,omitempty"` // Error category
	Result       interface{} `json:"result,omitempty"`     // Tool result if successful
	Confidence   float64     `json:"confidence"`           // Self-assessed confidence (0.0-1.0)
	WasReflected bool        `json:"was_reflected"`        // Was reflection applied?
	WasCorrected bool        `json:"was_corrected"`        // Was answer corrected after reflection?

	// Feedback: What did we learn?
	UserFeedback *Feedback `json:"user_feedback,omitempty"` // Optional user feedback
	Correction   string    `json:"correction,omitempty"`    // What should have been done

	// Metrics: Performance data
	LatencyMs  int64 `json:"latency_ms"`  // Total response time
	TokensUsed int   `json:"tokens_used"` // LLM tokens consumed
}

// Feedback represents user feedback on an experience
type Feedback struct {
	Rating    FeedbackRating `json:"rating"`            // Positive, Negative, Neutral
	Comment   string         `json:"comment,omitempty"` // Optional explanation
	Timestamp time.Time      `json:"timestamp"`         // When feedback was given
	Helpful   bool           `json:"helpful"`           // Was the answer helpful?
	Accurate  bool           `json:"accurate"`          // Was the answer accurate?
	Complete  bool           `json:"complete"`          // Was the answer complete?
}

// FeedbackRating represents the quality of an experience
type FeedbackRating int

const (
	FeedbackNegative FeedbackRating = -1
	FeedbackNeutral  FeedbackRating = 0
	FeedbackPositive FeedbackRating = 1
)

// ExperienceFilters defines criteria for querying experiences
type ExperienceFilters struct {
	// Time filters
	StartTime time.Time
	EndTime   time.Time

	// Context filters
	Query          string  // Semantic search by query
	Intent         string  // Filter by intent
	ReasoningMode  string  // CoT, ReAct, Simple
	ConversationID string  // Filter by session
	MinSimilarity  float64 // Minimum similarity for semantic search (0.0-1.0)

	// Outcome filters
	Success   *bool  // Filter by success/failure (nil = both)
	ToolUsed  string // Filter by tool name
	ErrorType string // Filter by error type

	// Quality filters
	MinConfidence float64 // Minimum confidence score
	WithFeedback  bool    // Only experiences with user feedback

	// Pagination
	Limit  int // Maximum results to return
	Offset int // Skip first N results
}

// ExperienceStore manages storage and retrieval of experiences for learning
type ExperienceStore struct {
	memory types.AdvancedMemory // Vector memory for semantic search
	// TODO: Add SQLite for structured queries in Phase 2.2
}

// NewExperienceStore creates a new experience store
func NewExperienceStore(memory types.AdvancedMemory) *ExperienceStore {
	return &ExperienceStore{
		memory: memory,
	}
}

// Record stores a new experience in the store
func (e *ExperienceStore) Record(ctx context.Context, exp Experience) error {
	if exp.ID == "" {
		return fmt.Errorf("experience ID is required")
	}

	// Serialize experience to JSON for storage
	data, err := json.Marshal(exp)
	if err != nil {
		return fmt.Errorf("failed to serialize experience: %w", err)
	}

	// Create message for vector storage
	msg := types.Message{
		Role:    types.RoleAssistant,
		Content: string(data),
		Metadata: map[string]interface{}{
			"category": types.CategoryExperience,
			"exp_id":   exp.ID,
			"intent":   exp.Intent,
			"success":  exp.Success,
		},
	}

	// Store in vector memory for semantic search
	if err := e.memory.Add(msg); err != nil {
		return fmt.Errorf("failed to store experience: %w", err)
	}

	return nil
}

// Query retrieves experiences matching the given filters
func (e *ExperienceStore) Query(ctx context.Context, filters ExperienceFilters) ([]Experience, error) {
	var results []Experience

	// If semantic search by query
	if filters.Query != "" {
		limit := filters.Limit
		if limit == 0 {
			limit = 10 // Default limit
		}

		// Search by semantic similarity
		messages, err := e.memory.SearchSemantic(ctx, filters.Query, limit)
		if err != nil {
			return nil, fmt.Errorf("semantic search failed: %w", err)
		}

		// Parse messages into experiences
		for _, msg := range messages {
			var exp Experience
			if err := json.Unmarshal([]byte(msg.Content), &exp); err != nil {
				continue // Skip invalid entries
			}

			// Apply additional filters
			if e.matchesFilters(exp, filters) {
				results = append(results, exp)
			}
		}
	} else {
		// TODO: Implement non-semantic queries using SQLite in Phase 2.2
		// For now, return error if no query provided
		return nil, fmt.Errorf("non-semantic queries not yet implemented")
	}

	// Apply limit
	if filters.Limit > 0 && len(results) > filters.Limit {
		results = results[:filters.Limit]
	}

	return results, nil
}

// matchesFilters checks if an experience matches the given filters
func (e *ExperienceStore) matchesFilters(exp Experience, filters ExperienceFilters) bool {
	// Time filters
	if !filters.StartTime.IsZero() && exp.Timestamp.Before(filters.StartTime) {
		return false
	}
	if !filters.EndTime.IsZero() && exp.Timestamp.After(filters.EndTime) {
		return false
	}

	// Context filters
	if filters.Intent != "" && exp.Intent != filters.Intent {
		return false
	}
	if filters.ReasoningMode != "" && exp.ReasoningMode != filters.ReasoningMode {
		return false
	}
	if filters.ConversationID != "" && exp.ConversationID != filters.ConversationID {
		return false
	}

	// Outcome filters
	if filters.Success != nil && exp.Success != *filters.Success {
		return false
	}
	if filters.ToolUsed != "" && exp.ToolCalled != filters.ToolUsed {
		return false
	}
	if filters.ErrorType != "" && exp.ErrorType != filters.ErrorType {
		return false
	}

	// Quality filters
	if exp.Confidence < filters.MinConfidence {
		return false
	}
	if filters.WithFeedback && exp.UserFeedback == nil {
		return false
	}

	return true
}

// GetToolSuccessRate calculates the success rate of a tool for a given intent pattern
func (e *ExperienceStore) GetToolSuccessRate(ctx context.Context, toolName string, intentPattern string) (float64, int, error) {
	// Query experiences using this tool with similar intent
	filters := ExperienceFilters{
		Query:         intentPattern,
		ToolUsed:      toolName,
		MinSimilarity: 0.7,
		Limit:         100, // Look at last 100 uses
	}

	experiences, err := e.Query(ctx, filters)
	if err != nil {
		return 0, 0, err
	}

	if len(experiences) == 0 {
		return 0, 0, nil // No data
	}

	// Calculate success rate
	successes := 0
	for _, exp := range experiences {
		if exp.Success {
			successes++
		}
	}

	successRate := float64(successes) / float64(len(experiences))
	return successRate, len(experiences), nil
}

// GetStats returns statistics about stored experiences
func (e *ExperienceStore) GetStats(ctx context.Context) (*ExperienceStats, error) {
	// TODO: Implement using SQLite queries in Phase 2.2
	// For now, return basic stats
	return &ExperienceStats{
		TotalExperiences: 0, // Placeholder
		SuccessRate:      0.0,
	}, nil
}

// ExperienceStats provides overview statistics
type ExperienceStats struct {
	TotalExperiences   int            `json:"total_experiences"`
	SuccessRate        float64        `json:"success_rate"`
	ToolUsageCount     map[string]int `json:"tool_usage_count"`
	IntentDistribution map[string]int `json:"intent_distribution"`
	AvgLatencyMs       int64          `json:"avg_latency_ms"`
	AvgConfidence      float64        `json:"avg_confidence"`
}
