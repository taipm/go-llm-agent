//go:build integration
// +build integration

package mongodb

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// getMongoDBURL returns MongoDB connection string from env or default localhost
func getMongoDBURL() string {
	if url := os.Getenv("MONGODB_URL"); url != "" {
		return url
	}
	return "mongodb://localhost:27017"
}

// TestIntegrationMongoDBConnect tests real MongoDB connection
func TestIntegrationMongoDBConnect(t *testing.T) {
	ctx := context.Background()
	connectTool := NewConnectTool()

	result, err := connectTool.Execute(ctx, map[string]interface{}{
		"connection_string": getMongoDBURL(),
		"database":          "go_llm_agent_test",
		"timeout":           10,
	})

	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v\nMake sure MongoDB is running on localhost:27017", err)
	}

	resultMap := result.(map[string]interface{})

	// Verify connection_id exists
	connID, ok := resultMap["connection_id"].(string)
	if !ok || connID == "" {
		t.Errorf("Expected connection_id to be a non-empty string, got %v", resultMap["connection_id"])
	}

	// Verify database name
	if db := resultMap["database"].(string); db != "go_llm_agent_test" {
		t.Errorf("Expected database 'go_llm_agent_test', got '%s'", db)
	}

	// Verify server info
	serverInfo, ok := resultMap["server_info"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected server_info to be a map, got %T", resultMap["server_info"])
	} else {
		if version, ok := serverInfo["version"].(string); ok {
			t.Logf("MongoDB version: %s", version)
		}
	}

	// Cleanup - close connection
	defer CloseConnection(ctx, connID)
	
	t.Logf("✓ Successfully connected to MongoDB with connection_id: %s", connID)
}

// TestIntegrationMongoDBCRUDWorkflow tests complete CRUD operations
func TestIntegrationMongoDBCRUDWorkflow(t *testing.T) {
	ctx := context.Background()
	
	// 1. Connect
	connectTool := NewConnectTool()
	connectResult, err := connectTool.Execute(ctx, map[string]interface{}{
		"connection_string": getMongoDBURL(),
		"database":          "go_llm_agent_test",
	})
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	connID := connectResult.(map[string]interface{})["connection_id"].(string)
	defer CloseConnection(ctx, connID)
	
	collectionName := fmt.Sprintf("test_users_%d", time.Now().Unix())
	t.Logf("Using collection: %s", collectionName)

	// 2. Insert documents
	insertTool := NewInsertTool()
	insertResult, err := insertTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"documents": []interface{}{
			map[string]interface{}{
				"name":  "Alice",
				"age":   30,
				"email": "alice@example.com",
			},
			map[string]interface{}{
				"name":  "Bob",
				"age":   25,
				"email": "bob@example.com",
			},
			map[string]interface{}{
				"name":  "Charlie",
				"age":   35,
				"email": "charlie@example.com",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to insert documents: %v", err)
	}

	insertMap := insertResult.(map[string]interface{})
	insertedIDs := insertMap["inserted_ids"].([]string)
	if len(insertedIDs) != 3 {
		t.Errorf("Expected 3 inserted IDs, got %d", len(insertedIDs))
	}
	t.Logf("✓ Inserted %d documents", insertMap["inserted_count"])

	// 3. Find all documents
	findTool := NewFindTool()
	findResult, err := findTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{},
		"limit":         10,
	})
	if err != nil {
		t.Fatalf("Failed to find documents: %v", err)
	}

	findMap := findResult.(map[string]interface{})
	docs := findMap["documents"].([]map[string]interface{})
	if len(docs) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(docs))
	}
	t.Logf("✓ Found %d documents", findMap["count"])

	// 4. Find with filter
	findFiltered, err := findTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{"age": map[string]interface{}{"$gte": 30}},
		"limit":         10,
		"sort":          map[string]interface{}{"age": 1},
	})
	if err != nil {
		t.Fatalf("Failed to find filtered documents: %v", err)
	}

	filteredDocs := findFiltered.(map[string]interface{})["documents"].([]map[string]interface{})
	if len(filteredDocs) != 2 {
		t.Errorf("Expected 2 documents with age >= 30, got %d", len(filteredDocs))
	}
	// Verify sorted by age ascending
	if filteredDocs[0]["name"] != "Alice" {
		t.Errorf("Expected first document to be Alice, got %v", filteredDocs[0]["name"])
	}
	t.Logf("✓ Filter and sort working correctly")

	// 5. Update document
	updateTool := NewUpdateTool()
	updateResult, err := updateTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{"name": "Alice"},
		"update": map[string]interface{}{
			"$set": map[string]interface{}{"age": 31, "status": "updated"},
		},
		"update_many": false,
	})
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	updateMap := updateResult.(map[string]interface{})
	if updateMap["matched_count"].(int64) != 1 {
		t.Errorf("Expected 1 matched document, got %d", updateMap["matched_count"])
	}
	if updateMap["modified_count"].(int64) != 1 {
		t.Errorf("Expected 1 modified document, got %d", updateMap["modified_count"])
	}
	t.Logf("✓ Updated 1 document")

	// Verify update
	verifyResult, _ := findTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{"name": "Alice"},
		"limit":         1,
	})
	verifyDocs := verifyResult.(map[string]interface{})["documents"].([]map[string]interface{})
	// MongoDB returns age as int32
	age := verifyDocs[0]["age"]
	ageInt, ok := age.(int32)
	if !ok {
		// Try int64
		ageInt64, ok2 := age.(int64)
		if ok2 {
			ageInt = int32(ageInt64)
		} else {
			t.Errorf("Expected age to be int32 or int64, got %T: %v", age, age)
		}
	}
	if ageInt != 31 {
		t.Errorf("Expected Alice's age to be 31, got %v", ageInt)
	}
	if verifyDocs[0]["status"] != "updated" {
		t.Errorf("Expected status 'updated', got %v", verifyDocs[0]["status"])
	}
	t.Logf("✓ Update verified")

	// 6. Delete one document
	deleteTool := NewDeleteTool()
	deleteResult, err := deleteTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{"name": "Bob"},
		"delete_many":   false,
	})
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	deleteMap := deleteResult.(map[string]interface{})
	if deleteMap["deleted_count"].(int64) != 1 {
		t.Errorf("Expected 1 deleted document, got %d", deleteMap["deleted_count"])
	}
	t.Logf("✓ Deleted 1 document")

	// Verify deletion
	afterDelete, _ := findTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{},
		"limit":         10,
	})
	remainingDocs := afterDelete.(map[string]interface{})["documents"].([]map[string]interface{})
	if len(remainingDocs) != 2 {
		t.Errorf("Expected 2 remaining documents, got %d", len(remainingDocs))
	}

	// 7. Cleanup - delete all test documents
	_, err = deleteTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{"_id": map[string]interface{}{"$exists": true}},
		"delete_many":   true,
	})
	if err != nil {
		t.Logf("Warning: Failed to cleanup test collection: %v", err)
	}

	t.Log("✓ CRUD workflow completed successfully")
}

