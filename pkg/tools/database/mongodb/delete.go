package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// DeleteTool implements MongoDB delete functionality
type DeleteTool struct {
	tools.BaseTool
}

// NewDeleteTool creates a new MongoDB delete tool
func NewDeleteTool() *DeleteTool {
	return &DeleteTool{
		BaseTool: tools.NewBaseTool(
			"mongodb_delete",
			"Delete documents from a MongoDB collection. WARNING: This is a destructive operation. Returns number of deleted documents. Parameters: connection_id (required, from mongodb_connect), collection (required), filter (required, query filter to match documents to delete), delete_many (optional, boolean, if true deletes all matching documents, default false for single document).",
			tools.CategoryDatabase,
			false, // Doesn't require auth
			false, // NOT safe - destructive operation
		),
	}
}

// Parameters implements Tool.Parameters
func (t *DeleteTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type:     "object",
		Required: []string{"connection_id", "collection", "filter"},
		Properties: map[string]*types.JSONSchema{
			"connection_id": {
				Type:        "string",
				Description: "Connection ID from mongodb_connect",
			},
			"collection": {
				Type:        "string",
				Description: "Name of the collection to delete from",
			},
			"filter": {
				Type:                 "object",
				Description:          "Query filter to match documents to delete (e.g., {\"status\": \"inactive\"}). Use {} with caution as it matches all documents!",
				AdditionalProperties: true, // Allow any MongoDB query operators
			},
			"delete_many": {
				Type:        "boolean",
				Description: "If true, deletes all matching documents. If false (default), deletes only the first match.",
			},
		},
	}
}

// Execute implements Tool.Execute
func (t *DeleteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
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

	// Extract filter
	filter, ok := params["filter"].(map[string]interface{})
	if !ok {
		return nil, errors.New("filter is required and must be an object")
	}

	// Safety check: warn if filter is empty (deletes all!)
	if len(filter) == 0 {
		return nil, errors.New("empty filter would delete all documents. Please provide a specific filter or use {\"_id\": {\"$exists\": true}} to explicitly delete all")
	}

	// Extract delete_many flag (default: false)
	deleteMany := false
	if dm, ok := params["delete_many"].(bool); ok {
		deleteMany = dm
	}

	// Get collection
	collection := conn.Client.Database(conn.Database).Collection(collectionName)

	// Create context with timeout
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var deletedCount int64

	if deleteMany {
		// Delete all matching documents
		result, err := collection.DeleteMany(deleteCtx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete documents: %w", err)
		}
		deletedCount = result.DeletedCount
	} else {
		// Delete single document
		result, err := collection.DeleteOne(deleteCtx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete document: %w", err)
		}
		deletedCount = result.DeletedCount
	}

	return map[string]interface{}{
		"deleted_count": deletedCount,
		"delete_many":   deleteMany,
		"filter":        filter,
		"collection":    collectionName,
	}, nil
}
