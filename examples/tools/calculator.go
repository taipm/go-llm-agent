package tools

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// CalculatorTool implements basic math operations
type CalculatorTool struct{}

func (c *CalculatorTool) Name() string {
	return "calculator"
}

func (c *CalculatorTool) Description() string {
	return "Performs basic math operations: add, subtract, multiply, divide, power, sqrt"
}

func (c *CalculatorTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"operation": {
				Type:        "string",
				Description: "The math operation to perform",
				Enum:        []interface{}{"add", "subtract", "multiply", "divide", "power", "sqrt"},
			},
			"a": {
				Type:        "number",
				Description: "First number",
			},
			"b": {
				Type:        "number",
				Description: "Second number (not needed for sqrt)",
			},
		},
		Required: []string{"operation", "a"},
	}
}

func (c *CalculatorTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	a, err := toFloat64(params["a"])
	if err != nil {
		return nil, fmt.Errorf("invalid parameter 'a': %w", err)
	}

	var result float64

	switch operation {
	case "sqrt":
		if a < 0 {
			return nil, fmt.Errorf("cannot take square root of negative number")
		}
		result = math.Sqrt(a)

	case "add", "subtract", "multiply", "divide", "power":
		b, err := toFloat64(params["b"])
		if err != nil {
			return nil, fmt.Errorf("invalid parameter 'b': %w", err)
		}

		switch operation {
		case "add":
			result = a + b
		case "subtract":
			result = a - b
		case "multiply":
			result = a * b
		case "divide":
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result = a / b
		case "power":
			result = math.Pow(a, b)
		}

	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return map[string]interface{}{
		"operation": operation,
		"result":    result,
	}, nil
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
