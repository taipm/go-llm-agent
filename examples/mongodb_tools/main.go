package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/taipm/go-llm-agent/pkg/tools/database/mongodb"
)

func main() {
	fmt.Println("=== MongoDB Tools Demo ===")

	ctx := context.Background()

	// Demo 1: Create Tools
	fmt.Println("1. Creating MongoDB tools...")
	connectTool := mongodb.NewConnectTool()
	findTool := mongodb.NewFindTool()
	insertTool := mongodb.NewInsertTool()
	updateTool := mongodb.NewUpdateTool()
	deleteTool := mongodb.NewDeleteTool()

	fmt.Printf("   ✓ Created 5 MongoDB tools\n")
	fmt.Printf("     - %s (safe: %v)\n", connectTool.Name(), connectTool.IsSafe())
	fmt.Printf("     - %s (safe: %v)\n", findTool.Name(), findTool.IsSafe())
	fmt.Printf("     - %s (safe: %v)\n", insertTool.Name(), insertTool.IsSafe())
	fmt.Printf("     - %s (safe: %v)\n", updateTool.Name(), updateTool.IsSafe())
	fmt.Printf("     - %s (safe: %v)\n\n", deleteTool.Name(), deleteTool.IsSafe())

	// Demo 2: Example Connect Usage
	fmt.Println("2. Example: Connect to MongoDB")
	fmt.Println("   Code:")
	fmt.Println(`   result, err := connectTool.Execute(ctx, map[string]interface{}{
       "connection_string": "mongodb://localhost:27017",
       "database": "myapp",
   })`)
	fmt.Println("\n   Expected Response:")
	exampleConnectResult := map[string]interface{}{
		"connection_id": "mongo_1a2b3c4d5e6f7g8h",
		"database":      "myapp",
		"connected_at":  "2024-01-20T10:30:00Z",
		"server_info": map[string]interface{}{
			"version":       "7.0.0",
			"max_bson_size": 16777216,
		},
		"pool_size": 1,
	}
	connectJSON, _ := json.MarshalIndent(exampleConnectResult, "   ", "  ")
	fmt.Printf("   %s\n\n", string(connectJSON))

	// Demo 3: Example Find Usage
	fmt.Println("3. Example: Query Documents")
	fmt.Println("   Code:")
	fmt.Println(`   result, err := findTool.Execute(ctx, map[string]interface{}{
       "connection_id": "mongo_1a2b3c4d5e6f7g8h",
       "collection": "users",
       "filter": map[string]interface{}{"age": map[string]interface{}{"$gt": 25}},
       "limit": 5,
   })`)
	fmt.Println("\n   Expected Response:")
	exampleFindResult := map[string]interface{}{
		"documents": []map[string]interface{}{
			{"_id": "507f191e810c19729de860ea", "name": "Alice", "age": 30},
			{"_id": "507f191e810c19729de860eb", "name": "Bob", "age": 28},
		},
		"count":  2,
		"filter": map[string]interface{}{"age": map[string]interface{}{"$gt": 25}},
		"limit":  5,
	}
	findJSON, _ := json.MarshalIndent(exampleFindResult, "   ", "  ")
	fmt.Printf("   %s\n\n", string(findJSON))

	// Demo 4: Example Insert Usage
	fmt.Println("4. Example: Insert Documents")
	fmt.Println("   Code:")
	fmt.Println(`   result, err := insertTool.Execute(ctx, map[string]interface{}{
       "connection_id": "mongo_1a2b3c4d5e6f7g8h",
       "collection": "users",
       "documents": []interface{}{
           map[string]interface{}{"name": "Charlie", "age": 35},
       },
   })`)
	fmt.Println("\n   Expected Response:")
	exampleInsertResult := map[string]interface{}{
		"inserted_ids":   []string{"507f191e810c19729de860ec"},
		"inserted_count": 1,
		"collection":     "users",
	}
	insertJSON, _ := json.MarshalIndent(exampleInsertResult, "   ", "  ")
	fmt.Printf("   %s\n\n", string(insertJSON))

	// Demo 5: Example Update Usage
	fmt.Println("5. Example: Update Documents")
	fmt.Println("   Code:")
	fmt.Println(`   result, err := updateTool.Execute(ctx, map[string]interface{}{
       "connection_id": "mongo_1a2b3c4d5e6f7g8h",
       "collection": "users",
       "filter": map[string]interface{}{"name": "Alice"},
       "update": map[string]interface{}{"$set": map[string]interface{}{"age": 31}},
   })`)
	fmt.Println("\n   Expected Response:")
	exampleUpdateResult := map[string]interface{}{
		"matched_count":  1,
		"modified_count": 1,
		"update_many":    false,
	}
	updateJSON, _ := json.MarshalIndent(exampleUpdateResult, "   ", "  ")
	fmt.Printf("   %s\n\n", string(updateJSON))

	// Demo 6: Example Delete Usage
	fmt.Println("6. Example: Delete Documents")
	fmt.Println("   Code:")
	fmt.Println(`   result, err := deleteTool.Execute(ctx, map[string]interface{}{
       "connection_id": "mongo_1a2b3c4d5e6f7g8h",
       "collection": "users",
       "filter": map[string]interface{}{"status": "inactive"},
   })`)
	fmt.Println("\n   Expected Response:")
	exampleDeleteResult := map[string]interface{}{
		"deleted_count": 3,
		"delete_many":   false,
	}
	deleteJSON, _ := json.MarshalIndent(exampleDeleteResult, "   ", "  ")
	fmt.Printf("   %s\n\n", string(deleteJSON))

	// Demo 7: Error Handling
	fmt.Println("7. Error Handling: Empty Delete Filter")
	_, err := deleteTool.Execute(ctx, map[string]interface{}{
		"connection_id": "test",
		"collection":    "users",
		"filter":        map[string]interface{}{},
	})
	if err != nil {
		fmt.Printf("   ✓ Error caught: %v\n\n", err)
	}

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("MongoDB tools provide:")
	fmt.Println("  ✓ Connection pooling (max 10)")
	fmt.Println("  ✓ Full CRUD operations")
	fmt.Println("  ✓ Query filtering & sorting")
	fmt.Println("  ✓ Batch inserts (up to 100)")
	fmt.Println("  ✓ Safety checks")
	fmt.Println("\nTo use with real MongoDB:")
	fmt.Println("  docker run -d -p 27017:27017 mongo")

	log.Println("Demo completed!")
}
