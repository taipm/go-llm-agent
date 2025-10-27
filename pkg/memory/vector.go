package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// VectorMemory implements AdvancedMemory with Qdrant vector database
type VectorMemory struct {
	client         *qdrant.Client
	collectionName string
	embedder       Embedder
	cache          *BufferMemory // Hot cache for recent messages
	dims           int
}

// VectorMemoryConfig holds configuration for VectorMemory
type VectorMemoryConfig struct {
	QdrantURL      string
	CollectionName string
	Embedder       Embedder
	CacheSize      int
}

// NewVectorMemory creates a new vector memory with Qdrant
func NewVectorMemory(ctx context.Context, config VectorMemoryConfig) (*VectorMemory, error) {
	if config.QdrantURL == "" {
		config.QdrantURL = "localhost:6334"
	}
	if config.CollectionName == "" {
		config.CollectionName = "agent_memory"
	}
	if config.Embedder == nil {
		config.Embedder = NewOllamaEmbedder("", "") // Default Ollama embedder
	}
	if config.CacheSize <= 0 {
		config.CacheSize = 100
	}

	// Connect to Qdrant
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: config.QdrantURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	vm := &VectorMemory{
		client:         client,
		collectionName: config.CollectionName,
		embedder:       config.Embedder,
		cache:          NewBuffer(config.CacheSize),
		dims:           config.Embedder.Dimensions(),
	}

	// Create collection if it doesn't exist
	if err := vm.ensureCollection(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup collection: %w", err)
	}

	return vm, nil
}

// ensureCollection creates the collection if it doesn't exist
func (v *VectorMemory) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	exists, err := v.client.CollectionExists(ctx, v.collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if exists {
		return nil // Collection already exists
	}

	// Create collection with vector configuration
	err = v.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: v.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(v.dims),
			Distance: qdrant.Distance_Cosine,
		}),
	})

	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

// Add implements types.Memory interface
func (v *VectorMemory) Add(message types.Message) error {
	ctx := context.Background()
	return v.AddWithEmbedding(ctx, message, nil)
}

// AddWithEmbedding implements types.AdvancedMemory interface
func (v *VectorMemory) AddWithEmbedding(ctx context.Context, message types.Message, embedding []float32) error {
	// Add to hot cache
	v.cache.Add(message)

	// Generate embedding if not provided
	if embedding == nil {
		emb, err := v.embedder.Embed(ctx, message.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		embedding = emb
	}

	// Prepare payload with message metadata
	payload := map[string]interface{}{
		"role":       string(message.Role),
		"content":    message.Content,
		"timestamp":  time.Now().Unix(),
		"tool_calls": message.ToolCalls,
		"tool_id":    message.ToolID,
	}

	// Add metadata if present
	if message.Metadata != nil {
		for k, v := range message.Metadata {
			payload[k] = v
		}
	}

	// Convert payload to JSON for Qdrant
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payloadMap); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create point ID
	pointID := uuid.New().String()

	// Upsert point to Qdrant
	_, err = v.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: v.collectionName,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewID(pointID),
				Vectors: qdrant.NewVectors(embedding...),
				Payload: qdrant.NewValueMap(payloadMap),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upsert point: %w", err)
	}

	return nil
}

// GetHistory implements types.Memory interface
func (v *VectorMemory) GetHistory(limit int) ([]types.Message, error) {
	// Use hot cache for recent messages
	return v.cache.GetHistory(limit)
}

// Clear implements types.Memory interface
func (v *VectorMemory) Clear() error {
	ctx := context.Background()

	// Clear cache
	v.cache.Clear()

	// Delete and recreate collection
	err := v.client.DeleteCollection(ctx, v.collectionName)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return v.ensureCollection(ctx)
}

// Size implements types.Memory interface
func (v *VectorMemory) Size() int {
	return v.cache.Size()
}

