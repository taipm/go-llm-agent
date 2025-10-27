# Research: New Built-in Tools - Math, MongoDB, Qdrant

**Date**: October 27, 2025  
**Version**: 0.3.0 Planning  
**Status**: Research Phase

---

## Executive Summary

This document provides research and design specifications for 3 new tool categories to extend go-llm-agent's built-in tools from 13 to 18+ tools:

1. **Math Tools (2 tools)** - Mathematical calculations and statistics
2. **MongoDB Tools (4-5 tools)** - NoSQL database operations
3. **Qdrant Tools (4-5 tools)** - Vector database operations

**Total**: 10-12 new tools across 3 categories

---

## 1. Math Tools (2 tools)

### Overview
Mathematical tools for calculations and statistical operations, enabling agents to perform numerical analysis without external services.

### Tool Category
- **Category**: `CategoryMath`
- **Priority**: MEDIUM (Phase 3 - v0.3.0)
- **Dependencies**: Standard library only
- **Safety**: High (read-only, deterministic)

---

### Tool 1.1: math_calculate

**Purpose**: Perform safe mathematical calculations and expression evaluation

#### Specifications

**Name**: `math_calculate`

**Description**: 
> Safely evaluate mathematical expressions and perform calculations. Supports basic arithmetic, trigonometry, logarithms, and common mathematical functions.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "expression": {
            "type": "string",
            "description": "Mathematical expression to evaluate (e.g., '2 + 2', 'sin(pi/2)', 'sqrt(16)')"
        },
        "precision": {
            "type": "integer",
            "description": "Decimal places for result (optional, default: 6)",
            "minimum": 0,
            "maximum": 15
        },
        "variables": {
            "type": "object",
            "description": "Variable definitions (optional)",
            "additionalProperties": {"type": "number"}
        }
    },
    "required": ["expression"]
}
```

**Returns**:
```json
{
    "type": "calculation_result",
    "expression": "2 + 2 * 3",
    "result": 8.0,
    "precision": 6,
    "variables_used": [],
    "steps": ["2 + (2 * 3)", "2 + 6", "8"]
}
```

#### Implementation Details

**File**: `pkg/tools/math/calculate.go` (~180 lines)

**Key Features**:
- Safe expression parser (prevent code injection)
- Support operators: +, -, *, /, ^, %, sqrt, abs
- Support functions: sin, cos, tan, log, ln, exp
- Support constants: pi, e
- Variable substitution
- Step-by-step calculation trace

**Security Considerations**:
```go
// Whitelist approach - only allow safe operations
var allowedFunctions = map[string]bool{
    "sin": true, "cos": true, "tan": true,
    "sqrt": true, "abs": true, "log": true,
    "ln": true, "exp": true, "ceil": true,
    "floor": true, "round": true,
}

// Validate expression before evaluation
func validateExpression(expr string) error {
    // No function calls except whitelisted
    // No variable assignments
    // No loops or conditionals
    // Maximum expression length: 1000 chars
}
```

**Example Usage**:
```go
calculate, _ := math.NewCalculateTool()

// Simple calculation
result, _ := calculate.Execute(ctx, map[string]interface{}{
    "expression": "2 + 2 * 3",
}) // returns 8

// With variables
result, _ := calculate.Execute(ctx, map[string]interface{}{
    "expression": "x * y + z",
    "variables": map[string]float64{
        "x": 5, "y": 3, "z": 2,
    },
}) // returns 17

// Trigonometry
result, _ := calculate.Execute(ctx, map[string]interface{}{
    "expression": "sin(pi/2)",
    "precision": 4,
}) // returns 1.0000
```

**Test Cases** (~100 lines):
- Basic arithmetic operations
- Order of operations (PEMDAS)
- Function calls (trigonometry, logarithms)
- Variable substitution
- Error handling (division by zero, invalid syntax)
- Security tests (code injection attempts)
- Edge cases (very large/small numbers, infinity, NaN)

---

### Tool 1.2: math_stats

**Purpose**: Calculate statistical measures on datasets

#### Specifications

**Name**: `math_stats`

**Description**: 
> Calculate statistical measures (mean, median, mode, standard deviation, variance) on numerical datasets.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "data": {
            "type": "array",
            "items": {"type": "number"},
            "description": "Numerical dataset",
            "minItems": 1
        },
        "operations": {
            "type": "array",
            "items": {
                "type": "string",
                "enum": ["mean", "median", "mode", "stddev", "variance", "min", "max", "sum", "count"]
            },
            "description": "Statistical operations to perform (default: all)"
        },
        "precision": {
            "type": "integer",
            "description": "Decimal places (default: 6)",
            "minimum": 0,
            "maximum": 15
        }
    },
    "required": ["data"]
}
```

