package learning

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/tools"
)

// ToolSelector learns which tools work best for different query types
// using ε-greedy exploration-exploitation strategy
type ToolSelector struct {
	experiences  *ExperienceStore
	toolRegistry *tools.Registry
	logger       logger.Logger

	// Learning parameters
	explorationRate float64 // Probability of trying random tool (default: 0.1)
	minConfidence   float64 // Minimum confidence to use learned strategy (default: 0.6)
	minSampleSize   int     // Minimum experiences needed for reliable recommendation (default: 3)

	// Random number generator
	rng *rand.Rand
}

// ToolRecommendation contains a tool recommendation with supporting evidence
type ToolRecommendation struct {
	ToolName   string  `json:"tool_name"`
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
	Reasoning  string  `json:"reasoning"`  // Human-readable explanation

	// Supporting evidence
	SuccessRate      float64  `json:"success_rate"`      // Percentage 0.0 to 1.0
	SampleSize       int      `json:"sample_size"`       // Number of past experiences
	AvgLatencyMs     int64    `json:"avg_latency_ms"`    // Average response time
	AlternativeTools []string `json:"alternative_tools"` // Other viable options

	// Decision metadata
	IsExploration    bool   `json:"is_exploration"`    // Was this an exploratory choice?
	DecisionStrategy string `json:"decision_strategy"` // "learned", "exploration", "fallback"
}

