package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// UpdateTool implements MongoDB update functionality
type UpdateTool struct {
	tools.BaseTool
}

// NewUpdateTool creates a new MongoDB update tool
func NewUpdateTool() *UpdateTool {
	return &UpdateTool{
		BaseTool: tools.NewBaseTool(
			"mongodb_update",
			"Update documents in a MongoDB collection. Returns number of matched and modified documents. Parameters: connection_id (required, from mongodb_connect), collection (required), filter (required, query filter to match documents), update (required, update operations as JSON object with operators like $set, $inc, etc.), update_many (optional, boolean, if true updates all matching documents, default false for single document).",
			tools.CategoryDatabase,
			false, // Doesn't require auth
			false, // NOT safe - modifies data
		),
	}
}

// Parameters implements Tool.Parameters
func (t *UpdateTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type:     "object",
		Required: []string{"connection_id", "collection", "filter", "update"},
		Properties: map[string]*types.JSONSchema{
			"connection_id": {
				Type:        "string",
				Description: "Connection ID from mongodb_connect",
			},
			"collection": {
				Type:        "string",
				Description: "Name of the collection to update",
			},
			"filter": {
				Type:                 "object",
				Description:          "Query filter to match documents (e.g., {\"age\": {\"$gt\": 25}})",
				AdditionalProperties: true, // Allow any MongoDB query operators
			},
			"update": {
				Type:                 "object",
				Description:          "Update operations using MongoDB operators (e.g., {\"$set\": {\"status\": \"active\"}})",
				AdditionalProperties: true, // Allow any MongoDB update operators
			},
			"update_many": {
				Type:        "boolean",
				Description: "If true, updates all matching documents. If false (default), updates only the first match.",
			},
		},
	}
}

// Execute implements Tool.Execute
func (t *UpdateTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
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

	// Extract update
	update, ok := params["update"].(map[string]interface{})
	if !ok {
		return nil, errors.New("update is required and must be an object")
	}

	// Extract update_many flag (default: false)
	updateMany := false
	if um, ok := params["update_many"].(bool); ok {
		updateMany = um
	}

	// Get collection
	collection := conn.Client.Database(conn.Database).Collection(collectionName)

	// Create context with timeout
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var matchedCount, modifiedCount int64

	if updateMany {
		// Update all matching documents
		result, err := collection.UpdateMany(updateCtx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update documents: %w", err)
		}
		matchedCount = result.MatchedCount
		modifiedCount = result.ModifiedCount
	} else {
		// Update single document
		result, err := collection.UpdateOne(updateCtx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update document: %w", err)
		}
		matchedCount = result.MatchedCount
		modifiedCount = result.ModifiedCount
	}

	return map[string]interface{}{
		"matched_count":  matchedCount,
		"modified_count": modifiedCount,
		"update_many":    updateMany,
		"filter":         filter,
		"update":         update,
	}, nil
}
