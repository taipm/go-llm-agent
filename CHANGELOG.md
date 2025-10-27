# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Gmail Tools Package** (4 email automation tools with OAuth2)
  - `gmail_send` - Send emails via Gmail API
    * Support for to, cc, bcc recipients
    * HTML and plain text email bodies
    * Returns message_id and thread_id
  - `gmail_read` - Read email messages by ID
    * Full message content with headers and body
    * Attachment metadata extraction
    * Three format options: full, metadata, minimal
    * Recursive multipart message parsing
  - `gmail_list` - List emails with filters and pagination
    * Gmail search query support
    * Label filtering (INBOX, UNREAD, etc.)
    * Configurable max results (up to 500)
    * Pagination with next_page_token
  - `gmail_search` - Advanced email search
    * Full Gmail search syntax (from:, to:, subject:, is:unread, etc.)
    * Optional metadata extraction (from, to, subject, date)
    * Up to 100 results per search
  - **OAuth2 Authentication Infrastructure**
    * AuthHelper for Gmail API service management
    * Token caching (credentials.json, token.json)
    * Interactive authorization flow for first-time setup
    * Automatic token refresh
  - **NOT loaded by default** (requires OAuth2 credentials setup)
    * Set `NoGmail: false` in builtin.Config to enable
    * Use `GetGmailTools()` for manual access
  - Official Google API library: google.golang.org/api v0.253.0
  - OAuth2 authentication: golang.org/x/oauth2 v0.32.0
  - Comprehensive README with OAuth2 setup guide and examples
  - CategoryEmail added to tool categories
  - Security best practices documented

- **Network Tools Package** (5 professional network diagnostic tools)
  - `network_dns_lookup` - DNS record queries using `miekg/dns` library
    * Support for A, AAAA, MX, TXT, NS, CNAME, SOA, PTR records
    * Custom DNS servers (Google DNS, Cloudflare, OpenDNS by default)
    * TCP/UDP support with TTL information
    * Reverse DNS (PTR) lookups
  - `network_ping` - ICMP ping and TCP connectivity checks using `go-ping/ping`
    * ICMP ping with packet loss and RTT statistics
    * TCP port availability testing
    * Connection latency measurement
  - `network_whois_lookup` - WHOIS queries using `likexian/whois-parser`
    * Domain registration information
    * Registrar, registrant, admin, tech contacts
    * Nameservers and domain status
  - `network_ssl_cert_check` - SSL/TLS certificate validation using `crypto/tls`
    * Certificate chain inspection
    * Expiration checking with warnings
    * Subject Alternative Names (SANs)
    * TLS version and cipher suite detection
  - `network_ip_info` - IP geolocation using `oschwald/geoip2-golang`
    * IP version and privacy status
    * Reverse DNS lookups
    * Geolocation (country, city, coordinates) with GeoIP2 database
    * ISP and ASN information
  - All network tools loaded automatically by default
  - Professional libraries used: miekg/dns, go-ping/ping, likexian/whois-parser, oschwald/geoip2-golang
  - Comprehensive README with examples and troubleshooting
  - Updated ToolCount from 20 to 24 (25 with GeoIP database)

### Changed

- Updated builtin tools configuration to include GmailConfig and NoGmail flag
- Added CategoryEmail to tool categories (now 8 categories)
- Updated builtin.GetRegistry() to optionally load Gmail tools
- Added GetGmailTools() helper function for manual Gmail tool access
- Updated builtin tools configuration to include NetworkConfig
- Added CategoryNetwork to tool categories
- Enhanced builtin.GetRegistry() to auto-load network tools

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
