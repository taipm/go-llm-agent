package mongodb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connection pool to manage MongoDB connections
var (
	connectionPool = sync.Map{}
	maxConnections = 10
	poolSize       = 0
	poolMutex      = sync.Mutex{}
)

// MongoConnection represents a stored MongoDB connection
type MongoConnection struct {
	Client      *mongo.Client
	Database    string
	ConnectedAt time.Time
}

// ConnectTool implements MongoDB connection functionality
type ConnectTool struct {
	tools.BaseTool
}

// NewConnectTool creates a new MongoDB connection tool
func NewConnectTool() *ConnectTool {
	return &ConnectTool{
		BaseTool: tools.NewBaseTool(
			"mongodb_connect",
			"Connect to a MongoDB database. Returns a connection_id that can be used with other MongoDB tools. Supports connection pooling (max 10 connections). Parameters: connection_string (required, MongoDB URI), database (required, database name), timeout (optional, connection timeout in seconds, default 10, max 60).",
			tools.CategoryDatabase,
			false, // Doesn't require auth
			true,  // Safe operation (read-only connection establishment)
		),
	}
}

// Parameters implements Tool.Parameters
func (t *ConnectTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type:     "object",
		Required: []string{"connection_string", "database"},
		Properties: map[string]*types.JSONSchema{
			"connection_string": {
				Type:        "string",
				Description: "MongoDB connection URI (e.g., mongodb://localhost:27017 or mongodb+srv://user:pass@cluster.mongodb.net/). Supports authentication, replica sets, and TLS/SSL.",
			},
			"database": {
				Type:        "string",
				Description: "Name of the database to use",
			},
			"timeout": {
				Type:        "integer",
				Description: "Connection timeout in seconds (default: 10, max: 60)",
			},
		},
	}
}

// Execute implements Tool.Execute
func (t *ConnectTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract connection string
	connectionString, ok := params["connection_string"].(string)
	if !ok || connectionString == "" {
		return nil, errors.New("connection_string is required and must be a non-empty string")
	}

	// Extract database name
	database, ok := params["database"].(string)
	if !ok || database == "" {
		return nil, errors.New("database is required and must be a non-empty string")
	}

	// Extract timeout (default: 10 seconds, max: 60)
	timeout := 10
	if t, ok := params["timeout"].(float64); ok {
		timeout = int(t)
		if timeout < 1 {
			timeout = 1
		}
		if timeout > 60 {
			timeout = 60
		}
	}

	// Check connection pool size
	poolMutex.Lock()
	if poolSize >= maxConnections {
		poolMutex.Unlock()
		return nil, fmt.Errorf("connection pool is full (max %d connections). Close existing connections before creating new ones", maxConnections)
	}
	poolSize++
	poolMutex.Unlock()

	// Create context with timeout
	connectCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Configure client options
	clientOptions := options.Client().ApplyURI(connectionString)

	// Connect to MongoDB
	client, err := mongo.Connect(connectCtx, clientOptions)
	if err != nil {
		poolMutex.Lock()
		poolSize--
		poolMutex.Unlock()
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		client.Disconnect(ctx)
		poolMutex.Lock()
		poolSize--
		poolMutex.Unlock()
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Generate unique connection ID
	connectionID := generateConnectionID()

	// Store connection in pool
	connectionPool.Store(connectionID, &MongoConnection{
		Client:      client,
		Database:    database,
		ConnectedAt: time.Now(),
	})

	// Get server info
	serverInfo, err := getServerInfo(ctx, client)
	if err != nil {
		// Non-fatal, just log
		serverInfo = map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"connection_id": connectionID,
		"database":      database,
		"connected_at":  time.Now().Format(time.RFC3339),
		"server_info":   serverInfo,
		"pool_size":     poolSize,
	}, nil
}

// GetConnection retrieves a connection from the pool
func GetConnection(connectionID string) (*MongoConnection, error) {
	conn, ok := connectionPool.Load(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection not found: %s. Use mongodb_connect to create a connection first", connectionID)
	}
	return conn.(*MongoConnection), nil
}

// CloseConnection closes and removes a connection from the pool
func CloseConnection(ctx context.Context, connectionID string) error {
	conn, err := GetConnection(connectionID)
	if err != nil {
		return err
	}

	if err := conn.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	connectionPool.Delete(connectionID)
	poolMutex.Lock()
	poolSize--
	poolMutex.Unlock()

	return nil
}

// CloseAllConnections closes all connections in the pool
func CloseAllConnections(ctx context.Context) error {
	var errors []error

	connectionPool.Range(func(key, value interface{}) bool {
		connectionID := key.(string)
		if err := CloseConnection(ctx, connectionID); err != nil {
			errors = append(errors, err)
		}
		return true
	})

	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d connections", len(errors))
	}

	return nil
}

// generateConnectionID creates a unique connection identifier
func generateConnectionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "mongo_" + hex.EncodeToString(bytes)
}

// getServerInfo retrieves server version and other info
func getServerInfo(ctx context.Context, client *mongo.Client) (map[string]interface{}, error) {
	adminDB := client.Database("admin")

	var result bson.M
	err := adminDB.RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{}
	if version, ok := result["version"].(string); ok {
		info["version"] = version
	}
	if gitVersion, ok := result["gitVersion"].(string); ok {
		info["git_version"] = gitVersion
	}
	if maxBsonObjectSize, ok := result["maxBsonObjectSize"].(int32); ok {
		info["max_bson_size"] = maxBsonObjectSize
	}

	return info, nil
}
