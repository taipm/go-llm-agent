package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertTool implements MongoDB insert functionality
type InsertTool struct {
	tools.BaseTool
}

// NewInsertTool creates a new MongoDB insert tool
func NewInsertTool() *InsertTool {
	return &InsertTool{
		BaseTool: tools.NewBaseTool(
			"mongodb_insert",
			"Insert one or more documents into a MongoDB collection. Returns inserted document IDs. Parameters: connection_id (required, from mongodb_connect), collection (required), documents (required, single document object or array of documents, max 100 documents per batch).",
			tools.CategoryDatabase,
			false, // Doesn't require auth
			false, // NOT safe - modifies data
		),
	}
}

// Parameters implements Tool.Parameters
func (t *InsertTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type:     "object",
		Required: []string{"connection_id", "collection", "documents"},
		Properties: map[string]*types.JSONSchema{
			"connection_id": {
				Type:        "string",
				Description: "Connection ID from mongodb_connect",
			},
			"collection": {
				Type:        "string",
				Description: "Name of the collection to insert into",
			},
			"documents": {
				Type:        "array",
				Description: "Array of document objects to insert. Each document is a JSON object. Max 100 documents per batch.",
				Items: &types.JSONSchema{
					Type:                 "object",
					Description:          "A document to insert",
					AdditionalProperties: true, // Allow any properties
				},
			},
		},
	}
}

// Execute implements Tool.Execute
func (t *InsertTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
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

	// Extract documents
	documentsParam, ok := params["documents"]
	if !ok {
		return nil, errors.New("documents is required")
	}

	// Get collection
	collection := conn.Client.Database(conn.Database).Collection(collectionName)

	// Create context with timeout
	insertCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check if single document or array
	var insertedIDs []interface{}
	var insertedCount int

	switch docs := documentsParam.(type) {
	case map[string]interface{}:
		// Single document
		result, err := collection.InsertOne(insertCtx, docs)
		if err != nil {
			return nil, fmt.Errorf("failed to insert document: %w", err)
		}
		insertedIDs = []interface{}{result.InsertedID}
		insertedCount = 1

	case []interface{}:
		// Multiple documents
		if len(docs) == 0 {
			return nil, errors.New("documents array is empty")
		}
		if len(docs) > 100 {
			return nil, fmt.Errorf("too many documents: %d (max 100 per batch)", len(docs))
		}

		result, err := collection.InsertMany(insertCtx, docs)
		if err != nil {
			return nil, fmt.Errorf("failed to insert documents: %w", err)
		}
		insertedIDs = result.InsertedIDs
		insertedCount = len(result.InsertedIDs)

	default:
		return nil, errors.New("documents must be an object or array of objects")
	}

	// Convert ObjectIDs to hex strings
	insertedIDsStr := make([]string, len(insertedIDs))
	for i, id := range insertedIDs {
		if objID, ok := id.(primitive.ObjectID); ok {
			insertedIDsStr[i] = objID.Hex()
		} else {
			insertedIDsStr[i] = fmt.Sprintf("%v", id)
		}
	}

	return map[string]interface{}{
		"inserted_ids":   insertedIDsStr,
		"inserted_count": insertedCount,
		"collection":     collectionName,
	}, nil
}