**Returns**:
```json
{
    "type": "statistics_result",
    "count": 10,
    "mean": 55.5,
    "median": 55.0,
    "mode": [60],
    "stddev": 18.027756377319946,
    "variance": 325.0,
    "min": 10,
    "max": 100,
    "sum": 555,
    "quartiles": {
        "q1": 40.0,
        "q2": 55.0,
        "q3": 70.0
    }
}
```

#### Implementation Details

**File**: `pkg/tools/math/statistics.go` (~160 lines)

**Key Features**:
- Mean (average)
- Median (middle value)
- Mode (most frequent)
- Standard deviation
- Variance
- Min/Max
- Sum
- Quartiles (Q1, Q2, Q3)
- Data validation (max size: 10,000 elements)

**Example Usage**:
```go
stats, _ := math.NewStatsTool()

result, _ := stats.Execute(ctx, map[string]interface{}{
    "data": []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
    "operations": []string{"mean", "median", "stddev"},
    "precision": 2,
})
```

**Test Cases** (~90 lines):
- All statistical operations
- Empty dataset handling
- Single element dataset
- Even/odd number of elements (median)
- Multiple modes
- Large datasets (performance)
- Edge cases (all same values, extreme values)

---

## 2. MongoDB Tools (4-5 tools)

### Overview
Tools for interacting with MongoDB databases, enabling agents to perform CRUD operations on NoSQL data.

### Tool Category
- **Category**: `CategoryDatabase` (new)
- **Priority**: HIGH (Phase 2 - v0.3.0)
- **Dependencies**: `go.mongodb.org/mongo-driver/mongo` (official MongoDB Go driver)
- **Safety**: MEDIUM (write operations possible, requires authentication)

---

### SDK Research

**Official MongoDB Go Driver**:
```bash
go get go.mongodb.org/mongo-driver/mongo
# Latest: v1.14.0 (stable)
```

**Key Imports**:
```go
import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)
```

**Connection Pattern**:
```go
clientOptions := options.Client().
    ApplyURI("mongodb://localhost:27017").
    SetConnectTimeout(10 * time.Second)

client, err := mongo.Connect(context.Background(), clientOptions)
defer client.Disconnect(context.Background())

// Ping to verify connection
err = client.Ping(context.Background(), nil)
```

---

### Tool 2.1: mongodb_connect

**Purpose**: Establish connection to MongoDB instance/cluster

#### Specifications

**Name**: `mongodb_connect`

**Description**: 
> Connect to a MongoDB instance or cluster. Returns connection ID for subsequent operations.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_string": {
            "type": "string",
            "description": "MongoDB connection string (mongodb://... or mongodb+srv://...)"
        },
        "database": {
            "type": "string",
            "description": "Default database name (optional)"
        },
        "timeout": {
            "type": "integer",
            "description": "Connection timeout in seconds (default: 10)",
            "minimum": 1,
            "maximum": 60
        }
    },
    "required": ["connection_string"]
}
```

**Returns**:
```json
{
    "type": "mongodb_connection",
    "connection_id": "conn_abc123",
    "server_info": {
        "version": "6.0.5",
        "databases": 3
    },
    "connected": true
}
```

#### Implementation Details

**File**: `pkg/tools/database/mongodb/connect.go` (~120 lines)

**Connection Management**:
```go
// Global connection pool (thread-safe)
var connections = sync.Map{}

type MongoConnection struct {
    Client     *mongo.Client
    Database   *mongo.Database
    ConnectedAt time.Time
}

