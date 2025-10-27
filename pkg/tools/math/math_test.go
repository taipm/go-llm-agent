package math

import (
	"context"
	"testing"
)

func TestCalculateBasic(t *testing.T) {
	tool := NewCalculateTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"expression": "2 + 2",
	}
	
	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	
	resultMap := result.(map[string]interface{})
	if resultMap["result"] != 4.0 {
		t.Errorf("Expected 4.0, got %v", resultMap["result"])
	}
}

func TestStatsBasic(t *testing.T) {
	tool := NewStatsTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"data": []interface{}{10.0, 20.0, 30.0, 40.0, 50.0},
	}
	
	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	
	resultMap := result.(map[string]interface{})
	if resultMap["mean"] != 30.0 {
		t.Errorf("Expected mean 30.0, got %v", resultMap["mean"])
	}
}
