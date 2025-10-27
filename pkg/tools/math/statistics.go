package math

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
	"gonum.org/v1/gonum/stat"
)

// StatsTool performs statistical calculations on datasets
type StatsTool struct {
	tools.BaseTool
}

// NewStatsTool creates a new statistics tool
func NewStatsTool() *StatsTool {
	return &StatsTool{
		BaseTool: tools.NewBaseTool(
			"math_stats",
			"Calculate statistical measures on numerical datasets. Supports mean, median, mode, standard deviation, variance, min, max, sum, count, and quartiles. Example: Calculate mean and median of [10, 20, 30, 40, 50].",
			tools.CategoryMath,
			false, // no auth required
			true,  // safe operation (read-only)
		),
	}
}

// Parameters implements Tool.Parameters
func (t *StatsTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"data": {
				Type:        "array",
				Items:       &types.JSONSchema{Type: "number"},
				Description: "Numerical dataset to analyze",
			},
			"operations": {
				Type:        "array",
				Items:       &types.JSONSchema{Type: "string"},
				Description: "Statistical operations to perform (mean, median, mode, stddev, variance, min, max, sum, count, quartiles, all). Default: all",
			},
			"precision": {
				Type:        "integer",
				Description: "Number of decimal places for results (optional, default: 6, max: 15)",
			},
		},
		Required: []string{"data"},
	}
}

// Execute implements Tool.Execute
func (t *StatsTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract and validate data
	data, err := extractDataArray(params)
	if err != nil {
		return nil, err
	}

	// Extract precision
	precision := extractPrecision(params)

	// Extract operations
	operations := extractOperations(params)

	// Compute statistics
	result := computeStatistics(data, operations, precision)

	return result, nil
}

// extractDataArray extracts and validates the data array
func extractDataArray(params map[string]interface{}) ([]float64, error) {
	dataInterface, ok := params["data"]
	if !ok {
		return nil, fmt.Errorf("data is required")
	}

	dataArray, ok := dataInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array")
	}

	if len(dataArray) == 0 {
		return nil, fmt.Errorf("data array cannot be empty")
	}

	if len(dataArray) > 10000 {
		return nil, fmt.Errorf("data array too large (max 10,000 elements)")
	}

	// Convert to float64 slice
	data := make([]float64, len(dataArray))
	for i, v := range dataArray {
		var val float64
		switch num := v.(type) {
		case float64:
			val = num
		case int:
			val = float64(num)
		case json.Number:
			if f, err := num.Float64(); err == nil {
				val = f
			} else {
				return nil, fmt.Errorf("invalid number at index %d", i)
			}
		default:
			return nil, fmt.Errorf("invalid data type at index %d: must be number", i)
		}
		data[i] = val
	}

	return data, nil
}

// extractOperations extracts the operations to perform
func extractOperations(params map[string]interface{}) []string {
	operations := []string{"all"}
	if ops, ok := params["operations"]; ok {
		if opsArray, ok := ops.([]interface{}); ok {
			operations = make([]string, len(opsArray))
			for i, op := range opsArray {
				if opStr, ok := op.(string); ok {
					operations[i] = opStr
				}
			}
		}
	}
	return operations
}

// computeStatistics computes requested statistical measures
func computeStatistics(data []float64, operations []string, precision int) map[string]interface{} {
	result := map[string]interface{}{
		"type":  "statistics_result",
		"count": len(data),
	}

	// Check if we should compute all operations
	computeAll := containsOperation(operations, "all")

	// Mean
	if computeAll || containsOperation(operations, "mean") {
		mean := stat.Mean(data, nil)
		result["mean"] = roundToPrecision(mean, precision)
	}

	// Median
	if computeAll || containsOperation(operations, "median") {
		result["median"] = roundToPrecision(calculateMedian(data), precision)
	}

	// Mode
	if computeAll || containsOperation(operations, "mode") {
		result["mode"] = calculateMode(data, precision)
	}

	// Standard Deviation
	if computeAll || containsOperation(operations, "stddev") {
		stddev := stat.StdDev(data, nil)
		result["stddev"] = roundToPrecision(stddev, precision)
	}

	// Variance
	if computeAll || containsOperation(operations, "variance") {
		variance := stat.Variance(data, nil)
		result["variance"] = roundToPrecision(variance, precision)
	}

	// Min
	if computeAll || containsOperation(operations, "min") {
		result["min"] = roundToPrecision(calculateMin(data), precision)
	}

	// Max
	if computeAll || containsOperation(operations, "max") {
		result["max"] = roundToPrecision(calculateMax(data), precision)
	}

	// Sum
	if computeAll || containsOperation(operations, "sum") {
		result["sum"] = roundToPrecision(calculateSum(data), precision)
	}

	// Quartiles
	if computeAll || containsOperation(operations, "quartiles") {
		result["quartiles"] = calculateQuartiles(data, precision)
	}

	return result
}

// containsOperation checks if an operation is in the operations list
func containsOperation(operations []string, target string) bool {
	for _, op := range operations {
		if op == target {
			return true
		}
	}
	return false
}

// calculateMedian calculates the median of a dataset
func calculateMedian(data []float64) float64 {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// calculateMode calculates the mode(s) of a dataset
func calculateMode(data []float64, precision int) []float64 {
	freqMap := make(map[float64]int)
	for _, val := range data {
		freqMap[val]++
	}

	maxFreq := 0
	modes := []float64{}
	for val, freq := range freqMap {
		if freq > maxFreq {
			maxFreq = freq
			modes = []float64{val}
		} else if freq == maxFreq {
			modes = append(modes, val)
		}
	}

	// Round modes
	roundedModes := make([]float64, len(modes))
	for i, mode := range modes {
		roundedModes[i] = roundToPrecision(mode, precision)
	}
	return roundedModes
}

// calculateMin finds the minimum value in a dataset
func calculateMin(data []float64) float64 {
	min := data[0]
	for _, val := range data {
		if val < min {
			min = val
		}
	}
	return min
}

// calculateMax finds the maximum value in a dataset
func calculateMax(data []float64) float64 {
	max := data[0]
	for _, val := range data {
		if val > max {
			max = val
		}
	}
	return max
}

// calculateSum calculates the sum of a dataset
func calculateSum(data []float64) float64 {
	sum := 0.0
	for _, val := range data {
		sum += val
	}
	return sum
}

// calculateQuartiles calculates Q1, Q2, Q3
func calculateQuartiles(data []float64, precision int) map[string]interface{} {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	return map[string]interface{}{
		"q1": roundToPrecision(stat.Quantile(0.25, stat.Empirical, sorted, nil), precision),
		"q2": roundToPrecision(stat.Quantile(0.50, stat.Empirical, sorted, nil), precision),
		"q3": roundToPrecision(stat.Quantile(0.75, stat.Empirical, sorted, nil), precision),
	}
}