// Store connection for reuse
connectionID := generateConnectionID()
connections.Store(connectionID, &MongoConnection{
    Client: client,
    Database: database,
})
```

**Security**:
- Support TLS/SSL connections
- Connection string validation
- No plain-text password logging
- Connection pooling (max 10 connections)

---

### Tool 2.2: mongodb_find

**Purpose**: Query documents from MongoDB collection

#### Specifications

**Name**: `mongodb_find`

**Description**: 
> Find documents in a MongoDB collection using query filters.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {
            "type": "string",
            "description": "Connection ID from mongodb_connect"
        },
        "database": {
            "type": "string",
            "description": "Database name (optional if set in connection)"
        },
        "collection": {
            "type": "string",
            "description": "Collection name"
        },
        "filter": {
            "type": "object",
            "description": "MongoDB query filter (BSON format)"
        },
        "limit": {
            "type": "integer",
            "description": "Maximum documents to return (default: 10, max: 1000)",
            "minimum": 1,
            "maximum": 1000
        },
        "sort": {
            "type": "object",
            "description": "Sort order (e.g., {\"created_at\": -1})"
        },
        "projection": {
            "type": "object",
            "description": "Fields to include/exclude"
        }
    },
    "required": ["connection_id", "collection"]
}
```

**Returns**:
```json
{
    "type": "mongodb_find_result",
    "collection": "users",
    "count": 3,
    "documents": [
        {"_id": "...", "name": "John", "age": 30},
        {"_id": "...", "name": "Jane", "age": 25}
    ],
    "has_more": false
}
```

#### Implementation Details

**File**: `pkg/tools/database/mongodb/find.go` (~150 lines)

**Query Execution**:
```go
collection := getCollection(connID, dbName, collectionName)

findOptions := options.Find()
findOptions.SetLimit(int64(limit))
if sort != nil {
    findOptions.SetSort(sort)
}
if projection != nil {
    findOptions.SetProjection(projection)
}

cursor, err := collection.Find(ctx, filter, findOptions)
defer cursor.Close(ctx)

var results []bson.M
cursor.All(ctx, &results)
```

---

### Tool 2.3: mongodb_insert

**Purpose**: Insert documents into MongoDB collection

#### Specifications

**Name**: `mongodb_insert`

**Description**: 
> Insert one or multiple documents into a MongoDB collection.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {"type": "string"},
        "database": {"type": "string"},
        "collection": {"type": "string"},
        "documents": {
            "type": "array",
            "items": {"type": "object"},
            "description": "Documents to insert",
            "minItems": 1,
            "maxItems": 100
        }
    },
    "required": ["connection_id", "collection", "documents"]
}
```

**Returns**:
```json
{
    "type": "mongodb_insert_result",
    "collection": "users",
    "inserted_count": 2,
    "inserted_ids": ["507f1f77bcf86cd799439011", "507f191e810c19729de860ea"]
}
```

#### Implementation Details

**File**: `pkg/tools/database/mongodb/insert.go` (~100 lines)

```go
if len(documents) == 1 {
    result, err := collection.InsertOne(ctx, documents[0])
} else {
    result, err := collection.InsertMany(ctx, documents)
}
```

**Safety**: IsSafe() = false (modifies data)

---

### Tool 2.4: mongodb_update

**Purpose**: Update documents in MongoDB collection

#### Specifications

**Name**: `mongodb_update`

**Description**: 
> Update documents matching a filter in MongoDB collection.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {"type": "string"},
        "database": {"type": "string"},
        "collection": {"type": "string"},
        "filter": {
            "type": "object",
            "description": "Filter to match documents"
        },
        "update": {
            "type": "object",
            "description": "Update operations (e.g., {\"$set\": {...}})"
        },
        "update_many": {
            "type": "boolean",
            "description": "Update all matching documents (default: false)"
        }
    },
    "required": ["connection_id", "collection", "filter", "update"]
}
```

**Returns**:
```json
{
    "type": "mongodb_update_result",
    "matched_count": 1,
    "modified_count": 1,
    "upserted_count": 0
}
```

#### Implementation Details

**File**: `pkg/tools/database/mongodb/update.go` (~120 lines)

