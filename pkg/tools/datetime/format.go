package datetime

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// FormatTool formats a datetime string from one format to another
type FormatTool struct {
	tools.BaseTool
}

// NewFormatTool creates a new datetime format conversion tool
func NewFormatTool() *FormatTool {
	return &FormatTool{
		BaseTool: tools.NewBaseTool(
			"datetime_format",
			"Convert datetime between different formats and timezones. Supports Unix timestamps, RFC formats, and custom formats.",
			tools.CategoryDateTime,
			false, // no auth required
			true,  // safe operation
		),
	}
}

// Parameters implements Tool.Parameters
func (t *FormatTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"datetime": {
				Type:        "string",
				Description: "The datetime string to format, or Unix timestamp (number as string)",
			},
			"from_format": {
				Type:        "string",
				Description: "Source format (RFC3339, RFC1123, Unix, UnixNano, or custom). Default: RFC3339",
				Enum: []interface{}{
					"RFC3339",
					"RFC3339Nano",
					"RFC822",
					"RFC822Z",
					"RFC1123",
					"RFC1123Z",
					"Kitchen",
					"ANSIC",
					"UnixDate",
					"Stamp",
					"Unix",
					"UnixMilli",
					"UnixNano",
					"custom",
				},
			},
			"from_custom_format": {
				Type:        "string",
				Description: "Custom source format (Go time format string). Only used when from_format='custom'",
			},
			"to_format": {
				Type:        "string",
				Description: "Target format. Default: RFC3339",
				Enum: []interface{}{
					"RFC3339",
					"RFC3339Nano",
					"RFC822",
					"RFC822Z",
					"RFC1123",
					"RFC1123Z",
					"Kitchen",
					"ANSIC",
					"UnixDate",
					"Stamp",
					"Unix",
					"UnixMilli",
					"UnixNano",
					"custom",
				},
			},
			"to_custom_format": {
				Type:        "string",
				Description: "Custom target format (Go time format string). Only used when to_format='custom'",
			},
			"from_timezone": {
				Type:        "string",
				Description: "Source timezone (IANA format, e.g., 'America/New_York'). Default: UTC",
			},
			"to_timezone": {
				Type:        "string",
				Description: "Target timezone (IANA format). Default: UTC",
			},
		},
		Required: []string{"datetime"},
	}
}

// Execute implements Tool.Execute
func (t *FormatTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	datetimeStr, fromFormat, fromCustom, toFormat, toCustom, fromTZ, toTZ, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Parse the input datetime
	parsedTime, err := t.parseDateTime(datetimeStr, fromFormat, fromCustom, fromTZ)
	if err != nil {
		return nil, err
	}

	// Convert timezone if needed
	if toTZ != fromTZ {
		loc, err := time.LoadLocation(toTZ)
		if err != nil {
			return nil, fmt.Errorf("invalid target timezone %s: %w", toTZ, err)
		}
		parsedTime = parsedTime.In(loc)
	}

	// Format the output
	formatted, err := t.formatDateTime(parsedTime, toFormat, toCustom)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"formatted":     formatted,
		"unix":          parsedTime.Unix(),
		"unix_milli":    parsedTime.UnixMilli(),
		"unix_nano":     parsedTime.UnixNano(),
		"from_format":   fromFormat,
		"to_format":     toFormat,
		"from_timezone": fromTZ,
		"to_timezone":   toTZ,
	}, nil
}