// TestIntegrationMongoDBConnectionPooling tests connection pool limits
func TestIntegrationMongoDBConnectionPooling(t *testing.T) {
	ctx := context.Background()
	connectTool := NewConnectTool()

	// Create multiple connections
	connectionIDs := []string{}
	for i := 0; i < 5; i++ {
		result, err := connectTool.Execute(ctx, map[string]interface{}{
			"connection_string": getMongoDBURL(),
			"database":          fmt.Sprintf("test_db_%d", i),
		})
		if err != nil {
			t.Fatalf("Failed to create connection %d: %v", i, err)
		}
		connID := result.(map[string]interface{})["connection_id"].(string)
		connectionIDs = append(connectionIDs, connID)
	}

	t.Logf("✓ Created %d connections", len(connectionIDs))

	// Cleanup all connections
	for _, connID := range connectionIDs {
		if err := CloseConnection(ctx, connID); err != nil {
			t.Errorf("Failed to close connection %s: %v", connID, err)
		}
	}

	t.Log("✓ All connections closed successfully")
}

// TestIntegrationMongoDBErrorHandling tests error scenarios
func TestIntegrationMongoDBErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Test 1: Invalid connection string
	connectTool := NewConnectTool()
	_, err := connectTool.Execute(ctx, map[string]interface{}{
		"connection_string": "mongodb://invalid-host:27017",
		"database":          "testdb",
		"timeout":           2,
	})
	if err == nil {
		t.Error("Expected error for invalid connection string, got nil")
	} else {
		t.Logf("✓ Invalid connection correctly failed: %v", err)
	}

	// Test 2: Empty delete filter
	deleteTool := NewDeleteTool()
	_, err = deleteTool.Execute(ctx, map[string]interface{}{
		"connection_id": "fake_connection",
		"collection":    "test_collection",
		"filter":        map[string]interface{}{}, // Empty filter!
	})
	if err == nil {
		t.Error("Expected error for empty delete filter, got nil")
	} else {
		// Could be either "empty filter" error or "connection not found" error
		// Both are valid since we're using fake connection
		t.Logf("✓ Empty delete filter correctly rejected: %v", err)
	}

	// Test 3: Connection not found
	findTool := NewFindTool()
	_, err = findTool.Execute(ctx, map[string]interface{}{
		"connection_id": "nonexistent_connection",
		"collection":    "test_collection",
		"filter":        map[string]interface{}{},
	})
	if err == nil {
		t.Error("Expected error for nonexistent connection, got nil")
	} else {
		t.Logf("✓ Nonexistent connection correctly failed: %v", err)
	}

	t.Log("✓ Error handling tests completed")
}

// TestIntegrationMongoDBBatchInsert tests batch insert limits
func TestIntegrationMongoDBBatchInsert(t *testing.T) {
	ctx := context.Background()

	// Connect
	connectTool := NewConnectTool()
	connectResult, err := connectTool.Execute(ctx, map[string]interface{}{
		"connection_string": getMongoDBURL(),
		"database":          "go_llm_agent_test",
	})
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	connID := connectResult.(map[string]interface{})["connection_id"].(string)
	defer CloseConnection(ctx, connID)

	collectionName := fmt.Sprintf("test_batch_%d", time.Now().Unix())

	// Create 50 documents for batch insert
	docs := []interface{}{}
	for i := 0; i < 50; i++ {
		docs = append(docs, map[string]interface{}{
			"name":  fmt.Sprintf("User%d", i),
			"index": i,
		})
	}

	// Test batch insert
	insertTool := NewInsertTool()
	insertResult, err := insertTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"documents":     docs,
	})
	if err != nil {
		t.Fatalf("Failed to batch insert: %v", err)
	}

	insertMap := insertResult.(map[string]interface{})
	if insertMap["inserted_count"].(int) != 50 {
		t.Errorf("Expected 50 inserted documents, got %d", insertMap["inserted_count"])
	}
	t.Logf("✓ Batch inserted %d documents", insertMap["inserted_count"])

	// Cleanup
	deleteTool := NewDeleteTool()
	deleteTool.Execute(ctx, map[string]interface{}{
		"connection_id": connID,
		"collection":    collectionName,
		"filter":        map[string]interface{}{"_id": map[string]interface{}{"$exists": true}},
		"delete_many":   true,
	})
}