```go
if updateMany {
    result, err := collection.UpdateMany(ctx, filter, update)
} else {
    result, err := collection.UpdateOne(ctx, filter, update)
}
```

**Safety**: IsSafe() = false (modifies data)

---

### Tool 2.5: mongodb_delete

**Purpose**: Delete documents from MongoDB collection

#### Specifications

**Name**: `mongodb_delete`

**Description**: 
> Delete documents matching a filter from MongoDB collection. **WARNING: Irreversible operation.**

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {"type": "string"},
        "database": {"type": "string"},
        "collection": {"type": "string"},
        "filter": {
            "type": "object",
            "description": "Filter to match documents to delete"
        },
        "delete_many": {
            "type": "boolean",
            "description": "Delete all matching documents (default: false)"
        }
    },
    "required": ["connection_id", "collection", "filter"]
}
```

**Returns**:
```json
{
    "type": "mongodb_delete_result",
    "deleted_count": 1
}
```

#### Implementation Details

**File**: `pkg/tools/database/mongodb/delete.go` (~100 lines)

**Safety**: 
- IsSafe() = false (dangerous operation)
- Require confirmation for delete_many
- Log all deletions

---

## 3. Qdrant Tools (4-5 tools)

### Overview
Tools for interacting with Qdrant vector database, enabling agents to perform semantic search and vector operations.

### Tool Category
- **Category**: `CategoryVectorDB` (new)
- **Priority**: HIGH (Phase 2 - v0.3.0)
- **Dependencies**: `github.com/qdrant/go-client` (official Qdrant Go client)
- **Safety**: MEDIUM (write operations possible)

---

### SDK Research

**Official Qdrant Go Client**:
```bash
go get github.com/qdrant/go-client
# Latest: v1.15.2 (Oct 2024)
```

**Key Imports**:
```go
import (
    "context"
    "github.com/qdrant/go-client/qdrant"
)
```

**Connection Pattern**:
```go
// Local Qdrant
client, err := qdrant.NewClient(&qdrant.Config{
    Host: "localhost",
    Port: 6334,
})

// Qdrant Cloud
client, err := qdrant.NewClient(&qdrant.Config{
    Host:   "xyz-example.eu-central.aws.cloud.qdrant.io",
    Port:   6334,
    APIKey: "<api-key>",
    UseTLS: true,
    Cloud:  true,
})
```

**Basic Operations**:
```go
// Create collection
client.CreateCollection(ctx, &qdrant.CreateCollection{
    CollectionName: "my_collection",
    VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
        Size:     384,  // embedding dimension
        Distance: qdrant.Distance_Cosine,
    }),
})

// Insert points
client.Upsert(ctx, &qdrant.UpsertPoints{
    CollectionName: "my_collection",
    Points: []*qdrant.PointStruct{
        {
            Id:      qdrant.NewIDNum(1),
            Vectors: qdrant.NewVectors(0.05, 0.61, ...),
            Payload: qdrant.NewValueMap(map[string]any{"text": "Hello"}),
        },
    },
})

// Search
searchResult, err := client.Query(ctx, &qdrant.QueryPoints{
    CollectionName: "my_collection",
    Query:          qdrant.NewQuery(0.2, 0.1, ...),
    Limit:          10,
})
```

---

### Tool 3.1: qdrant_connect

**Purpose**: Connect to Qdrant vector database instance

#### Specifications

**Name**: `qdrant_connect`

**Description**: 
> Connect to a Qdrant vector database instance (local or cloud).

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "host": {
            "type": "string",
            "description": "Qdrant host (default: localhost)"
        },
        "port": {
            "type": "integer",
            "description": "Qdrant port (default: 6334)"
        },
        "api_key": {
            "type": "string",
            "description": "API key for authentication (optional)"
        },
        "use_tls": {
            "type": "boolean",
            "description": "Use TLS connection (default: false)"
        },
        "cloud": {
            "type": "boolean",
            "description": "Connect to Qdrant Cloud (default: false)"
        }
    }
}
```

**Returns**:
```json
{
    "type": "qdrant_connection",
    "connection_id": "qdrant_xyz123",
    "host": "localhost:6334",
    "collections_count": 3,
    "connected": true
}
```

#### Implementation Details

