package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindTool implements MongoDB find/query functionality
type FindTool struct {
	tools.BaseTool
}

// NewFindTool creates a new MongoDB find tool
func NewFindTool() *FindTool {
	return &FindTool{
		BaseTool: tools.NewBaseTool(
			"mongodb_find",
			"Query documents from a MongoDB collection. Returns matching documents. Parameters: connection_id (required, from mongodb_connect), collection (required), filter (optional, query filter as JSON object, default {}), limit (optional, max documents to return, default 10, max 1000), sort (optional, sort specification as JSON object), projection (optional, fields to include/exclude as JSON object).",
			tools.CategoryDatabase,
			false, // Doesn't require auth
			true,  // Safe operation (read-only)
		),
	}
}

// Parameters implements Tool.Parameters
func (t *FindTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type:     "object",
		Required: []string{"connection_id", "collection"},
		Properties: map[string]*types.JSONSchema{
			"connection_id": {
				Type:        "string",
				Description: "Connection ID from mongodb_connect",
			},
			"collection": {
				Type:        "string",
				Description: "Name of the collection to query",
			},
			"filter": {
				Type:                 "object",
				Description:          "MongoDB query filter (e.g., {\"age\": {\"$gt\": 25}} or {\"name\": \"John\"}). Default: {} (all documents)",
				AdditionalProperties: true, // Allow any MongoDB query operators
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of documents to return (default: 10, max: 1000)",
			},
			"sort": {
				Type:                 "object",
				Description:          "Sort specification (e.g., {\"age\": -1} for descending, {\"name\": 1} for ascending)",
				AdditionalProperties: true, // Allow any field names
			},
			"projection": {
				Type:                 "object",
				Description:          "Fields to include or exclude (e.g., {\"name\": 1, \"age\": 1, \"_id\": 0})",
				AdditionalProperties: true, // Allow any field names
			},
		},
	}
}

// Execute implements Tool.Execute
func (t *FindTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract connection ID
	connectionID, ok := params["connection_id"].(string)
	if !ok || connectionID == "" {
		return nil, errors.New("connection_id is required and must be a non-empty string")
	}

	// Get connection from pool
	conn, err := GetConnection(connectionID)
	if err != nil {
		return nil, err
	}

	// Extract collection name
	collectionName, ok := params["collection"].(string)
	if !ok || collectionName == "" {
		return nil, errors.New("collection is required and must be a non-empty string")
	}

	// Extract filter (default: empty filter = all documents)
	filter := bson.M{}
	if filterParam, ok := params["filter"].(map[string]interface{}); ok {
		filter = filterParam
	}

	// Extract limit (default: 10, max: 1000)
	limit := int64(10)
	if limitParam, ok := params["limit"].(float64); ok {
		limit = int64(limitParam)
		if limit < 1 {
			limit = 1
		}
		if limit > 1000 {
			limit = 1000
		}
	}

	// Build find options
	findOptions := options.Find().SetLimit(limit)

	// Extract sort
	if sortParam, ok := params["sort"].(map[string]interface{}); ok {
		findOptions.SetSort(sortParam)
	}

	// Extract projection
	if projectionParam, ok := params["projection"].(map[string]interface{}); ok {
		findOptions.SetProjection(projectionParam)
	}

	// Get collection
	collection := conn.Client.Database(conn.Database).Collection(collectionName)

	// Execute find with timeout
	findCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(findCtx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to execute find: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode all results
	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	// Convert ObjectIDs to strings for better JSON serialization
	for _, doc := range results {
		convertBSONTypes(doc)
	}

	return map[string]interface{}{
		"documents": results,
		"count":     len(results),
		"filter":    filter,
		"limit":     limit,
	}, nil
}

// convertBSONTypes converts BSON types to JSON-friendly types
func convertBSONTypes(doc map[string]interface{}) {
	for key, value := range doc {
		switch v := value.(type) {
		case primitive.ObjectID:
			doc[key] = v.Hex()
		case map[string]interface{}:
			convertBSONTypes(v)
		case []interface{}:
			for i, item := range v {
				if nested, ok := item.(map[string]interface{}); ok {
					convertBSONTypes(nested)
				} else if objID, ok := item.(primitive.ObjectID); ok {
					v[i] = objID.Hex()
				}
			}
		}
	}
}
