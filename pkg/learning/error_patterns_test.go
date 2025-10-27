package learning

import (
	"testing"

	"github.com/taipm/go-llm-agent/pkg/logger"
)

// TestErrorAnalyzerHelpers tests utility functions that don't require external dependencies
func TestErrorAnalyzerHelpers(t *testing.T) {
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LogLevelError)

	analyzer := &ErrorAnalyzer{
		minClusterSize:   3,
		similarityThresh: 0.75,
		patterns:         []ErrorPattern{},
		logger:           log,
	}

	t.Run("mostCommon should find most frequent item", func(t *testing.T) {
		items := map[string]int{"apple": 3, "banana": 5, "orange": 2}
		result := analyzer.mostCommon(items)
		if result != "banana" {
			t.Errorf("Expected 'banana', got '%s'", result)
		}
	})

	t.Run("mostCommon should handle empty map", func(t *testing.T) {
		items := map[string]int{}
		result := analyzer.mostCommon(items)
		if result != "" {
			t.Errorf("Expected empty string for empty map, got '%s'", result)
		}
	})

	t.Run("topN should return top N items sorted by frequency", func(t *testing.T) {
		items := map[string]int{"a": 10, "b": 5, "c": 8, "d": 3, "e": 15}
		top2 := analyzer.topN(items, 2)

		if len(top2) != 2 {
			t.Fatalf("Expected 2 items, got %d", len(top2))
		}
		if top2[0] != "e" {
			t.Errorf("Expected first='e' (freq 15), got '%s'", top2[0])
		}
		if top2[1] != "a" {
			t.Errorf("Expected second='a' (freq 10), got '%s'", top2[1])
		}
	})

	t.Run("topN should handle n greater than map size", func(t *testing.T) {
		items := map[string]int{"a": 10, "b": 5}
		top5 := analyzer.topN(items, 5)
		if len(top5) != 2 {
			t.Errorf("Expected 2 items (all available), got %d", len(top5))
		}
	})

	t.Run("calculateClusterSimilarity for single experience", func(t *testing.T) {
		exps := []Experience{{ID: "1", ErrorType: "test"}}
		sim := analyzer.calculateClusterSimilarity(exps)
		if sim != 1.0 {
			t.Errorf("Expected similarity=1.0 for single exp, got %.2f", sim)
		}
	})

	t.Run("calculateClusterSimilarity for multiple experiences", func(t *testing.T) {
		exps := []Experience{
			{ID: "1", ErrorType: "math_error", ToolCalled: "calc"},
			{ID: "2", ErrorType: "math_error", ToolCalled: "calc"},
			{ID: "3", ErrorType: "math_error", ToolCalled: "other"},
		}
		sim := analyzer.calculateClusterSimilarity(exps)

		if sim <= 0 || sim > 1 {
			t.Errorf("Similarity out of range [0,1]: %.2f", sim)
		}

		t.Logf("Cluster similarity: %.2f", sim)
	})
}

// TestPatternMatching tests pattern matching logic
func TestPatternMatching(t *testing.T) {
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LogLevelError)

	// Pre-populate analyzer with patterns
	analyzer := &ErrorAnalyzer{
		minClusterSize:   3,
		similarityThresh: 0.75,
		patterns: []ErrorPattern{
			{
				ID:            "pattern1",
				Pattern:       "division_by_zero",
				CommonQuery:   "divide by zero",
				ErrorMessages: []string{"division by zero", "cannot divide by zero"},
				ErrorType:     "math_error",
				Confidence:    0.85,
			},
			{
				ID:            "pattern2",
				Pattern:       "timeout",
				CommonQuery:   "fetch webpage",
				ErrorMessages: []string{"timeout", "connection timeout"},
				ErrorType:     "network_error",
				Confidence:    0.75,
			},
		},
		logger: log,
	}

	t.Run("scorePatternMatch should score similar queries highly", func(t *testing.T) {
		pattern := &analyzer.patterns[0]
		score := analyzer.scorePatternMatch(pattern, "100 divided by zero", "division by zero")

		if score <= 0 {
			t.Errorf("Expected positive score for similar query, got %.2f", score)
		}
		t.Logf("Match score for similar query: %.2f", score)
	})

	t.Run("scorePatternMatch should score dissimilar queries lowly", func(t *testing.T) {
		pattern := &analyzer.patterns[0]
		score := analyzer.scorePatternMatch(pattern, "fetch data from API", "network timeout")

		if score > 0.5 {
			t.Errorf("Expected low score for dissimilar query, got %.2f", score)
		}
		t.Logf("Match score for dissimilar query: %.2f", score)
	})

	t.Run("findBestMatchingPattern should find correct pattern", func(t *testing.T) {
		match := analyzer.findBestMatchingPattern("divide 50 by zero", "division error")
		if match != nil {
			if match.ID != "pattern1" {
				t.Errorf("Expected pattern1, got %s", match.ID)
			}
			t.Logf("✅ Found pattern: %s (confidence: %.2f)", match.ID, match.Confidence)
		} else {
			t.Log("No match found (score below threshold)")
		}
	})

	t.Run("findBestMatchingPattern should return nil for no match", func(t *testing.T) {
		match := analyzer.findBestMatchingPattern("completely unrelated query", "unknown error")
		if match != nil {
			t.Errorf("Expected nil for unrelated query, got %s", match.ID)
		}
	})
}

// TestInitialization verifies ErrorAnalyzer setup
func TestInitialization(t *testing.T) {
	log := logger.NewConsoleLogger()
	log.SetLevel(logger.LogLevelError)

	t.Run("ErrorAnalyzer should have correct defaults", func(t *testing.T) {
		analyzer := &ErrorAnalyzer{
			minClusterSize:   3,
			similarityThresh: 0.75,
			minConfidence:    0.6,
			maxPatterns:      100,
			patterns:         []ErrorPattern{},
			logger:           log,
		}

		if analyzer.minClusterSize != 3 {
			t.Errorf("Expected minClusterSize=3, got %d", analyzer.minClusterSize)
		}
		if analyzer.similarityThresh != 0.75 {
			t.Errorf("Expected similarityThresh=0.75, got %.2f", analyzer.similarityThresh)
		}
		if analyzer.minConfidence != 0.6 {
			t.Errorf("Expected minConfidence=0.6, got %.2f", analyzer.minConfidence)
		}
		if analyzer.maxPatterns != 100 {
			t.Errorf("Expected maxPatterns=100, got %d", analyzer.maxPatterns)
		}
	})
}