**File**: `pkg/tools/vectordb/qdrant/connect.go` (~130 lines)

**Connection Management**:
```go
var qdrantConnections = sync.Map{}

type QdrantConnection struct {
    Client      *qdrant.Client
    ConnectedAt time.Time
    Config      *qdrant.Config
}
```

---

### Tool 3.2: qdrant_create_collection

**Purpose**: Create a new vector collection in Qdrant

#### Specifications

**Name**: `qdrant_create_collection`

**Description**: 
> Create a new collection for storing vectors with specified parameters.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {"type": "string"},
        "collection_name": {"type": "string"},
        "vector_size": {
            "type": "integer",
            "description": "Dimension of vectors",
            "minimum": 1
        },
        "distance": {
            "type": "string",
            "enum": ["Cosine", "Euclid", "Dot"],
            "description": "Distance metric (default: Cosine)"
        },
        "on_disk": {
            "type": "boolean",
            "description": "Store vectors on disk (default: false)"
        }
    },
    "required": ["connection_id", "collection_name", "vector_size"]
}
```

**Returns**:
```json
{
    "type": "qdrant_collection_created",
    "collection_name": "my_embeddings",
    "vector_size": 384,
    "distance": "Cosine",
    "created": true
}
```

#### Implementation Details

**File**: `pkg/tools/vectordb/qdrant/create_collection.go` (~110 lines)

```go
distance := qdrant.Distance_Cosine // default
if params["distance"] == "Euclid" {
    distance = qdrant.Distance_Euclid
} else if params["distance"] == "Dot" {
    distance = qdrant.Distance_Dot
}

err := client.CreateCollection(ctx, &qdrant.CreateCollection{
    CollectionName: collectionName,
    VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
        Size:     uint64(vectorSize),
        Distance: distance,
        OnDisk:   onDisk,
    }),
})
```

---

### Tool 3.3: qdrant_upsert

**Purpose**: Insert or update vectors in Qdrant collection

#### Specifications

**Name**: `qdrant_upsert`

**Description**: 
> Insert or update vector points in a Qdrant collection. Upsert operation replaces existing points with same ID.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {"type": "string"},
        "collection_name": {"type": "string"},
        "points": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "id": {
                        "type": ["integer", "string"],
                        "description": "Point ID"
                    },
                    "vector": {
                        "type": "array",
                        "items": {"type": "number"},
                        "description": "Vector embedding"
                    },
                    "payload": {
                        "type": "object",
                        "description": "Metadata (optional)"
                    }
                },
                "required": ["id", "vector"]
            },
            "minItems": 1,
            "maxItems": 100
        }
    },
    "required": ["connection_id", "collection_name", "points"]
}
```

**Returns**:
```json
{
    "type": "qdrant_upsert_result",
    "collection_name": "my_embeddings",
    "upserted_count": 5,
    "operation_id": 12345,
    "status": "completed"
}
```

#### Implementation Details

**File**: `pkg/tools/vectordb/qdrant/upsert.go` (~140 lines)

```go
// Convert points to Qdrant format
qdrantPoints := make([]*qdrant.PointStruct, len(points))
for i, p := range points {
    id := qdrant.NewIDNum(p.ID) // or NewIDString for UUID
    vectors := qdrant.NewVectors(p.Vector...)
    payload := qdrant.NewValueMap(p.Payload)
    
    qdrantPoints[i] = &qdrant.PointStruct{
        Id:      id,
        Vectors: vectors,
        Payload: payload,
    }
}

operationInfo, err := client.Upsert(ctx, &qdrant.UpsertPoints{
    CollectionName: collectionName,
    Points:         qdrantPoints,
})
```

**Safety**: IsSafe() = false (modifies data)

---

### Tool 3.4: qdrant_search

**Purpose**: Perform vector similarity search in Qdrant

#### Specifications

**Name**: `qdrant_search`

