package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// WeatherTool provides weather information (mock data for demo)
type WeatherTool struct{}

func (w *WeatherTool) Name() string {
	return "get_weather"
}

func (w *WeatherTool) Description() string {
	return "Get current weather information for a specific location"
}

func (w *WeatherTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"location": {
				Type:        "string",
				Description: "City name or location (e.g., 'Tokyo', 'New York', 'London')",
			},
			"unit": {
				Type:        "string",
				Description: "Temperature unit: 'celsius' or 'fahrenheit'",
				Enum:        []interface{}{"celsius", "fahrenheit"},
			},
		},
		Required: []string{"location"},
	}
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	location, ok := params["location"].(string)
	if !ok || location == "" {
		return nil, fmt.Errorf("location must be a non-empty string")
	}

	unit := "celsius"
	if u, ok := params["unit"].(string); ok {
		unit = u
	}

	// Mock weather data based on location
	weather := getMockWeather(location, unit)

	return weather, nil
}

// getMockWeather returns mock weather data
func getMockWeather(location, unit string) map[string]interface{} {
	// Simple mock based on location hash
	temps := map[string]int{
		"Tokyo":     22,
		"New York":  18,
		"London":    15,
		"Paris":     17,
		"Sydney":    25,
		"Singapore": 30,
		"Moscow":    5,
		"Dubai":     35,
	}

	conditions := []string{"Sunny", "Cloudy", "Partly Cloudy", "Rainy", "Clear"}

	// Default temperature
	temp := 20
	if t, exists := temps[location]; exists {
		temp = t
	}

	// Convert if fahrenheit
	tempStr := fmt.Sprintf("%d°C", temp)
	if unit == "fahrenheit" {
		fahrenheit := (temp * 9 / 5) + 32
		tempStr = fmt.Sprintf("%d°F", fahrenheit)
	}

	// Pick condition based on location length (simple mock)
	condition := conditions[len(location)%len(conditions)]

	return map[string]interface{}{
		"location":    location,
		"temperature": tempStr,
		"condition":   condition,
		"humidity":    fmt.Sprintf("%d%%", 50+(len(location)*3)%40),
		"wind_speed":  fmt.Sprintf("%d km/h", 5+(len(location)*2)%20),
		"timestamp":   time.Now().Format(time.RFC3339),
		"forecast":    fmt.Sprintf("%s conditions expected to continue", condition),
	}
}
