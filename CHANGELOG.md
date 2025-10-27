# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2025-01-27

### Added

- **Professional Logging System** - Comprehensive logging inspired by AutoGPT/CrewAI
  - Logger interface for extensibility
  - ConsoleLogger with ANSI colors and timestamps
  - NoopLogger for disabling logs
  - Multiple log levels: Debug, Info, Warn, Error
  - Visual indicators: 👤 (user), 🤔 (thinking), 🔧 (tool), ✓ (success), 💬 (response), 💾 (memory)
  - Helper functions: LogUserMessage, LogThinking, LogToolCall, LogToolResult, LogResponse, LogMemory, LogIteration
  - Configuration options: WithLogger(), WithLogLevel(), DisableLogging()
  - Integrated throughout agent lifecycle

- **MongoDB Tools** (6 tools) - Complete MongoDB database operations
  - `mongodb_connect` - Connect to MongoDB with connection string
  - `mongodb_insert` - Insert documents into collection
  - `mongodb_find` - Find documents with query
  - `mongodb_update` - Update documents with filter
  - `mongodb_delete` - Delete documents from collection
  - `mongodb_aggregate` - Run aggregation pipelines
  - Supports connection pooling and error handling
  - Integration tests with real MongoDB

- **Math Tools** (2 tools) - Mathematical operations and statistics
  - `math_calculate` - Evaluate mathematical expressions using govaluate
  - `math_stats` - Calculate statistics (mean, median, mode, std dev) using gonum

- **Integration Testing** - MongoDB integration tests
  - 5 test suites: Connect, CRUD workflow, Connection pooling, Error handling, Batch insert
  - Testing with real MongoDB instance (not containers)
  - Documentation: `pkg/tools/database/mongodb/INTEGRATION_TESTS.md`
  - Build tag: `//go:build integration`

- **Examples**
  - `examples/agent_with_builtin_tools` - Comprehensive demo of all 20 built-in tools
  - `examples/agent_with_logging` - Vietnamese conversation demo showing logging features

### Fixed

- **Critical Memory Bug** - Agent now correctly persists conversation history
  - Issue: runLoop() was not saving messages to memory
  - Fixed: Added memory.Add() calls for assistant messages, tool results, and final responses
  - Validation: TestChatWithMemory passing, multi-turn conversations working
  - Impact: Multi-provider support (Gemini, Ollama) now maintains context correctly

### Changed

- Agent struct now includes `logger Logger` field (defaults to ConsoleLogger at INFO level)
- Agent.New() initializes default logger automatically
- Logging is integrated by default but can be disabled or customized
- Updated `.gitignore` to exclude example binaries

### Testing

- All 200+ unit tests passing
- 5 MongoDB integration tests passing
- Validated with:
  - Gemini API (gemini-2.5-flash)
  - Ollama (qwen3:1.7b)
  - Vietnamese language scenarios
  - Multi-turn conversations with memory
  - Tool calling with multiple tools

### Dependencies

- Added: `go.mongodb.org/mongo-driver v1.17.4`
- Added: `github.com/Knetic/govaluate v3.0.0`
- Added: `gonum.org/v1/gonum v0.16.0`

## [0.1.0] - 2025-01-20

### Added

- Initial release
- Agent framework with LLM provider abstraction
- Tool system with 11 built-in tools across 4 categories:
  - DateTime: datetime_now, datetime_calc, datetime_format
  - File: file_read, file_write, file_list
  - Web: web_fetch, web_search
  - System: system_exec, system_env, system_info
- Multi-provider support:
  - Ollama (local)
  - OpenAI (cloud)
  - Gemini (cloud)
- Memory system (BufferMemory with 50 message capacity)
- Function calling support
- Comprehensive unit tests
- Examples: basic agent, simple conversation

### Documentation

- README.md with installation and usage
- Examples for all major features
- API documentation in code

---

## Release Notes

### v0.1.1 - The Logging & MongoDB Release

This release adds professional-grade logging to help users understand what their agents are doing, plus MongoDB database tools and critical memory fixes.

**Key Highlights:**

1. **See What Your Agent Is Doing** - New logging system with colored output and visual indicators makes it easy to track agent behavior in real-time

2. **MongoDB Support** - 6 new tools for complete MongoDB operations (connect, insert, find, update, delete, aggregate)

3. **Memory Fixed** - Critical bug fix ensures agents remember conversation context across multiple turns

4. **Math Tools** - Calculate expressions and statistics programmatically

5. **Production Ready** - All 200+ tests passing, validated with Gemini and Ollama, multi-language support

**Upgrade Notes:**

- Logging is enabled by default at INFO level
- To disable: `agent.New(llm, agent.DisableLogging())`
- To adjust: `agent.New(llm, agent.WithLogLevel(agent.LogLevelDebug))`
- No breaking changes - fully backward compatible

**What's Next (v0.1.2):**

- Qdrant vector search tools
- Data processing tools (JSON, CSV, XML)
- Additional database connectors
- Performance optimizations