// NewToolSelector creates a new tool selector with learning capabilities
func NewToolSelector(experiences *ExperienceStore, registry *tools.Registry, log logger.Logger) *ToolSelector {
	return &ToolSelector{
		experiences:     experiences,
		toolRegistry:    registry,
		logger:          log,
		explorationRate: 0.1, // 10% exploration
		minConfidence:   0.6, // Require 60% confidence
		minSampleSize:   3,   // Need at least 3 samples
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SetExplorationRate sets the exploration rate (0.0 to 1.0)
func (t *ToolSelector) SetExplorationRate(rate float64) {
	if rate < 0.0 {
		rate = 0.0
	}
	if rate > 1.0 {
		rate = 1.0
	}
	t.explorationRate = rate
}

// SetMinConfidence sets the minimum confidence threshold
func (t *ToolSelector) SetMinConfidence(threshold float64) {
	t.minConfidence = threshold
}

// RecommendTool recommends the best tool for a given query based on past experiences
func (t *ToolSelector) RecommendTool(ctx context.Context, query string, intent string) (*ToolRecommendation, error) {
	// Step 1: Decide exploration vs exploitation
	shouldExplore := t.rng.Float64() < t.explorationRate

	if shouldExplore {
		t.logger.Debug("🎲 Exploration mode: trying random tool selection")
		return t.exploratorySelection(intent)
	}

	// Step 2: Query similar past experiences
	filters := ExperienceFilters{
		Query:         query,
		Intent:        intent,
		MinSimilarity: 0.7,
		Limit:         100, // Look at last 100 similar queries
	}

	similarExperiences, err := t.experiences.Query(ctx, filters)
	if err != nil {
		t.logger.Debug("Failed to query experiences: %v", err)
		return t.fallbackSelection(intent, "query_failed")
	}

	// Step 3: No past experiences - use fallback
	if len(similarExperiences) == 0 {
		t.logger.Debug("No similar experiences found, using fallback")
		return t.fallbackSelection(intent, "no_data")
	}

	// Step 4: Calculate tool statistics
	toolStats := t.calculateToolStats(similarExperiences)

	// Step 5: Select best tool based on success rate and latency
	bestTool := t.selectBestTool(toolStats)

	if bestTool == nil {
		return t.fallbackSelection(intent, "no_valid_tools")
	}

	// Step 6: Check if we have enough confidence
	if bestTool.Confidence < t.minConfidence {
		t.logger.Debug("Low confidence (%.2f < %.2f), using fallback", bestTool.Confidence, t.minConfidence)
		return t.fallbackSelection(intent, "low_confidence")
	}

	bestTool.DecisionStrategy = "learned"
	bestTool.IsExploration = false

	return bestTool, nil
}

// calculateToolStats computes statistics for each tool from experiences
func (t *ToolSelector) calculateToolStats(experiences []Experience) map[string]*ToolStats {
	stats := make(map[string]*ToolStats)

	for _, exp := range experiences {
		if exp.ToolCalled == "" {
			continue // Skip experiences without tool usage
		}

		tool := exp.ToolCalled
		if stats[tool] == nil {
			stats[tool] = &ToolStats{
				ToolName:   tool,
				TotalCalls: 0,
				Successes:  0,
				Failures:   0,
				Latencies:  []int64{},
			}
		}

		stats[tool].TotalCalls++
		if exp.Success {
			stats[tool].Successes++
		} else {
			stats[tool].Failures++
		}

		if exp.LatencyMs > 0 {
			stats[tool].Latencies = append(stats[tool].Latencies, exp.LatencyMs)
		}
	}

	// Calculate derived metrics
	for _, stat := range stats {
		if stat.TotalCalls > 0 {
			stat.SuccessRate = float64(stat.Successes) / float64(stat.TotalCalls)
		}

		if len(stat.Latencies) > 0 {
			sum := int64(0)
			for _, lat := range stat.Latencies {
				sum += lat
			}
			stat.AvgLatency = sum / int64(len(stat.Latencies))
		}
	}

	return stats
}

// selectBestTool selects the tool with best performance metrics
func (t *ToolSelector) selectBestTool(toolStats map[string]*ToolStats) *ToolRecommendation {
	if len(toolStats) == 0 {
		return nil
	}

	// Create list of candidates
	type candidate struct {
		tool  string
		stats *ToolStats
		score float64
	}

	candidates := make([]candidate, 0)

	for toolName, stats := range toolStats {
		// Skip tools with insufficient data
		if stats.TotalCalls < t.minSampleSize {
			continue
		}

		// Calculate composite score:
		// - Success rate (70% weight)
		// - Inverse latency (30% weight, normalized)
		successScore := stats.SuccessRate * 0.7

		latencyScore := 0.0
		if stats.AvgLatency > 0 {
			// Normalize latency: faster = better
			// Assume 5000ms is worst case, 100ms is best case
			normalizedLatency := 1.0 - (float64(stats.AvgLatency) / 5000.0)
			if normalizedLatency < 0 {
				normalizedLatency = 0
			}
			latencyScore = normalizedLatency * 0.3
		}

		score := successScore + latencyScore

		candidates = append(candidates, candidate{
			tool:  toolName,
			stats: stats,
			score: score,
		})
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	best := candidates[0]

	// Collect alternatives (top 3)
	alternatives := make([]string, 0)
	for i := 1; i < len(candidates) && i < 4; i++ {
		alternatives = append(alternatives, candidates[i].tool)
	}

	// Build recommendation
	rec := &ToolRecommendation{
		ToolName:         best.tool,
		Confidence:       best.score, // Composite score as confidence
		SuccessRate:      best.stats.SuccessRate,
		SampleSize:       best.stats.TotalCalls,
		AvgLatencyMs:     best.stats.AvgLatency,
		AlternativeTools: alternatives,
		IsExploration:    false,
		DecisionStrategy: "learned",
	}

	// Generate reasoning
	rec.Reasoning = fmt.Sprintf(
		"Used successfully %d/%d times (%.0f%%) with avg latency %dms",
		best.stats.Successes,
		best.stats.TotalCalls,
		best.stats.SuccessRate*100,
		best.stats.AvgLatency,
	)

	return rec
}

// exploratorySelection picks a random tool to explore new possibilities
func (t *ToolSelector) exploratorySelection(intent string) (*ToolRecommendation, error) {
	// Get all available tools
	allTools := t.toolRegistry.All()
	if len(allTools) == 0 {
		return nil, fmt.Errorf("no tools available")
	}

	// Pick random tool
	randomIdx := t.rng.Intn(len(allTools))
	randomTool := allTools[randomIdx]

	return &ToolRecommendation{
		ToolName:         randomTool.Name(),
		Confidence:       0.5, // Medium confidence for exploration
		Reasoning:        "Exploratory selection to discover new tool usage patterns",
		SuccessRate:      0.0,
		SampleSize:       0,
		AvgLatencyMs:     0,
		AlternativeTools: []string{},
		IsExploration:    true,
		DecisionStrategy: "exploration",
	}, nil
}

// fallbackSelection provides a reasonable default when learning data is insufficient
func (t *ToolSelector) fallbackSelection(intent string, reason string) (*ToolRecommendation, error) {
	// Intent-based heuristics for fallback
	var toolName string

	switch intent {
	case "calculation":
		toolName = t.findToolByName("math_calculate")
	case "information_retrieval":
		toolName = t.findToolByName("web_search")
	case "file_operation":
		toolName = t.findToolByName("file_read")
	default:
		// Pick first available tool as ultimate fallback
		allTools := t.toolRegistry.All()
		if len(allTools) > 0 {
			toolName = allTools[0].Name()
		}
	}

	if toolName == "" {
		return nil, fmt.Errorf("no suitable tool found for intent: %s", intent)
	}

	return &ToolRecommendation{
		ToolName:         toolName,
		Confidence:       0.3, // Low confidence for fallback
		Reasoning:        fmt.Sprintf("Fallback selection (reason: %s, intent: %s)", reason, intent),
		SuccessRate:      0.0,
		SampleSize:       0,
		AvgLatencyMs:     0,
		AlternativeTools: []string{},
		IsExploration:    false,
		DecisionStrategy: "fallback",
	}, nil
}

// findToolByName finds a tool by name (case-insensitive, partial match)
func (t *ToolSelector) findToolByName(namePattern string) string {
	allTools := t.toolRegistry.All()
	for _, tool := range allTools {
		if strings.Contains(strings.ToLower(tool.Name()), strings.ToLower(namePattern)) {
			return tool.Name()
		}
	}
	return ""
}

// ToolStats holds statistics about a tool's performance
type ToolStats struct {
	ToolName    string
	TotalCalls  int
	Successes   int
	Failures    int
	SuccessRate float64
	Latencies   []int64
	AvgLatency  int64
}

// GetToolStats returns statistics for a specific tool
func (t *ToolSelector) GetToolStats(ctx context.Context, toolName string, intent string) (*ToolStats, error) {
	// Query experiences for this tool
	filters := ExperienceFilters{
		ToolUsed: toolName,
		Intent:   intent,
		Limit:    1000,
	}

	// For now, we need to use semantic search with a generic query
	// TODO: Add non-semantic filtering in Phase 2.2 with SQLite
	filters.Query = intent + " " + toolName

	experiences, err := t.experiences.Query(ctx, filters)
	if err != nil {
		return nil, err
	}

	stats := &ToolStats{
		ToolName:   toolName,
		TotalCalls: len(experiences),
		Latencies:  make([]int64, 0),
	}

	for _, exp := range experiences {
		if exp.Success {
			stats.Successes++
		} else {
			stats.Failures++
		}

		if exp.LatencyMs > 0 {
			stats.Latencies = append(stats.Latencies, exp.LatencyMs)
		}
	}

	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(stats.Successes) / float64(stats.TotalCalls)
	}

	if len(stats.Latencies) > 0 {
		sum := int64(0)
		for _, lat := range stats.Latencies {
			sum += lat
		}
		stats.AvgLatency = sum / int64(len(stats.Latencies))
	}

	return stats, nil
}