// SearchSemantic implements types.AdvancedMemory interface
func (v *VectorMemory) SearchSemantic(ctx context.Context, query string, limit int) ([]types.Message, error) {
	// Generate query embedding
	queryEmbedding, err := v.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search in Qdrant
	searchResult, err := v.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: v.collectionName,
		Query:          qdrant.NewQuery(queryEmbedding...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	// Convert results to messages
	messages := make([]types.Message, 0, len(searchResult))
	for _, point := range searchResult {
		msg, err := v.pointToMessage(point)
		if err != nil {
			continue // Skip invalid points
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetByCategory implements types.AdvancedMemory interface
func (v *VectorMemory) GetByCategory(ctx context.Context, category types.MessageCategory, limit int) ([]types.Message, error) {
	// Create filter for category
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewMatch("category", string(category)),
		},
	}

	// Scroll through points with filter
	scrollResult, err := v.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: v.collectionName,
		Filter:         filter,
		Limit:          qdrant.PtrOf(uint32(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get by category: %w", err)
	}

	// Convert results to messages
	messages := make([]types.Message, 0, len(scrollResult))
	for _, point := range scrollResult {
		msg, err := v.retrievedPointToMessage(point)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetMostImportant implements types.AdvancedMemory interface
func (v *VectorMemory) GetMostImportant(ctx context.Context, limit int) ([]types.Message, error) {
	// Scroll all points and sort by importance
	scrollResult, err := v.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: v.collectionName,
		Limit:          qdrant.PtrOf(uint32(limit * 2)), // Get more, then filter
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get points: %w", err)
	}

	// Convert and sort by importance
	type msgWithImportance struct {
		msg        types.Message
		importance float64
	}

	candidates := make([]msgWithImportance, 0)
	for _, point := range scrollResult {
		msg, err := v.retrievedPointToMessage(point)
		if err != nil {
			continue
		}

		importance := 0.0
		if msg.Metadata != nil {
			if imp, ok := msg.Metadata["importance"].(float64); ok {
				importance = imp
			}
		}

		candidates = append(candidates, msgWithImportance{msg, importance})
	}

	// Simple sort by importance (descending)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].importance > candidates[i].importance {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Return top N
	messages := make([]types.Message, 0, limit)
	for i := 0; i < limit && i < len(candidates); i++ {
		messages = append(messages, candidates[i].msg)
	}

	return messages, nil
}

// HybridSearch implements types.AdvancedMemory interface
func (v *VectorMemory) HybridSearch(ctx context.Context, query string, limit int) ([]types.Message, error) {
	// For now, use semantic search (can be enhanced with BM25 later)
	return v.SearchSemantic(ctx, query, limit)
}

// GetStats implements types.AdvancedMemory interface
func (v *VectorMemory) GetStats(ctx context.Context) (*types.MemoryStats, error) {
	// Get collection info
	collectionInfo, err := v.client.GetCollectionInfo(ctx, v.collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	stats := &types.MemoryStats{
		TotalMessages: int(collectionInfo.GetPointsCount()),
		VectorCount:   int(collectionInfo.GetVectorsCount()),
	}

	return stats, nil
}

// Archive implements types.AdvancedMemory interface
func (v *VectorMemory) Archive(ctx context.Context, olderThan time.Duration) error {
	// Filter points older than threshold
	cutoff := time.Now().Add(-olderThan).Unix()

	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewRange("timestamp", &qdrant.Range{
				Lt: qdrant.PtrOf(float64(cutoff)),
			}),
		},
	}

	// Delete old points
	_, err := v.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: v.collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
				Filter: filter,
			},
		},
	})

	return err
}

// Export implements types.AdvancedMemory interface
func (v *VectorMemory) Export(ctx context.Context, path string) error {
	// Create snapshot
	snapshotResult, err := v.client.CreateSnapshot(ctx, v.collectionName)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Snapshot name is returned
	_ = snapshotResult.GetName()

	// Note: Actual file download would require additional Qdrant API calls
	// This is a simplified version
	return nil
}

// pointToMessage converts Qdrant point to Message
func (v *VectorMemory) pointToMessage(point *qdrant.ScoredPoint) (types.Message, error) {
	payload := point.GetPayload()

	msg := types.Message{
		Metadata: make(map[string]interface{}),
	}

	// Extract role
	if roleVal := payload["role"]; roleVal != nil {
		if roleStr, ok := roleVal.GetKind().(*qdrant.Value_StringValue); ok {
			msg.Role = types.Role(roleStr.StringValue)
		}
	}

	// Extract content
	if contentVal := payload["content"]; contentVal != nil {
		if contentStr, ok := contentVal.GetKind().(*qdrant.Value_StringValue); ok {
			msg.Content = contentStr.StringValue
		}
	}

	// Extract all other fields as metadata
	for key, val := range payload {
		if key != "role" && key != "content" {
			msg.Metadata[key] = extractValue(val)
		}
	}

	return msg, nil
}

// retrievedPointToMessage converts RetrievedPoint to Message
func (v *VectorMemory) retrievedPointToMessage(point *qdrant.RetrievedPoint) (types.Message, error) {
	payload := point.GetPayload()

	msg := types.Message{
		Metadata: make(map[string]interface{}),
	}

	// Extract role
	if roleVal := payload["role"]; roleVal != nil {
		if roleStr, ok := roleVal.GetKind().(*qdrant.Value_StringValue); ok {
			msg.Role = types.Role(roleStr.StringValue)
		}
	}

	// Extract content
	if contentVal := payload["content"]; contentVal != nil {
		if contentStr, ok := contentVal.GetKind().(*qdrant.Value_StringValue); ok {
			msg.Content = contentStr.StringValue
		}
	}

	// Extract all other fields as metadata
	for key, val := range payload {
		if key != "role" && key != "content" {
			msg.Metadata[key] = extractValue(val)
		}
	}

	return msg, nil
}

// extractValue converts Qdrant value to Go value
func extractValue(val *qdrant.Value) interface{} {
	if val == nil {
		return nil
	}

	switch v := val.GetKind().(type) {
	case *qdrant.Value_StringValue:
		return v.StringValue
	case *qdrant.Value_IntegerValue:
		return v.IntegerValue
	case *qdrant.Value_DoubleValue:
		return v.DoubleValue
	case *qdrant.Value_BoolValue:
		return v.BoolValue
	default:
		return nil
	}
}

// Close closes the Qdrant client connection
func (v *VectorMemory) Close() error {
	if v.client != nil {
		return v.client.Close()
	}
	return nil
}