**Description**: 
> Search for similar vectors in a Qdrant collection using vector similarity.

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {"type": "string"},
        "collection_name": {"type": "string"},
        "query_vector": {
            "type": "array",
            "items": {"type": "number"},
            "description": "Query vector for similarity search"
        },
        "limit": {
            "type": "integer",
            "description": "Maximum results to return (default: 10)",
            "minimum": 1,
            "maximum": 100
        },
        "filter": {
            "type": "object",
            "description": "Payload filter conditions (optional)"
        },
        "with_payload": {
            "type": "boolean",
            "description": "Include payload in results (default: true)"
        },
        "with_vectors": {
            "type": "boolean",
            "description": "Include vectors in results (default: false)"
        },
        "score_threshold": {
            "type": "number",
            "description": "Minimum similarity score (optional)",
            "minimum": 0,
            "maximum": 1
        }
    },
    "required": ["connection_id", "collection_name", "query_vector"]
}
```

**Returns**:
```json
{
    "type": "qdrant_search_result",
    "collection_name": "my_embeddings",
    "query_time_ms": 12,
    "count": 5,
    "results": [
        {
            "id": 42,
            "score": 0.95,
            "payload": {"text": "Hello world", "category": "greeting"},
            "vector": [0.1, 0.2, ...] // if with_vectors=true
        }
    ]
}
```

#### Implementation Details

**File**: `pkg/tools/vectordb/qdrant/search.go` (~160 lines)

```go
queryPoints := &qdrant.QueryPoints{
    CollectionName: collectionName,
    Query:          qdrant.NewQuery(queryVector...),
    Limit:          uint64(limit),
    WithPayload:    qdrant.NewWithPayload(withPayload),
}

// Add filter if provided
if filter != nil {
    queryPoints.Filter = &qdrant.Filter{
        Must: convertToConditions(filter),
    }
}

// Add score threshold
if scoreThreshold > 0 {
    queryPoints.ScoreThreshold = &scoreThreshold
}

searchResult, err := client.Query(ctx, queryPoints)
```

**Performance Considerations**:
- Default limit: 10 (prevent large result sets)
- Query timeout: 30 seconds
- Cache query vectors (optional)

---

### Tool 3.5: qdrant_delete

**Purpose**: Delete points from Qdrant collection

#### Specifications

**Name**: `qdrant_delete`

**Description**: 
> Delete points from a Qdrant collection by ID or filter. **WARNING: Irreversible operation.**

**Parameters**:
```go
{
    "type": "object",
    "properties": {
        "connection_id": {"type": "string"},
        "collection_name": {"type": "string"},
        "point_ids": {
            "type": "array",
            "items": {"type": ["integer", "string"]},
            "description": "Point IDs to delete (optional if filter provided)"
        },
        "filter": {
            "type": "object",
            "description": "Delete points matching filter (optional)"
        }
    },
    "required": ["connection_id", "collection_name"]
}
```

**Returns**:
```json
{
    "type": "qdrant_delete_result",
    "collection_name": "my_embeddings",
    "deleted_count": 3,
    "operation_id": 12346,
    "status": "completed"
}
```

#### Implementation Details

**File**: `pkg/tools/vectordb/qdrant/delete.go` (~120 lines)

```go
if len(pointIDs) > 0 {
    // Delete by IDs
    ids := make([]*qdrant.PointId, len(pointIDs))
    for i, id := range pointIDs {
        ids[i] = qdrant.NewIDNum(id)
    }
    
    operationInfo, err := client.Delete(ctx, &qdrant.DeletePoints{
        CollectionName: collectionName,
        PointsSelector: &qdrant.PointsSelector{
            Points: &qdrant.PointsIdsList{
                Ids: ids,
            },
        },
    })
} else if filter != nil {
    // Delete by filter
    operationInfo, err := client.Delete(ctx, &qdrant.DeletePoints{
        CollectionName: collectionName,
        PointsSelector: &qdrant.PointsSelector{
            Filter: convertToFilter(filter),
        },
    })
}
```

**Safety**: 
- IsSafe() = false (dangerous operation)
- Require one of: point_ids or filter
- Log all deletions

---

## Implementation Plan

### Phase 2: Database Tools (v0.3.0)
**Target**: November-December 2025  
**Priority**: HIGH

1. **MongoDB Tools (4-5 tools)** - Week 1-2
   - mongodb_connect
   - mongodb_find
   - mongodb_insert
   - mongodb_update
   - mongodb_delete
   - Test suite (~400 lines)

2. **Qdrant Tools (4-5 tools)** - Week 3-4
   - qdrant_connect
   - qdrant_create_collection
   - qdrant_upsert
   - qdrant_search
   - qdrant_delete
   - Test suite (~400 lines)

### Phase 3: Math & Utilities (v0.3.0)
**Target**: December 2025  
**Priority**: MEDIUM

1. **Math Tools (2 tools)** - Week 1
   - math_calculate
   - math_stats
   - Test suite (~190 lines)

---

## Dependency Management

### New Dependencies
```go
// go.mod additions
require (
    go.mongodb.org/mongo-driver v1.14.0
    github.com/qdrant/go-client v1.15.2
)
```

### Total Package Size Impact
- MongoDB driver: ~5MB
- Qdrant client: ~2MB
- Math tools: Standard library only (0MB)
- **Total**: ~7MB additional

---

## Testing Strategy

### Unit Tests
- **Math Tools**: ~190 lines, 20 test cases
- **MongoDB Tools**: ~400 lines, 30 test cases
- **Qdrant Tools**: ~400 lines, 30 test cases
- **Total**: ~990 lines, 80 test cases

### Integration Tests
- MongoDB: Requires MongoDB instance (Docker container)
- Qdrant: Requires Qdrant instance (Docker container)
- Math: No external dependencies

### Test Infrastructure
```bash
# Docker compose for test services
docker-compose.test.yml:
  mongodb:
    image: mongo:6.0
    ports: [27017:27017]
  
  qdrant:
    image: qdrant/qdrant:v1.7.4
    ports: [6334:6334]
