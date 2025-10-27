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
			"Get the current date and time in a specified format and timezone",
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
				Description: "Time format string (RFC3339, RFC822, Kitchen, or custom Go format). Default: RFC3339",
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
				Description: "IANA timezone (e.g., 'America/New_York', 'Asia/Tokyo'). Default: UTC",
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
