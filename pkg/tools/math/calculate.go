package math

import (
	"context"
	"encoding/json"
	"fmt"
	stdmath "math"

	"github.com/Knetic/govaluate"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// CalculateTool performs safe mathematical calculations and expression evaluation
type CalculateTool struct {
	tools.BaseTool
}

// NewCalculateTool creates a new math calculation tool
func NewCalculateTool() *CalculateTool {
	return &CalculateTool{
		BaseTool: tools.NewBaseTool(
			"math_calculate",
			"Safely evaluate mathematical expressions and perform calculations. Supports basic arithmetic (+, -, *, /, ^, %), trigonometry (sin, cos, tan), logarithms (log, ln), and common functions (sqrt, abs, ceil, floor). Example: '2 + 2 * 3' returns 8, 'sin(PI/2)' returns 1.",
			tools.CategoryMath,
			false, // no auth required
			true,  // safe operation (read-only)
		),
	}
}

// Parameters implements Tool.Parameters
func (t *CalculateTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"expression": {
				Type:        "string",
				Description: "Mathematical expression to evaluate (e.g., '2 + 2', 'sin(PI/2)', 'sqrt(16)', 'log(100)')",
			},
			"precision": {
				Type:        "integer",
				Description: "Number of decimal places for result (optional, default: 6, max: 15)",
			},
			"variables": {
				Type:        "object",
				Description: "Variable definitions for the expression (optional, e.g., {\"x\": 5, \"y\": 3})",
			},
		},
		Required: []string{"expression"},
	}
}

// Execute implements Tool.Execute
func (t *CalculateTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract and validate expression
	expression, err := extractExpression(params)
	if err != nil {
		return nil, err
	}

	// Extract precision
	precision := extractPrecision(params)

	// Extract variables and add constants
	variables := extractVariables(params)

	// Create evaluator with math functions
	expr, err := createEvaluator(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid expression: %w", err)
	}

	// Evaluate expression
	result, err := expr.Evaluate(variables)
	if err != nil {
		return nil, fmt.Errorf("evaluation error: %w", err)
	}

	// Convert and validate result
	resultFloat, err := convertToFloat64(result)
	if err != nil {
		return nil, err
	}

	// Round to specified precision
	roundedResult := roundToPrecision(resultFloat, precision)

	// Build result
	variablesUsed := getVariablesUsed(variables)

	return map[string]interface{}{
		"type":           "calculation_result",
		"expression":     expression,
		"result":         roundedResult,
		"precision":      precision,
		"variables_used": variablesUsed,
	}, nil
}

// extractExpression extracts and validates the expression parameter
func extractExpression(params map[string]interface{}) (string, error) {
	expression, ok := params["expression"].(string)
	if !ok || expression == "" {
		return "", fmt.Errorf("expression is required and must be a string")
	}

	if len(expression) > 1000 {
		return "", fmt.Errorf("expression too long (max 1000 characters)")
	}

	return expression, nil
}

// extractPrecision extracts precision parameter with default value
func extractPrecision(params map[string]interface{}) int {
	precision := 6 // default
	if p, ok := params["precision"]; ok {
		switch v := p.(type) {
		case float64:
			precision = int(v)
		case int:
			precision = v
		case json.Number:
			if pInt, err := v.Int64(); err == nil {
				precision = int(pInt)
			}
		}
		if precision < 0 {
			precision = 0
		} else if precision > 15 {
			precision = 15
		}
	}
	return precision
}

// extractVariables extracts variables and adds mathematical constants
func extractVariables(params map[string]interface{}) map[string]interface{} {
	variables := make(map[string]interface{})
	if v, ok := params["variables"]; ok {
		if varsMap, ok := v.(map[string]interface{}); ok {
			variables = varsMap
		}
	}

	// Add mathematical constants
	variables["PI"] = stdmath.Pi
	variables["E"] = stdmath.E
	variables["pi"] = stdmath.Pi
	variables["e"] = stdmath.E

	return variables
}

// createEvaluator creates an expression evaluator with math functions
func createEvaluator(expression string) (*govaluate.EvaluableExpression, error) {
	functions := map[string]govaluate.ExpressionFunction{
		"sin":   func(args ...interface{}) (interface{}, error) { return stdmath.Sin(args[0].(float64)), nil },
		"cos":   func(args ...interface{}) (interface{}, error) { return stdmath.Cos(args[0].(float64)), nil },
		"tan":   func(args ...interface{}) (interface{}, error) { return stdmath.Tan(args[0].(float64)), nil },
		"sqrt":  func(args ...interface{}) (interface{}, error) { return stdmath.Sqrt(args[0].(float64)), nil },
		"abs":   func(args ...interface{}) (interface{}, error) { return stdmath.Abs(args[0].(float64)), nil },
		"log":   func(args ...interface{}) (interface{}, error) { return stdmath.Log10(args[0].(float64)), nil },
		"ln":    func(args ...interface{}) (interface{}, error) { return stdmath.Log(args[0].(float64)), nil },
		"exp":   func(args ...interface{}) (interface{}, error) { return stdmath.Exp(args[0].(float64)), nil },
		"ceil":  func(args ...interface{}) (interface{}, error) { return stdmath.Ceil(args[0].(float64)), nil },
		"floor": func(args ...interface{}) (interface{}, error) { return stdmath.Floor(args[0].(float64)), nil },
		"round": func(args ...interface{}) (interface{}, error) { return stdmath.Round(args[0].(float64)), nil },
		"pow": func(args ...interface{}) (interface{}, error) {
			return stdmath.Pow(args[0].(float64), args[1].(float64)), nil
		},
		"min": func(args ...interface{}) (interface{}, error) {
			return stdmath.Min(args[0].(float64), args[1].(float64)), nil
		},
		"max": func(args ...interface{}) (interface{}, error) {
			return stdmath.Max(args[0].(float64), args[1].(float64)), nil
		},
	}

	return govaluate.NewEvaluableExpressionWithFunctions(expression, functions)
}

// convertToFloat64 converts evaluation result to float64 and validates
func convertToFloat64(result interface{}) (float64, error) {
	var resultFloat float64
	switch v := result.(type) {
	case float64:
		resultFloat = v
	case int:
		resultFloat = float64(v)
	case int64:
		resultFloat = float64(v)
	default:
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	if stdmath.IsNaN(resultFloat) {
		return 0, fmt.Errorf("result is not a number (NaN)")
	}
	if stdmath.IsInf(resultFloat, 0) {
		return 0, fmt.Errorf("result is infinity")
	}

	return resultFloat, nil
}

// roundToPrecision rounds a float to the specified number of decimal places
func roundToPrecision(value float64, precision int) float64 {
	multiplier := stdmath.Pow(10, float64(precision))
	return stdmath.Round(value*multiplier) / multiplier
}

// getVariablesUsed returns list of user-provided variables (excluding constants)
func getVariablesUsed(variables map[string]interface{}) []string {
	used := []string{}
	for key := range variables {
		if key != "PI" && key != "E" && key != "pi" && key != "e" {
			used = append(used, key)
		}
	}
	return used
}