```

---

## Security Considerations

### MongoDB Tools
- ✅ Connection string validation (prevent injection)
- ✅ Authentication required
- ✅ TLS support
- ✅ Query size limits (max 1000 docs)
- ✅ Batch operation limits (max 100 inserts)
- ⚠️ No direct shell command execution
- ⚠️ Log all write operations

### Qdrant Tools
- ✅ API key authentication
- ✅ TLS support
- ✅ Query result limits
- ✅ Batch operation limits
- ⚠️ Log all write operations

### Math Tools
- ✅ Expression validation (whitelist approach)
- ✅ No code execution
- ✅ Input size limits
- ✅ No file system access
- ✅ Deterministic results

---

## Documentation Requirements

### For Each Tool
1. Detailed parameter descriptions
2. Return value schemas
3. Usage examples (3-5 per tool)
4. Security notes
5. Error handling guide
6. Performance considerations

### Examples
- Complete examples for each category
- Integration with LLM agents
- Best practices guide
- Troubleshooting guide

---

## Success Metrics

### v0.3.0 Goals
- ✅ 10-12 new tools implemented
- ✅ 80 new test cases passing
- ✅ Test coverage >= 80% for new tools
- ✅ Documentation for all new tools
- ✅ 3 new example applications
- ✅ Total built-in tools: 23-25 tools

### Performance Targets
- Math calculations: < 1ms
- MongoDB operations: < 100ms (local), < 500ms (remote)
- Qdrant operations: < 50ms (local), < 200ms (remote)

---

## Alternative Approaches Considered

### Math Tools
- **Alternative**: Use expression evaluation library (e.g., govaluate)
- **Decision**: Implement custom parser for better security control

### MongoDB
- **Alternative**: Generic database interface (support multiple databases)
- **Decision**: MongoDB-specific tools first, generic interface later

### Qdrant
- **Alternative**: Generic vector database interface
- **Decision**: Qdrant-specific tools first, add others (Pinecone, Weaviate) in v0.4.0

---

## Next Steps

1. **Review & Approve Design** (Current)
2. **Create Implementation Tickets** (1 day)
3. **Phase 2 Implementation** (4 weeks)
   - Week 1-2: MongoDB tools
   - Week 3-4: Qdrant tools
4. **Phase 3 Implementation** (1 week)
   - Math tools
5. **Testing & Documentation** (1 week)
6. **v0.3.0 Release** (Target: End of December 2025)

---

**Document Version**: 1.0  
**Last Updated**: October 27, 2025  
**Status**: Ready for Review
