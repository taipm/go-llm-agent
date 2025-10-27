package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/pkg/builtin"
)

func main() {
	fmt.Println("=== Math Tools Demo ===\n")

	// Get registry with all builtin tools (including math)
	registry := builtin.GetRegistry()

	ctx := context.Background()

	// Demo 1: Basic arithmetic with math_calculate
	fmt.Println("1. Basic Arithmetic:")
	result, err := registry.Execute(ctx, "math_calculate", map[string]interface{}{
		"expression": "2 + 2 * 3",
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("2 + 2 * 3", result)

	// Demo 2: Trigonometry
	fmt.Println("\n2. Trigonometry:")
	result, err = registry.Execute(ctx, "math_calculate", map[string]interface{}{
		"expression": "sin(PI/2)",
		"precision":  4,
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("sin(PI/2)", result)

	// Demo 3: Complex expression with variables
	fmt.Println("\n3. Expression with Variables:")
	result, err = registry.Execute(ctx, "math_calculate", map[string]interface{}{
		"expression": "sqrt(x^2 + y^2)",
		"variables": map[string]interface{}{
			"x": 3.0,
			"y": 4.0,
		},
		"precision": 2,
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("sqrt(x^2 + y^2) where x=3, y=4", result)

	// Demo 4: Logarithms and exponentials
	fmt.Println("\n4. Logarithms:")
	result, err = registry.Execute(ctx, "math_calculate", map[string]interface{}{
		"expression": "log(100) + ln(E)",
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("log(100) + ln(E)", result)

	// Demo 5: Statistics - Basic measures
	fmt.Println("\n5. Statistics - Basic Measures:")
	data := []interface{}{10.0, 20.0, 30.0, 40.0, 50.0}
	result, err = registry.Execute(ctx, "math_stats", map[string]interface{}{
		"data":       data,
		"operations": []interface{}{"mean", "median", "stddev"},
		"precision":  2,
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("Dataset: [10, 20, 30, 40, 50]", result)

	// Demo 6: Statistics - All operations
	fmt.Println("\n6. Statistics - All Operations:")
	salesData := []interface{}{100.5, 200.3, 150.0, 175.8, 225.4, 180.2, 195.6}
	result, err = registry.Execute(ctx, "math_stats", map[string]interface{}{
		"data":      salesData,
		"precision": 2,
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("Sales Data (7 days)", result)

	// Demo 7: Mode calculation
	fmt.Println("\n7. Finding Mode:")
	scoreData := []interface{}{85.0, 90.0, 85.0, 92.0, 88.0, 85.0, 90.0}
	result, err = registry.Execute(ctx, "math_stats", map[string]interface{}{
		"data":       scoreData,
		"operations": []interface{}{"mode"},
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("Test Scores", result)

	// Demo 8: Quartiles for distribution analysis
	fmt.Println("\n8. Quartiles Analysis:")
	dataset := []interface{}{10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0}
	result, err = registry.Execute(ctx, "math_stats", map[string]interface{}{
		"data":       dataset,
		"operations": []interface{}{"quartiles", "mean"},
		"precision":  1,
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("Dataset Distribution", result)

	// Demo 9: Financial calculation
	fmt.Println("\n9. Financial Calculation (Compound Interest):")
	// Formula: A = P * (1 + r)^t
	result, err = registry.Execute(ctx, "math_calculate", map[string]interface{}{
		"expression": "P * pow(1 + r, t)",
		"variables": map[string]interface{}{
			"P": 10000.0, // Principal
			"r": 0.05,    // 5% annual rate
			"t": 10.0,    // 10 years
		},
		"precision": 2,
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("$10,000 at 5% for 10 years", result)

	// Demo 10: Scientific calculation
	fmt.Println("\n10. Scientific Formula (Distance):")
	// Distance traveled: d = v*t + 0.5*a*t^2
	result, err = registry.Execute(ctx, "math_calculate", map[string]interface{}{
		"expression": "v * t + 0.5 * a * pow(t, 2)",
		"variables": map[string]interface{}{
			"v": 20.0, // initial velocity (m/s)
			"a": 2.0,  // acceleration (m/s^2)
			"t": 5.0,  // time (s)
		},
		"precision": 2,
	})
	if err != nil {
		log.Fatal(err)
	}
	printResult("Distance with v=20m/s, a=2m/s², t=5s", result)

	fmt.Println("\n=== Math Tools Available ===")
	mathTools := builtin.GetMathTools()
	for _, tool := range mathTools {
		fmt.Printf("- %s: %s\n", tool.Name(), tool.Description())
	}
}

func printResult(description string, result interface{}) {
	fmt.Printf("   %s\n", description)
	if resultMap, ok := result.(map[string]interface{}); ok {
		for key, value := range resultMap {
			if key != "type" {
				fmt.Printf("   - %s: %v\n", key, value)
			}
		}
	} else {
		fmt.Printf("   Result: %v\n", result)
	}
}
