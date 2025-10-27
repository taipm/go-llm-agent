package datetime

import (
	"context"
	"fmt"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// NowTool returns the current date and time
type NowTool struct {
	tools.BaseTool
}

// NewNowTool creates a new current time tool
func NewNowTool() *NowTool {
	return &NowTool{
		BaseTool: tools.NewBaseTool(
			"datetime_now",
			"Get the current date and time. Returns current datetime in RFC3339 format with timezone by default (e.g., '2025-10-27T15:30:00Z' or '2025-10-27T22:30:00+07:00'). Use timezone parameter to get time in specific timezone. Use this tool to get current time before calculating time differences or age.",
			tools.CategoryDateTime,
			false, // no auth required
			true,  // safe operation
		),
	}
}

// Parameters implements Tool.Parameters
func (t *NowTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"format": {
				Type:        "string",
				Description: "Output format. Default: RFC3339 (recommended, includes timezone). Use 'RFC3339' for datetime calculations. Examples: RFC3339='2025-10-27T15:30:00Z', RFC1123='Mon, 27 Oct 2025 15:30:00 UTC'",
				Enum: []interface{}{
					"RFC3339",     // 2006-01-02T15:04:05Z07:00
					"RFC3339Nano", // 2006-01-02T15:04:05.999999999Z07:00
					"RFC822",      // 02 Jan 06 15:04 MST
					"RFC822Z",     // 02 Jan 06 15:04 -0700
					"RFC1123",     // Mon, 02 Jan 2006 15:04:05 MST
					"RFC1123Z",    // Mon, 02 Jan 2006 15:04:05 -0700
					"Kitchen",     // 3:04PM
					"ANSIC",       // Mon Jan _2 15:04:05 2006
					"UnixDate",    // Mon Jan _2 15:04:05 MST 2006
					"Stamp",       // Jan _2 15:04:05
					"custom",      // Use custom_format field
				},
			},
			"custom_format": {
				Type:        "string",
				Description: "Custom Go time format (e.g., '2006-01-02 15:04:05'). Only used when format='custom'",
			},
			"timezone": {
				Type:        "string",
				Description: "IANA timezone name. Examples: 'UTC' (default), 'Asia/Ho_Chi_Minh' (Vietnam GMT+7), 'America/New_York' (US EST/EDT), 'Europe/London'. Use this to get current time in user's local timezone.",
			},
		},
	}
}

// Execute implements Tool.Execute
func (t *NowTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Get current time
	now := time.Now()

	// Handle timezone
	timezone := "UTC"
	if tz, ok := params["timezone"].(string); ok && tz != "" {
		timezone = tz
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}
	now = now.In(loc)

	// Handle format
	formatName := "RFC3339"
	if f, ok := params["format"].(string); ok && f != "" {
		formatName = f
	}

	var formatted string
	switch formatName {
	case "RFC3339":
		formatted = now.Format(time.RFC3339)
	case "RFC3339Nano":
		formatted = now.Format(time.RFC3339Nano)
	case "RFC822":
		formatted = now.Format(time.RFC822)
	case "RFC822Z":
		formatted = now.Format(time.RFC822Z)
	case "RFC1123":
		formatted = now.Format(time.RFC1123)
	case "RFC1123Z":
		formatted = now.Format(time.RFC1123Z)
	case "Kitchen":
		formatted = now.Format(time.Kitchen)
	case "ANSIC":
		formatted = now.Format(time.ANSIC)
	case "UnixDate":
		formatted = now.Format(time.UnixDate)
	case "Stamp":
		formatted = now.Format(time.Stamp)
	case "custom":
		customFormat, ok := params["custom_format"].(string)
		if !ok || customFormat == "" {
			return nil, fmt.Errorf("custom_format required when format='custom'")
		}
		formatted = now.Format(customFormat)
	default:
		return nil, fmt.Errorf("unsupported format: %s", formatName)
	}

	return map[string]interface{}{
		"datetime":  formatted,
		"unix":      now.Unix(),
		"unix_nano": now.UnixNano(),
		"timezone":  timezone,
		"format":    formatName,
	}, nil
}
