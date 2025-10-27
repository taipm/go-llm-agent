package mongodb

import (
	"context"
	"testing"
)

// TestConnectTool verifies the connect tool can be created
func TestConnectTool(t *testing.T) {
	tool := NewConnectTool()
	if tool.Name() != "mongodb_connect" {
		t.Errorf("Expected name 'mongodb_connect', got '%s'", tool.Name())
	}
	if tool.Category() != "database" {
		t.Errorf("Expected category 'database', got '%s'", tool.Category())
	}
	if !tool.IsSafe() {
		t.Error("Connect should be safe")
	}
}

// TestFindTool verifies the find tool can be created
func TestFindTool(t *testing.T) {
	tool := NewFindTool()
	if tool.Name() != "mongodb_find" {
		t.Errorf("Expected name 'mongodb_find', got '%s'", tool.Name())
	}
	if !tool.IsSafe() {
		t.Error("Find should be safe (read-only)")
	}
}

// TestInsertTool verifies the insert tool can be created
func TestInsertTool(t *testing.T) {
	tool := NewInsertTool()
	if tool.Name() != "mongodb_insert" {
		t.Errorf("Expected name 'mongodb_insert', got '%s'", tool.Name())
	}
	if tool.IsSafe() {
		t.Error("Insert should NOT be safe (modifies data)")
	}
}

// TestUpdateTool verifies the update tool can be created
func TestUpdateTool(t *testing.T) {
	tool := NewUpdateTool()
	if tool.Name() != "mongodb_update" {
		t.Errorf("Expected name 'mongodb_update', got '%s'", tool.Name())
	}
	if tool.IsSafe() {
		t.Error("Update should NOT be safe (modifies data)")
	}
}

// TestDeleteTool verifies the delete tool can be created
func TestDeleteTool(t *testing.T) {
	tool := NewDeleteTool()
	if tool.Name() != "mongodb_delete" {
		t.Errorf("Expected name 'mongodb_delete', got '%s'", tool.Name())
	}
	if tool.IsSafe() {
		t.Error("Delete should NOT be safe (destructive)")
	}
}

// TestGetConnectionNotFound verifies error when connection doesn't exist
func TestGetConnectionNotFound(t *testing.T) {
	_, err := GetConnection("nonexistent_id")
	if err == nil {
		t.Error("Expected error for nonexistent connection")
	}
}

// TestDeleteEmptyFilter verifies safety check for empty filter
func TestDeleteEmptyFilter(t *testing.T) {
	tool := NewDeleteTool()
	ctx := context.Background()
	
	params := map[string]interface{}{
		"connection_id": "test_conn",
		"collection":    "test_collection",
		"filter":        map[string]interface{}{}, // Empty filter
	}
	
	_, err := tool.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for empty filter in delete")
	}
}