// extractParams extracts and validates parameters
func (t *FormatTool) extractParams(params map[string]interface{}) (string, string, string, string, string, string, string, error) {
	// datetime (required)
	datetimeVal, ok := params["datetime"]
	if !ok {
		return "", "", "", "", "", "", "", fmt.Errorf("datetime parameter is required")
	}
	datetime, ok := datetimeVal.(string)
	if !ok {
		return "", "", "", "", "", "", "", fmt.Errorf("datetime must be a string")
	}
	if datetime == "" {
		return "", "", "", "", "", "", "", fmt.Errorf("datetime cannot be empty")
	}

	// from_format (optional, default: RFC3339)
	fromFormat := "RFC3339"
	if f, ok := params["from_format"].(string); ok && f != "" {
		fromFormat = f
	}

	// from_custom_format (optional)
	fromCustom := ""
	if f, ok := params["from_custom_format"].(string); ok {
		fromCustom = f
	}

	// to_format (optional, default: RFC3339)
	toFormat := "RFC3339"
	if f, ok := params["to_format"].(string); ok && f != "" {
		toFormat = f
	}

	// to_custom_format (optional)
	toCustom := ""
	if f, ok := params["to_custom_format"].(string); ok {
		toCustom = f
	}

	// from_timezone (optional, default: UTC)
	fromTZ := "UTC"
	if tz, ok := params["from_timezone"].(string); ok && tz != "" {
		fromTZ = tz
	}

	// to_timezone (optional, default: UTC)
	toTZ := "UTC"
	if tz, ok := params["to_timezone"].(string); ok && tz != "" {
		toTZ = tz
	}

	return datetime, fromFormat, fromCustom, toFormat, toCustom, fromTZ, toTZ, nil
}

// parseDateTime parses a datetime string according to the specified format
func (t *FormatTool) parseDateTime(datetimeStr, format, customFormat, timezone string) (time.Time, error) {
	var parsedTime time.Time
	var err error

	// Handle Unix timestamps
	switch format {
	case "Unix":
		timestamp, err := strconv.ParseInt(datetimeStr, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid Unix timestamp: %w", err)
		}
		parsedTime = time.Unix(timestamp, 0)

	case "UnixMilli":
		timestamp, err := strconv.ParseInt(datetimeStr, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid Unix millisecond timestamp: %w", err)
		}
		parsedTime = time.UnixMilli(timestamp)

	case "UnixNano":
		timestamp, err := strconv.ParseInt(datetimeStr, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid Unix nanosecond timestamp: %w", err)
		}
		parsedTime = time.Unix(0, timestamp)

	case "custom":
		if customFormat == "" {
			return time.Time{}, fmt.Errorf("from_custom_format required when from_format='custom'")
		}
		parsedTime, err = time.Parse(customFormat, datetimeStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse datetime with custom format: %w", err)
		}

	default:
		// Use standard Go time format constants
		goFormat, err := t.getGoFormat(format)
		if err != nil {
			return time.Time{}, err
		}
		parsedTime, err = time.Parse(goFormat, datetimeStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse datetime: %w", err)
		}
	}

	// Apply timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// If parsed time doesn't have timezone info, assume it's in the specified timezone
	if parsedTime.Location().String() == "UTC" && timezone != "UTC" {
		parsedTime = time.Date(
			parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
			parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(),
			parsedTime.Nanosecond(), loc,
		)
	} else {
		parsedTime = parsedTime.In(loc)
	}

	return parsedTime, nil
}

// formatDateTime formats a time.Time according to the specified format
func (t *FormatTool) formatDateTime(dt time.Time, format, customFormat string) (string, error) {
	switch format {
	case "Unix":
		return strconv.FormatInt(dt.Unix(), 10), nil
	case "UnixMilli":
		return strconv.FormatInt(dt.UnixMilli(), 10), nil
	case "UnixNano":
		return strconv.FormatInt(dt.UnixNano(), 10), nil
	case "custom":
		if customFormat == "" {
			return "", fmt.Errorf("to_custom_format required when to_format='custom'")
		}
		return dt.Format(customFormat), nil
	default:
		goFormat, err := t.getGoFormat(format)
		if err != nil {
			return "", err
		}
		return dt.Format(goFormat), nil
	}
}

// getGoFormat returns the Go time format constant for a format name
func (t *FormatTool) getGoFormat(format string) (string, error) {
	switch format {
	case "RFC3339":
		return time.RFC3339, nil
	case "RFC3339Nano":
		return time.RFC3339Nano, nil
	case "RFC822":
		return time.RFC822, nil
	case "RFC822Z":
		return time.RFC822Z, nil
	case "RFC1123":
		return time.RFC1123, nil
	case "RFC1123Z":
		return time.RFC1123Z, nil
	case "Kitchen":
		return time.Kitchen, nil
	case "ANSIC":
		return time.ANSIC, nil
	case "UnixDate":
		return time.UnixDate, nil
	case "Stamp":
		return time.Stamp, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
