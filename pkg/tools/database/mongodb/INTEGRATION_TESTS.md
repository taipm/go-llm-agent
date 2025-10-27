# MongoDB Integration Tests

## Overview

Integration tests for MongoDB tools that test against **real MongoDB instances** instead of mocks. These tests verify actual CRUD operations, connection pooling, and error handling.

## Prerequisites

### Option 1: Local MongoDB (Fastest)
Install and run MongoDB locally:

```bash
# macOS with Homebrew
brew install mongodb-community
brew services start mongodb-community

# Verify MongoDB is running
mongosh --eval "db.version()"
```

### Option 2: Custom MongoDB URL
Set environment variable if MongoDB is running on a different host/port:

```bash
export MONGODB_URL="mongodb://username:password@hostname:27017"
```

Default: `mongodb://localhost:27017`

## Running Tests

### Run all integration tests
```bash
go test -v -tags=integration ./pkg/tools/database/mongodb/
```

### Run specific test
```bash
go test -v -tags=integration -run TestIntegrationMongoDBCRUDWorkflow ./pkg/tools/database/mongodb/
```

### Run with coverage
```bash
go test -v -tags=integration -cover ./pkg/tools/database/mongodb/
```

## Test Suites

### 1. TestIntegrationMongoDBConnect
Tests MongoDB connection establishment:
- Connects to real MongoDB instance
- Verifies connection_id generation
- Retrieves and validates server info (version, git version)
- Tests connection cleanup

**Duration:** ~30ms

### 2. TestIntegrationMongoDBCRUDWorkflow
Complete CRUD workflow test:
1. **Connect** - Establishes connection to test database
2. **Insert** - Batch insert 3 documents (Alice, Bob, Charlie)
3. **Find All** - Query all documents, verify count = 3
4. **Find Filtered** - Query with filter `{age: {$gte: 30}}`, verify sorting
5. **Update** - Update Alice's age from 30 to 31, add status field
6. **Verify Update** - Confirm changes persisted
7. **Delete One** - Delete Bob's document
8. **Verify Delete** - Confirm count = 2
9. **Cleanup** - Remove all test documents

**Duration:** ~40ms  
**Database:** `go_llm_agent_test`  
**Collections:** Temporary with timestamp suffix (auto-cleanup)

### 3. TestIntegrationMongoDBConnectionPooling
Tests connection pool management:
- Creates 5 concurrent connections
- Verifies each gets unique connection_id
- Tests connection cleanup
- Validates no resource leaks

**Duration:** ~20ms

### 4. TestIntegrationMongoDBErrorHandling
Tests error scenarios:
1. **Invalid connection string** - Tests timeout and error handling
2. **Empty delete filter** - Verifies safety check (prevents accidental delete all)
3. **Connection not found** - Tests error for non-existent connection_id

**Duration:** ~5s (includes 2s timeout for invalid connection)

### 5. TestIntegrationMongoDBBatchInsert
Tests batch insert operations:
- Inserts 50 documents in one operation
- Verifies batch size limits (max 100)
- Tests cleanup of batch data

**Duration:** ~30ms

## Test Results

```
PASS: TestIntegrationMongoDBConnect (0.01s)
PASS: TestIntegrationMongoDBCRUDWorkflow (0.04s)
PASS: TestIntegrationMongoDBConnectionPooling (0.02s)
PASS: TestIntegrationMongoDBErrorHandling (5.00s)
PASS: TestIntegrationMongoDBBatchInsert (0.03s)
PASS: TestConnectTool (0.00s)
PASS: TestFindTool (0.00s)
PASS: TestInsertTool (0.00s)
PASS: TestUpdateTool (0.00s)
PASS: TestDeleteTool (0.00s)
PASS: TestGetConnectionNotFound (0.00s)
PASS: TestDeleteEmptyFilter (0.00s)

Total: 12 tests
Duration: 5.6s
Coverage: Full CRUD + connection pooling + error handling
```

## Test Data Management

### Databases
- **Test database:** `go_llm_agent_test`
- Automatically created if not exists
- Can be safely deleted after tests

### Collections
- Temporary collections with timestamp suffix: `test_users_1761535991`
- Each test run creates unique collections
- All test data is cleaned up in `defer` statements
- No persistent data left after tests

### Cleanup
```bash
# Remove test database (optional)
mongosh go_llm_agent_test --eval "db.dropDatabase()"
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mongodb:
        image: mongo:8.0
        ports:
          - 27017:27017
        options: >-
          --health-cmd "mongosh --eval 'db.runCommand({ping: 1})'"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Run integration tests
        run: go test -v -tags=integration ./pkg/tools/database/mongodb/
```

## Troubleshooting

### MongoDB not running
```
Error: failed to ping MongoDB: server selection error
```
**Solution:** Start MongoDB service
```bash
brew services start mongodb-community
```

### Connection timeout
```
Error: context deadline exceeded
```
**Solution:** Check MongoDB is accessible and increase timeout in test

### Permission denied
```
Error: not authorized on go_llm_agent_test
```
**Solution:** Configure MongoDB authentication or use unauthenticated local instance

## Comparison: Unit vs Integration Tests

| Aspect | Unit Tests | Integration Tests |
|--------|-----------|------------------|
| **Speed** | < 1ms | ~5s total |
| **Dependencies** | None | MongoDB required |
| **Coverage** | Tool structure | Real operations |
| **Use case** | Development | Pre-deployment |
| **Build tag** | (default) | `-tags=integration` |

## Best Practices

1. **Run unit tests frequently** during development
2. **Run integration tests** before commits
3. **Run both in CI/CD** pipeline
4. **Use separate test database** - never test on production data
5. **Clean up test data** in defer statements
6. **Use unique collection names** with timestamps

## MongoDB Version Compatibility

Tested with:
- MongoDB 8.0.15 (latest)
- MongoDB 7.x (compatible)
- MongoDB 6.x (compatible)

Driver: `go.mongodb.org/mongo-driver v1.17.4`

## Additional Resources

- [MongoDB Go Driver Documentation](https://www.mongodb.com/docs/drivers/go/current/)
- [MongoDB Testing Best Practices](https://www.mongodb.com/docs/manual/tutorial/test-mongodb/)
- [Go Build Tags](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
