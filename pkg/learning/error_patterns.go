package learning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/logger"
)

// ErrorPattern represents a recurring error pattern detected from experiences
type ErrorPattern struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`     // Human-readable pattern description
	Description string    `json:"description"` // Detailed explanation
	ErrorType   string    `json:"error_type"`  // Category of error
	Frequency   int       `json:"frequency"`   // Number of occurrences
	FirstSeen   time.Time `json:"first_seen"`  // When first detected
	LastSeen    time.Time `json:"last_seen"`   // Most recent occurrence

	// Pattern characteristics
	CommonQuery   string   `json:"common_query"`   // Representative query
	FailedTools   []string `json:"failed_tools"`   // Tools that failed
	CommonIntents []string `json:"common_intents"` // Related intents
	ErrorMessages []string `json:"error_messages"` // Common error messages
	AvgConfidence float64  `json:"avg_confidence"` // Average confidence when failing
	SuccessRate   float64  `json:"success_rate"`   // How often this pattern succeeds

	// Solutions
	Correction string  `json:"correction"` // How to fix this error
	Prevention string  `json:"prevention"` // How to avoid this error
	BestTool   string  `json:"best_tool"`  // Tool that works for similar queries
	Confidence float64 `json:"confidence"` // Confidence in this pattern (0.0-1.0)

	// Evidence
	ExperienceIDs []string `json:"experience_ids"` // IDs of experiences in this pattern
}

// ErrorCluster represents a cluster of similar errors
type ErrorCluster struct {
	Centroid    []float64    // Vector representation of cluster center
	Experiences []Experience // Experiences in this cluster
	Similarity  float64      // Average similarity within cluster
	Size        int          // Number of experiences
}

// ErrorAnalyzer detects recurring error patterns and suggests corrections
type ErrorAnalyzer struct {
	experiences *ExperienceStore
	logger      logger.Logger

	// Detection parameters
	minClusterSize   int     // Minimum experiences to form a pattern (default: 3)
	similarityThresh float64 // Minimum similarity to group errors (default: 0.75)
	minConfidence    float64 // Minimum confidence for recommendations (default: 0.6)
	maxPatterns      int     // Maximum patterns to track (default: 100)

	// Cached patterns
	patterns []ErrorPattern
	lastScan time.Time
}

// NewErrorAnalyzer creates a new error pattern analyzer
func NewErrorAnalyzer(experiences *ExperienceStore, log logger.Logger) *ErrorAnalyzer {
	return &ErrorAnalyzer{
		experiences:      experiences,
		logger:           log,
		minClusterSize:   3,
		similarityThresh: 0.75,
		minConfidence:    0.6,
		maxPatterns:      100,
		patterns:         make([]ErrorPattern, 0),
	}
}

// DetectPatterns analyzes failed experiences to find recurring error patterns
func (e *ErrorAnalyzer) DetectPatterns(ctx context.Context) ([]ErrorPattern, error) {
	e.logger.Info("🔍 Starting error pattern detection...")

	// Step 1: Query all failed experiences using efficient category-based retrieval
	failed, err := e.experiences.GetAllFailures(ctx, 500) // Analyze last 500 failures
	if err != nil {
		return nil, fmt.Errorf("failed to query experiences: %w", err)
	}

	e.logger.Debug("Found %d failed experiences to analyze", len(failed))

	if len(failed) < e.minClusterSize {
		e.logger.Debug("Not enough failures to detect patterns (need at least %d)", e.minClusterSize)
		return []ErrorPattern{}, nil
	}

	// Step 2: Cluster similar errors
	clusters := e.clusterErrors(ctx, failed)
	e.logger.Debug("Detected %d error clusters", len(clusters))

	// Step 3: Extract patterns from clusters
	patterns := make([]ErrorPattern, 0)

	for i, cluster := range clusters {
		// Only create patterns for clusters meeting minimum size
		if cluster.Size < e.minClusterSize {
			continue
		}

		pattern := e.extractPattern(ctx, cluster, i)
		if pattern != nil {
			patterns = append(patterns, *pattern)
		}
	}

	e.logger.Info("✅ Detected %d error patterns", len(patterns))

	// Step 4: Sort by frequency (most common first)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Frequency > patterns[j].Frequency
	})

	// Limit number of patterns
	if len(patterns) > e.maxPatterns {
		patterns = patterns[:e.maxPatterns]
	}

	// Cache patterns
	e.patterns = patterns
	e.lastScan = time.Now()

	return patterns, nil
}

// clusterErrors groups similar errors using semantic similarity
func (e *ErrorAnalyzer) clusterErrors(ctx context.Context, failures []Experience) []ErrorCluster {
	if len(failures) == 0 {
		return []ErrorCluster{}
	}

	clusters := make([]ErrorCluster, 0)
	used := make(map[int]bool) // Track which experiences are already clustered

	// Simple clustering algorithm:
	// For each failure, find all similar failures and group them
	for i, exp := range failures {
		if used[i] {
			continue // Already in a cluster
		}

		cluster := ErrorCluster{
			Experiences: []Experience{exp},
			Size:        1,
		}
		used[i] = true

		// Find similar failures
		filters := ExperienceFilters{
			Query:         exp.Query,
			MinSimilarity: e.similarityThresh,
			Limit:         50,
		}

		similar, err := e.experiences.Query(ctx, filters)
		if err != nil {
			e.logger.Debug("Failed to find similar experiences: %v", err)
			continue
		}

		// Add similar failures to cluster
		for _, sim := range similar {
			// Only add if it's actually a failure and not already used
			if !sim.Success {
				// Check if this experience is in our failures list
				for j, fail := range failures {
					if !used[j] && fail.ID == sim.ID {
						cluster.Experiences = append(cluster.Experiences, sim)
						cluster.Size++
						used[j] = true
						break
					}
				}
			}
		}

		// Calculate average similarity within cluster
		if cluster.Size > 1 {
			cluster.Similarity = e.calculateClusterSimilarity(cluster.Experiences)
			clusters = append(clusters, cluster)
		}
	}

	return clusters
}

// calculateClusterSimilarity estimates average similarity within a cluster
func (e *ErrorAnalyzer) calculateClusterSimilarity(experiences []Experience) float64 {
	if len(experiences) <= 1 {
		return 1.0
	}

	// Heuristic: if they share the same error type or tool, they're more similar
	errorTypes := make(map[string]int)
	tools := make(map[string]int)

	for _, exp := range experiences {
		if exp.ErrorType != "" {
			errorTypes[exp.ErrorType]++
		}
		if exp.ToolCalled != "" {
			tools[exp.ToolCalled]++
		}
	}

	// Calculate similarity based on commonality
	maxErrorCount := 0
	for _, count := range errorTypes {
		if count > maxErrorCount {
			maxErrorCount = count
		}
	}

	maxToolCount := 0
	for _, count := range tools {
		if count > maxToolCount {
			maxToolCount = count
		}
	}

	errorSimilarity := float64(maxErrorCount) / float64(len(experiences))
	toolSimilarity := float64(maxToolCount) / float64(len(experiences))

	// Average of error and tool similarity
	return (errorSimilarity + toolSimilarity) / 2.0
}

// extractPattern creates an error pattern from a cluster
func (e *ErrorAnalyzer) extractPattern(ctx context.Context, cluster ErrorCluster, index int) *ErrorPattern {
	if cluster.Size == 0 {
		return nil
	}

	// Aggregate cluster data
	errorTypes := make(map[string]int)
	tools := make(map[string]int)
	intents := make(map[string]int)
	errorMsgs := make([]string, 0)
	experienceIDs := make([]string, 0)
	totalConfidence := 0.0
	var firstSeen, lastSeen time.Time

	for i, exp := range cluster.Experiences {
		if exp.ErrorType != "" {
			errorTypes[exp.ErrorType]++
		}
		if exp.ToolCalled != "" {
			tools[exp.ToolCalled]++
		}
		if exp.Intent != "" {
			intents[exp.Intent]++
		}
		if exp.Error != "" && len(errorMsgs) < 5 {
			errorMsgs = append(errorMsgs, exp.Error)
		}

		experienceIDs = append(experienceIDs, exp.ID)
		totalConfidence += exp.Confidence

		if i == 0 || exp.Timestamp.Before(firstSeen) {
			firstSeen = exp.Timestamp
		}
		if i == 0 || exp.Timestamp.After(lastSeen) {
			lastSeen = exp.Timestamp
		}
	}

	// Find most common error type
	commonErrorType := e.mostCommon(errorTypes)

	// Find most common failed tools
	failedTools := e.topN(tools, 3)

	// Find most common intents
	commonIntents := e.topN(intents, 3)

	// Calculate average confidence
	avgConfidence := totalConfidence / float64(cluster.Size)

	// Generate pattern ID
	patternID := fmt.Sprintf("pattern_%d_%d", time.Now().Unix(), index)

	// Create representative query (use first experience)
	commonQuery := cluster.Experiences[0].Query

	// Generate pattern description
	description := e.generateDescription(commonErrorType, failedTools, commonIntents, cluster.Size)

	// Find correction (look for successful similar queries)
	correction, bestTool := e.findCorrection(ctx, cluster.Experiences[0])

	pattern := &ErrorPattern{
		ID:            patternID,
		Pattern:       commonErrorType,
		Description:   description,
		ErrorType:     commonErrorType,
		Frequency:     cluster.Size,
		FirstSeen:     firstSeen,
		LastSeen:      lastSeen,
		CommonQuery:   commonQuery,
		FailedTools:   failedTools,
		CommonIntents: commonIntents,
		ErrorMessages: errorMsgs,
		AvgConfidence: avgConfidence,
		SuccessRate:   0.0, // Failed pattern
		Correction:    correction,
		Prevention:    e.generatePrevention(commonErrorType, failedTools),
		BestTool:      bestTool,
		Confidence:    e.calculatePatternConfidence(cluster),
		ExperienceIDs: experienceIDs,
	}

	return pattern
}

// mostCommon returns the most common string in a map
func (e *ErrorAnalyzer) mostCommon(items map[string]int) string {
	maxCount := 0
	common := ""

	for item, count := range items {
		if count > maxCount {
			maxCount = count
			common = item
		}
	}

	return common
}

// topN returns the top N items from a frequency map
func (e *ErrorAnalyzer) topN(items map[string]int, n int) []string {
	type pair struct {
		key   string
		value int
	}

	pairs := make([]pair, 0)
	for k, v := range items {
		pairs = append(pairs, pair{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	result := make([]string, 0)
	for i := 0; i < len(pairs) && i < n; i++ {
		result = append(result, pairs[i].key)
	}

	return result
}

// generateDescription creates a human-readable pattern description
func (e *ErrorAnalyzer) generateDescription(errorType string, tools []string, intents []string, frequency int) string {
	parts := make([]string, 0)

	if errorType != "" {
		parts = append(parts, fmt.Sprintf("'%s' errors", errorType))
	} else {
		parts = append(parts, "Errors")
	}

	if len(tools) > 0 {
		parts = append(parts, fmt.Sprintf("when using %s", strings.Join(tools, ", ")))
	}

	if len(intents) > 0 {
		parts = append(parts, fmt.Sprintf("for %s tasks", strings.Join(intents, ", ")))
	}

	parts = append(parts, fmt.Sprintf("(occurred %d times)", frequency))

	return strings.Join(parts, " ")
}

// findCorrection looks for successful similar queries to suggest a correction
func (e *ErrorAnalyzer) findCorrection(ctx context.Context, failedExp Experience) (string, string) {
	// Search for successful experiences with similar query
	filters := ExperienceFilters{
		Query:         failedExp.Query,
		MinSimilarity: 0.7,
		Limit:         20,
	}

	successful := []Experience{}
	successTrue := true
	filters.Success = &successTrue

	similar, err := e.experiences.Query(ctx, filters)
	if err != nil {
		return "No correction available - no similar successful queries found", ""
	}

	// Filter for actual successes (since we can't filter directly yet)
	for _, exp := range similar {
		if exp.Success {
			successful = append(successful, exp)
		}
	}

	if len(successful) == 0 {
		return "No correction available - no similar successful queries found", ""
	}

	// Find most common successful tool
	toolCounts := make(map[string]int)
	for _, exp := range successful {
		if exp.ToolCalled != "" {
			toolCounts[exp.ToolCalled]++
		}
	}

	bestTool := e.mostCommon(toolCounts)

	correction := fmt.Sprintf(
		"Try using '%s' tool instead (succeeded %d/%d times for similar queries)",
		bestTool,
		len(successful),
		len(similar),
	)

	return correction, bestTool
}

// generatePrevention creates prevention advice
func (e *ErrorAnalyzer) generatePrevention(errorType string, failedTools []string) string {
	prevention := make([]string, 0)

	if errorType != "" {
		switch strings.ToLower(errorType) {
		case "tool_not_found":
			prevention = append(prevention, "Verify tool availability before use")
		case "invalid_arguments":
			prevention = append(prevention, "Validate arguments before calling tool")
		case "timeout":
			prevention = append(prevention, "Use tools with better performance characteristics")
		case "api_error":
			prevention = append(prevention, "Add retry logic and error handling")
		default:
			prevention = append(prevention, "Add error handling for this error type")
		}
	}

	if len(failedTools) > 0 {
		prevention = append(prevention,
			fmt.Sprintf("Avoid using %s for this type of query", strings.Join(failedTools, ", ")))
	}

	if len(prevention) == 0 {
		return "Review error logs and adjust tool selection strategy"
	}

	return strings.Join(prevention, "; ")
}

// calculatePatternConfidence determines confidence in this pattern
func (e *ErrorAnalyzer) calculatePatternConfidence(cluster ErrorCluster) float64 {
	// Confidence based on:
	// - Cluster size (more occurrences = more confident)
	// - Cluster similarity (tighter cluster = more confident)

	sizeScore := math.Min(float64(cluster.Size)/10.0, 1.0) // Max at 10 occurrences
	similarityScore := cluster.Similarity

	confidence := (sizeScore*0.6 + similarityScore*0.4)

	return confidence
}

// SuggestCorrection finds the best correction for a given error
func (e *ErrorAnalyzer) SuggestCorrection(ctx context.Context, query string, errorMsg string) (*ErrorPattern, error) {
	e.logger.Debug("🔧 Looking for correction suggestion for query: %s", query)

	// If we don't have recent patterns, detect them
	if len(e.patterns) == 0 || time.Since(e.lastScan) > 5*time.Minute {
		patterns, err := e.DetectPatterns(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to detect patterns: %w", err)
		}
		e.patterns = patterns
	}

	// Try to match query to existing patterns
	bestMatch := e.findBestMatchingPattern(query, errorMsg)

	if bestMatch != nil && bestMatch.Confidence >= e.minConfidence {
		e.logger.Info("✅ Found matching error pattern: %s (confidence: %.2f)",
			bestMatch.Description, bestMatch.Confidence)
		return bestMatch, nil
	}

	// No matching pattern - try to find ad-hoc correction
	correction, bestTool := e.findCorrection(ctx, Experience{
		Query: query,
		Error: errorMsg,
	})

	adhocPattern := &ErrorPattern{
		ID:          fmt.Sprintf("adhoc_%d", time.Now().Unix()),
		Pattern:     "Unknown pattern",
		Description: "No known pattern, using ad-hoc correction",
		Correction:  correction,
		BestTool:    bestTool,
		Confidence:  0.3, // Low confidence for ad-hoc
	}

	e.logger.Debug("No matching pattern found, suggesting ad-hoc correction")
	return adhocPattern, nil
}

// findBestMatchingPattern finds the pattern that best matches a query
func (e *ErrorAnalyzer) findBestMatchingPattern(query string, errorMsg string) *ErrorPattern {
	if len(e.patterns) == 0 {
		return nil
	}

	var bestMatch *ErrorPattern
	bestScore := 0.0

	for i := range e.patterns {
		pattern := &e.patterns[i]
		score := e.scorePatternMatch(pattern, query, errorMsg)

		if score > bestScore {
			bestScore = score
			bestMatch = pattern
		}
	}

	// Only return if score is above threshold
	if bestScore < 0.5 {
		return nil
	}

	return bestMatch
}

// scorePatternMatch calculates how well a pattern matches a query
func (e *ErrorAnalyzer) scorePatternMatch(pattern *ErrorPattern, query string, errorMsg string) float64 {
	score := 0.0

	// Simple text similarity (could be improved with embeddings)
	queryLower := strings.ToLower(query)
	patternQueryLower := strings.ToLower(pattern.CommonQuery)

	// Check for common words
	queryWords := strings.Fields(queryLower)
	patternWords := strings.Fields(patternQueryLower)

	commonWords := 0
	for _, qw := range queryWords {
		for _, pw := range patternWords {
			if qw == pw {
				commonWords++
				break
			}
		}
	}

	if len(queryWords) > 0 {
		score += float64(commonWords) / float64(len(queryWords)) * 0.5
	}

	// Check error message similarity
	if errorMsg != "" && len(pattern.ErrorMessages) > 0 {
		errorLower := strings.ToLower(errorMsg)
		for _, patternErr := range pattern.ErrorMessages {
			if strings.Contains(errorLower, strings.ToLower(patternErr)) ||
				strings.Contains(strings.ToLower(patternErr), errorLower) {
				score += 0.3
				break
			}
		}
	}

	// Add pattern confidence
	score += pattern.Confidence * 0.2

	return math.Min(score, 1.0)
}

// GetPatterns returns all detected patterns
func (e *ErrorAnalyzer) GetPatterns() []ErrorPattern {
	return e.patterns
}

// GetPatternStats returns statistics about detected patterns
func (e *ErrorAnalyzer) GetPatternStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_patterns":   len(e.patterns),
		"last_scan":        e.lastScan,
		"min_cluster_size": e.minClusterSize,
	}

	if len(e.patterns) > 0 {
		totalFreq := 0
		highConfidence := 0

		for _, p := range e.patterns {
			totalFreq += p.Frequency
			if p.Confidence >= 0.7 {
				highConfidence++
			}
		}

		stats["total_occurrences"] = totalFreq
		stats["high_confidence_patterns"] = highConfidence
		stats["avg_frequency"] = float64(totalFreq) / float64(len(e.patterns))
	}

	return stats
}
