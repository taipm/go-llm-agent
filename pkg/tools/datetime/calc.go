package datetime

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// CalcTool performs datetime arithmetic operations
type CalcTool struct {
	tools.BaseTool
}

// NewCalcTool creates a new datetime calculation tool
func NewCalcTool() *CalcTool {
	return &CalcTool{
		BaseTool: tools.NewBaseTool(
			"datetime_calc",
			"Perform date and time calculations: add/subtract durations, find differences between dates.",
			tools.CategoryDateTime,
			false, // no auth required
			true,  // safe operation
		),
	}
}

// Parameters implements Tool.Parameters
func (t *CalcTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"datetime": {
				Type:        "string",
				Description: "The base datetime string (RFC3339, Unix timestamp, or custom format)",
			},
			"operation": {
				Type:        "string",
				Description: "Operation to perform: 'add' or 'subtract' duration, or 'diff' to find difference",
				Enum:        []interface{}{"add", "subtract", "diff"},
			},
			"duration": {
				Type:        "string",
				Description: "Duration to add/subtract (e.g., '2h30m', '24h', '7d'). Supports: ns, us/µs, ms, s, m, h, d (days)",
			},
			"target_datetime": {
				Type:        "string",
				Description: "Target datetime for 'diff' operation (calculates datetime - target)",
			},
			"format": {
				Type:        "string",
				Description: "Input datetime format. Default: RFC3339",
				Enum: []interface{}{
					"RFC3339",
					"RFC3339Nano",
					"RFC822",
					"RFC822Z",
					"RFC1123",
					"RFC1123Z",
					"Unix",
					"UnixMilli",
					"UnixNano",
					"custom",
				},
			},
			"custom_format": {
				Type:        "string",
				Description: "Custom format string (Go time format). Only used when format='custom'",
			},
			"timezone": {
				Type:        "string",
				Description: "Timezone for datetime (IANA format). Default: UTC",
			},
			"output_format": {
				Type:        "string",
				Description: "Output format for result. Default: RFC3339",
				Enum: []interface{}{
					"RFC3339",
					"RFC3339Nano",
					"RFC1123",
					"Unix",
					"UnixMilli",
					"UnixNano",
					"custom",
				},
			},
			"output_custom_format": {
				Type:        "string",
				Description: "Custom output format. Only used when output_format='custom'",
			},
		},
		Required: []string{"datetime", "operation"},
	}
}

// Execute implements Tool.Execute
func (t *CalcTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	datetimeStr, operation, durationStr, targetStr, format, customFormat, timezone, outputFormat, outputCustom, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Parse base datetime
	formatTool := NewFormatTool()
	baseTime, err := formatTool.parseDateTime(datetimeStr, format, customFormat, timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime: %w", err)
	}

	var resultTime time.Time
	var diffDuration time.Duration

	switch operation {
	case "add":
		if durationStr == "" {
			return nil, fmt.Errorf("duration parameter required for 'add' operation")
		}
		duration, err := t.parseDuration(durationStr)
		if err != nil {
			return nil, err
		}
		resultTime = baseTime.Add(duration)

	case "subtract":
		if durationStr == "" {
			return nil, fmt.Errorf("duration parameter required for 'subtract' operation")
		}
		duration, err := t.parseDuration(durationStr)
		if err != nil {
			return nil, err
		}
		resultTime = baseTime.Add(-duration)

	case "diff":
		if targetStr == "" {
			return nil, fmt.Errorf("target_datetime parameter required for 'diff' operation")
		}
		targetTime, err := formatTool.parseDateTime(targetStr, format, customFormat, timezone)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target_datetime: %w", err)
		}
		diffDuration = baseTime.Sub(targetTime)
		resultTime = baseTime // For diff, result is the base time

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	// Format output
	formatted, err := formatTool.formatDateTime(resultTime, outputFormat, outputCustom)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"operation": operation,
		"result":    formatted,
		"unix":      resultTime.Unix(),
		"unix_nano": resultTime.UnixNano(),
		"timezone":  timezone,
	}

	// Add duration info for diff operation
	if operation == "diff" {
		result["difference"] = map[string]interface{}{
			"duration":     diffDuration.String(),
			"nanoseconds":  diffDuration.Nanoseconds(),
			"microseconds": diffDuration.Microseconds(),
			"milliseconds": diffDuration.Milliseconds(),
			"seconds":      diffDuration.Seconds(),
			"minutes":      diffDuration.Minutes(),
			"hours":        diffDuration.Hours(),
			"days":         diffDuration.Hours() / 24,
		}
	}

	// Add input duration info for add/subtract
	if operation == "add" || operation == "subtract" {
		result["duration"] = durationStr
	}

	return result, nil
}

// extractParams extracts and validates parameters
func (t *CalcTool) extractParams(params map[string]interface{}) (string, string, string, string, string, string, string, string, string, error) {
	// datetime (required)
	datetimeVal, ok := params["datetime"]
	if !ok {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("datetime parameter is required")
	}
	datetime, ok := datetimeVal.(string)
	if !ok {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("datetime must be a string")
	}
	if datetime == "" {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("datetime cannot be empty")
	}

	// operation (required)
	operationVal, ok := params["operation"]
	if !ok {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("operation parameter is required")
	}
	operation, ok := operationVal.(string)
	if !ok {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("operation must be a string")
	}

	// duration (optional)
	duration := ""
	if d, ok := params["duration"].(string); ok {
		duration = d
	}

	// target_datetime (optional)
	target := ""
	if t, ok := params["target_datetime"].(string); ok {
		target = t
	}

	// format (optional, default: RFC3339)
	format := "RFC3339"
	if f, ok := params["format"].(string); ok && f != "" {
		format = f
	}

	// custom_format (optional)
	customFormat := ""
	if f, ok := params["custom_format"].(string); ok {
		customFormat = f
	}

	// timezone (optional, default: UTC)
	timezone := "UTC"
	if tz, ok := params["timezone"].(string); ok && tz != "" {
		timezone = tz
	}

	// output_format (optional, default: RFC3339)
	outputFormat := "RFC3339"
	if f, ok := params["output_format"].(string); ok && f != "" {
		outputFormat = f
	}

	// output_custom_format (optional)
	outputCustom := ""
	if f, ok := params["output_custom_format"].(string); ok {
		outputCustom = f
	}

	return datetime, operation, duration, target, format, customFormat, timezone, outputFormat, outputCustom, nil
}

// parseDuration parses a duration string with support for days
func (t *CalcTool) parseDuration(durationStr string) (time.Duration, error) {
	// Check if duration includes days (not supported by time.ParseDuration)
	if len(durationStr) > 0 {
		lastChar := durationStr[len(durationStr)-1]
		if lastChar == 'd' || lastChar == 'D' {
			// Extract number of days
			daysStr := durationStr[:len(durationStr)-1]
			days, err := strconv.ParseFloat(daysStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid duration: %w", err)
			}
			return time.Duration(days * 24 * float64(time.Hour)), nil
		}
	}

	// Use standard time.ParseDuration for other units
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %s: %w (supported units: ns, us/µs, ms, s, m, h, d)", durationStr, err)
	}

	return duration, nil
}
