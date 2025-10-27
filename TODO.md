# TODO List - Active Tasks

> **Note**: Completed tasks have been moved to [DONE.md](DONE.md)

## Current Status

**Project**: go-llm-agent  
**Version**: v0.2.0 (80% complete) → v0.3.0 (Built-in Tools)  
**Last Updated**: October 27, 2025

---

## 🎯 NEW PRIORITY: Complete Built-in Tools Phase 1 (10/10 tools)

**Current Status**: 3/10 tools complete (30%)  
**Target**: Complete remaining 7 tools before v0.2.0 release  
**Estimated Time**: 2-3 days  
**Status**: � IN PROGRESS

### Rationale
- Built-in tools are essential infrastructure for v0.3.0
- Better to release v0.2.0 with complete built-in tools system
- Testing v0.2.0 will benefit from having all tools available
- More impressive release with 10 built-in tools included

---

## 🔧 Phase 1: Complete Remaining Built-in Tools (7 tools)

### Day 1: File Tools Completion (2 tools) - 4-5 hours

#### Tool 1: file_write - Write content to files
**Priority**: HIGH  
**Estimated Time**: 2-2.5 hours  
**File**: `pkg/tools/file/write.go` (~180 lines)

**Tasks**:
- [ ] Design WriteToolConfig struct
  - [ ] AllowedPaths []string (security whitelist)
  - [ ] MaxFileSize int64 (default 10MB)
  - [ ] CreateDirs bool (create parent directories)
  - [ ] Backup bool (backup existing files)
  - [ ] AppendMode bool (append vs overwrite)
- [ ] Implement WriteTool struct
  - [ ] Name() returns "file_write"
  - [ ] Description() clear explanation
  - [ ] Parameters() JSON schema (path, content, append)
  - [ ] Category() returns ToolCategoryFile
  - [ ] IsSafe() returns false (can modify filesystem)
- [ ] Implement Execute() method
  - [ ] Validate path (prevent traversal)
  - [ ] Check path in AllowedPaths
  - [ ] Check content size <= MaxFileSize
  - [ ] Backup existing file if enabled
  - [ ] Create parent directories if needed
  - [ ] Write or append content
  - [ ] Return success with file info
- [ ] Add comprehensive error handling
  - [ ] Path validation errors
  - [ ] Permission denied errors
  - [ ] Disk space errors
  - [ ] Size limit errors
- [ ] Unit tests (~120 lines)
  - [ ] Test normal write
  - [ ] Test append mode
  - [ ] Test backup creation
  - [ ] Test directory creation
  - [ ] Test path validation
  - [ ] Test size limits
  - [ ] Test security restrictions
- [ ] Integration test with LLM
  - [ ] LLM can write files via tool calling
  - [ ] Proper error messages returned

#### Tool 2: file_delete - Delete files/directories
**Priority**: MEDIUM  
**Estimated Time**: 2-2.5 hours  
**File**: `pkg/tools/file/delete.go` (~150 lines)

**Tasks**:
- [ ] Design DeleteToolConfig struct
  - [ ] AllowedPaths []string
  - [ ] RecursiveDelete bool
  - [ ] RequireConfirmation bool
  - [ ] ProtectedPaths []string (cannot delete)
- [ ] Implement DeleteTool struct
  - [ ] Name() returns "file_delete"
  - [ ] Description() with safety warnings
  - [ ] Parameters() JSON schema (path, recursive)
  - [ ] Category() returns ToolCategoryFile
  - [ ] IsSafe() returns false (dangerous operation)
- [ ] Implement Execute() method
  - [ ] Validate path
  - [ ] Check not in ProtectedPaths (/, /etc, /usr, etc.)
  - [ ] Check path in AllowedPaths
  - [ ] Handle file vs directory
  - [ ] Recursive delete if directory and enabled
  - [ ] Return deletion summary
- [ ] Safety features
  - [ ] Block system directories
  - [ ] Block hidden files by default
  - [ ] Confirmation mechanism
- [ ] Unit tests (~100 lines)
  - [ ] Test file deletion
  - [ ] Test directory deletion
  - [ ] Test recursive deletion
  - [ ] Test protected paths
  - [ ] Test security restrictions
- [ ] Integration test with LLM

**Day 1 Deliverables**:
- ✅ 2 file tools implemented
- ✅ ~330 lines production code
- ✅ ~220 lines test code
- ✅ File tools: 4/4 complete (100%)

---

### Day 2: Web Tools Implementation (3 tools) - 6-7 hours

#### Tool 3: web_fetch - HTTP GET requests
**Priority**: HIGH  
**Estimated Time**: 2.5-3 hours  
**File**: `pkg/tools/web/fetch.go` (~200 lines)

**Tasks**:
- [ ] Design FetchToolConfig struct
  - [ ] Timeout time.Duration (default 30s)
  - [ ] MaxResponseSize int64 (default 1MB)
  - [ ] AllowedDomains []string (whitelist)
  - [ ] FollowRedirects bool
  - [ ] MaxRedirects int
- [ ] Implement FetchTool struct
  - [ ] Name() returns "web_fetch"
  - [ ] Description() explains HTTP GET
  - [ ] Parameters() JSON schema (url, headers)
  - [ ] Category() returns ToolCategoryWeb
  - [ ] IsSafe() returns true (read-only)
- [ ] Implement Execute() method
  - [ ] Validate URL (prevent SSRF)
  - [ ] Check domain whitelist
  - [ ] Create HTTP client with timeout
  - [ ] Set custom headers if provided
  - [ ] Execute GET request
  - [ ] Check response size
  - [ ] Read response body
  - [ ] Return status, headers, body
- [ ] Security features
  - [ ] Block private IP ranges (SSRF prevention)
  - [ ] Block localhost/127.0.0.1
  - [ ] URL validation
  - [ ] Response size limits
- [ ] Unit tests (~150 lines)
  - [ ] Test successful fetch
  - [ ] Test custom headers
  - [ ] Test timeout
  - [ ] Test size limits
  - [ ] Test SSRF prevention
  - [ ] Test domain whitelist
  - [ ] Mock HTTP server for tests
- [ ] Integration test with LLM

#### Tool 4: web_post - HTTP POST requests
**Priority**: MEDIUM  
**Estimated Time**: 2-2.5 hours  
**File**: `pkg/tools/web/post.go` (~170 lines)

**Tasks**:
- [ ] Reuse FetchToolConfig (similar config)
- [ ] Implement PostTool struct
  - [ ] Name() returns "web_post"
  - [ ] Description() explains HTTP POST
  - [ ] Parameters() JSON schema (url, headers, body, content_type)
  - [ ] Category() returns ToolCategoryWeb
  - [ ] IsSafe() returns false (can modify data)
- [ ] Implement Execute() method
  - [ ] Same security checks as fetch
  - [ ] Support JSON and form data
  - [ ] Set Content-Type header
  - [ ] Execute POST request
  - [ ] Return response
- [ ] Unit tests (~120 lines)
  - [ ] Test JSON POST
  - [ ] Test form data POST
  - [ ] Test custom headers
  - [ ] Security tests
- [ ] Integration test with LLM

#### Tool 5: web_scrape - Web scraping with CSS selectors
**Priority**: LOW  
**Estimated Time**: 2-2.5 hours  
**File**: `pkg/tools/web/scrape.go` (~220 lines)

**Tasks**:
- [ ] Add dependency: `github.com/PuerkitoBio/goquery`
- [ ] Design ScrapeToolConfig struct
  - [ ] Extends FetchToolConfig
  - [ ] RateLimitDelay time.Duration
  - [ ] RespectRobotsTxt bool
- [ ] Implement ScrapeTool struct
  - [ ] Name() returns "web_scrape"
  - [ ] Description() explains CSS selector extraction
  - [ ] Parameters() JSON schema (url, selectors)
  - [ ] Category() returns ToolCategoryWeb
  - [ ] IsSafe() returns true
- [ ] Implement Execute() method
  - [ ] Fetch HTML (reuse fetch logic)
  - [ ] Parse with goquery
  - [ ] Apply CSS selectors
  - [ ] Extract text/attributes
  - [ ] Return structured data
- [ ] Rate limiting
  - [ ] Delay between requests
  - [ ] Respect robots.txt (basic check)
- [ ] Unit tests (~130 lines)
  - [ ] Test CSS selectors
  - [ ] Test multiple selectors
  - [ ] Test attribute extraction
  - [ ] Mock HTML pages
- [ ] Integration test with LLM

**Day 2 Deliverables**:
- ✅ 3 web tools implemented
- ✅ ~590 lines production code
- ✅ ~400 lines test code
- ✅ Web tools: 3/3 complete (100%)

---

### Day 3: DateTime & Math Tools (2 tools) - 3-4 hours

#### Tool 6: datetime_format - Format conversion
**Priority**: MEDIUM  
**Estimated Time**: 1.5-2 hours  
**File**: `pkg/tools/datetime/format.go` (~120 lines)

**Tasks**:
- [ ] Design FormatToolConfig struct
  - [ ] DefaultTimezone string (default "UTC")
  - [ ] SupportedFormats []string
- [ ] Implement FormatTool struct
  - [ ] Name() returns "datetime_format"
  - [ ] Description() explains format conversion
  - [ ] Parameters() JSON schema (input, from_format, to_format, timezone)
  - [ ] Category() returns ToolCategoryDateTime
  - [ ] IsSafe() returns true
- [ ] Implement Execute() method
  - [ ] Parse input with from_format
  - [ ] Convert timezone if specified
  - [ ] Format with to_format
  - [ ] Support common formats (RFC3339, RFC822, etc.)
  - [ ] Return formatted string
- [ ] Unit tests (~80 lines)
  - [ ] Test format conversions
  - [ ] Test timezone conversions
  - [ ] Test error cases
- [ ] Integration test with LLM

#### Tool 7: datetime_calc - Date calculations
**Priority**: MEDIUM  
**Estimated Time**: 1.5-2 hours  
**File**: `pkg/tools/datetime/calc.go` (~140 lines)

**Tasks**:
- [ ] Design CalcToolConfig struct
  - [ ] DefaultTimezone string
- [ ] Implement CalcTool struct
  - [ ] Name() returns "datetime_calc"
  - [ ] Description() explains date math
  - [ ] Parameters() JSON schema (operation, date, amount, unit)
  - [ ] Category() returns ToolCategoryDateTime
  - [ ] IsSafe() returns true
- [ ] Implement Execute() method
  - [ ] Support operations: add, subtract, diff
  - [ ] Support units: years, months, days, hours, minutes, seconds
  - [ ] Parse dates
  - [ ] Perform calculations
  - [ ] Return result with explanation
- [ ] Unit tests (~90 lines)
  - [ ] Test add operations
  - [ ] Test subtract operations
  - [ ] Test diff calculations
  - [ ] Test edge cases
- [ ] Integration test with LLM

**Day 3 Deliverables**:
- ✅ 2 datetime tools implemented
- ✅ ~260 lines production code
- ✅ ~170 lines test code
- ✅ DateTime tools: 3/3 complete (100%)

---

### Summary: Phase 1 Completion

**Total Effort**: 2-3 days (~13-16 hours)

**Deliverables**:
- ✅ 7 new tools implemented
- ✅ ~1,180 lines production code
- ✅ ~790 lines test code
- ✅ All tools tested with LLM integration
- ✅ Phase 1: 10/10 tools complete (100%)

**Tools Completed**:
1. ✅ file_read (Day 4 - Oct 27)
2. ✅ file_list (Day 4 - Oct 27)
3. ✅ datetime_now (Day 4 - Oct 27)
4. 🔄 file_write (Day 1)
5. 🔄 file_delete (Day 1)
6. 🔄 web_fetch (Day 2)
7. 🔄 web_post (Day 2)
8. 🔄 web_scrape (Day 2)
9. 🔄 datetime_format (Day 3)
10. 🔄 datetime_calc (Day 3)

---

## 🚀 THEN: Sprint 3 Day 5 - v0.2.0 Release

**Target Date**: October 30-31, 2025  
**Estimated Time**: 4-5 hours  
**Status**: ⏸️ PENDING (after built-in tools)

### Tasks

#### 1. Code Review & Cleanup (2-3 hours)
- [ ] Review all provider implementations for consistency
  - [ ] Check OpenAI provider error messages
  - [ ] Check Gemini provider error messages
  - [ ] Check Ollama provider error messages
  - [ ] Verify consistent error format across providers
- [ ] Verify all examples run successfully
  - [ ] `examples/simple_chat` with all 3 providers
  - [ ] `examples/openai_chat` with real API
  - [ ] `examples/gemini_chat` with real API
  - [ ] `examples/multi_provider` with all providers
  - [ ] `examples/builtin_tools` with file and datetime tools
- [ ] Clean up commented code and TODOs
  - [ ] Search for `// TODO` comments
  - [ ] Remove or document all TODOs
  - [ ] Clean up debug print statements
- [ ] Check for unused imports and dead code

#### 2. Final Testing (1-2 hours)
- [ ] Run full test suite
  ```bash
  go test ./... -v -cover
  ```
- [ ] Verify test coverage >= 71.8%
  ```bash
  go test ./... -coverprofile=coverage.out
  go tool cover -html=coverage.out
  ```
- [ ] Test all examples with all 3 providers
  - [ ] Set LLM_PROVIDER=ollama and test
  - [ ] Set LLM_PROVIDER=openai and test
  - [ ] Set LLM_PROVIDER=gemini and test
- [ ] Verify documentation accuracy
  - [ ] README examples run without errors
  - [ ] QUICKSTART examples work
  - [ ] MIGRATION guide examples are correct

#### 3. Version Tagging & Release Notes (30-60 min)
- [ ] Create comprehensive CHANGELOG.md
  - [ ] Document all changes since v0.1.0
  - [ ] Breaking changes (none expected)
  - [ ] New features (multi-provider, factory pattern, 10 built-in tools)
  - [ ] Bug fixes
  - [ ] Documentation updates
- [ ] Write v0.2.0 release notes
  - [ ] Highlight key features
  - [ ] Include code examples
  - [ ] Migration guide summary
  - [ ] Known limitations
- [ ] Update version in relevant files
  - [ ] Check go.mod version comments
  - [ ] Update README badges if needed
- [ ] Create git tag v0.2.0
  ```bash
  git tag -a v0.2.0 -m "Release v0.2.0: Multi-Provider Support + Built-in Tools"
  ```

#### 4. Release & Publish (30 min)
- [ ] Push v0.2.0 tag to GitHub
  ```bash
  git push origin v0.2.0
  ```
- [ ] Create GitHub Release
  - [ ] Attach CHANGELOG
  - [ ] Add release notes highlighting 10 built-in tools
  - [ ] Link to documentation
  - [ ] Include upgrade instructions
- [ ] Verify pkg.go.dev updates automatically
- [ ] Update README badges if needed
- [ ] Optional: Announce release (Twitter, Reddit, etc.)

### Acceptance Criteria

- ✅ All tests pass (100% success rate)
- ✅ Code coverage >= 75% (improved with tool tests)
- ✅ All 5 examples work with all applicable providers
- ✅ All 10 built-in tools tested and working
- ✅ Built-in tools have 100% test coverage
- ✅ Documentation is complete and accurate
- ✅ v0.2.0 tag created and pushed
- ✅ GitHub release published with comprehensive notes
- ✅ No broken links in documentation
- ✅ CHANGELOG.md is complete

---

## 📦 AFTER v0.2.0: v0.3.0 Planning - System & Data Tools

**Target Date**: November 2025  
**Estimated Time**: 2 weeks  
**Status**: ⏸️ FUTURE (after v0.2.0 release)

**Note**: Phase 1 tools (10 tools) will be completed BEFORE v0.2.0 release.  
v0.3.0 will focus on Phase 2-3 tools (8 additional tools).

### Phase 2: System & Data Tools (6 tools)

#### System Tools (3/3 new)
- [ ] **system_exec** - Execute shell commands (DANGEROUS)
  - [ ] Implement with strict whitelist (est. 200 lines)
  - [ ] Command whitelist configuration
  - [ ] Timeout enforcement
  - [ ] Output size limits
  - [ ] Mark as unsafe, requires confirmation
  - [ ] Comprehensive security tests

- [ ] **system_env** - Get environment variables
  - [ ] Implement env getter (est. 80 lines)
  - [ ] Filter sensitive variables (passwords, keys)
  - [ ] Unit tests

- [ ] **system_info** - System information
  - [ ] Implement system info tool (est. 150 lines)
  - [ ] CPU, memory, disk info
  - [ ] OS information
  - [ ] Unit tests

#### Data Tools (3/3 new)
- [ ] **data_json_parse** - JSON parsing and querying
  - [ ] Implement JSON tool (est. 180 lines)
  - [ ] JSONPath query support
  - [ ] Pretty printing
  - [ ] Validation
  - [ ] Unit tests

- [ ] **data_csv_parse** - CSV processing
  - [ ] Implement CSV tool (est. 150 lines)
  - [ ] Parse to JSON
  - [ ] Custom delimiters
  - [ ] Header detection
  - [ ] Unit tests

- [ ] **data_xml_parse** - XML processing
  - [ ] Implement XML tool (est. 150 lines)
  - [ ] XPath support
  - [ ] Parse to JSON
  - [ ] Unit tests

**Deliverables**:
- 6 new tools implemented
- ~900 lines of production code
- ~400 lines of tests

### Phase 3: Math Tools & Polish (2 tools)

#### Math Tools (2/2 new)
- [ ] **math_calculate** - Mathematical calculations
  - [ ] Implement calculator (est. 120 lines)
  - [ ] Safe expression evaluation
  - [ ] Prevent code injection
  - [ ] Unit tests

- [ ] **math_stats** - Statistical functions
  - [ ] Implement stats tool (est. 100 lines)
  - [ ] Mean, median, mode, stddev
  - [ ] Dataset validation
  - [ ] Unit tests

#### Documentation & Examples
- [ ] Update BUILTIN_TOOLS_DESIGN.md
  - [ ] Mark completed tools
  - [ ] Update progress tracking
- [ ] Create comprehensive examples
  - [ ] File operations agent example
  - [ ] Web scraper agent example
  - [ ] Data processor agent example
  - [ ] Multi-tool agent example
- [ ] Write tool usage guide
  - [ ] Best practices
  - [ ] Security considerations
  - [ ] Performance tips

#### Testing & Quality
- [ ] Achieve 100% test coverage for tools
- [ ] Security audit for all tools
  - [ ] Path traversal tests
  - [ ] Command injection tests
  - [ ] SSRF tests
  - [ ] Size limit tests
- [ ] Performance benchmarks
- [ ] Integration tests with all providers

**Week 3 Deliverables**:
- 2 new tools implemented
- Complete documentation
- 100% test coverage
- All 18 tools ready for production

---

## 🔮 v0.4.0 Planning - Advanced Features

**Target Date**: December 2025  
**Status**: 💡 IDEAS

### Agent Builder Pattern
- [ ] Create fluent builder API for agents
- [ ] Pre-configured agent templates
- [ ] Tool presets (file agent, web agent, etc.)
- [ ] Conversation flow management

### Persistent Memory
- [ ] SQLite backend implementation
- [ ] PostgreSQL backend implementation
- [ ] Message persistence and retrieval
- [ ] Conversation search

### Vector Database Integration
- [ ] Vector store interface
- [ ] Qdrant integration
- [ ] Weaviate integration
- [ ] Semantic search for memory

### Multi-Agent Coordination
- [ ] Agent-to-agent communication
- [ ] Shared memory spaces
- [ ] Task delegation patterns
- [ ] Consensus mechanisms

### Advanced Streaming
- [ ] Function calling in streaming mode
- [ ] Parallel tool execution
- [ ] Progressive result updates

### Cost Tracking & Monitoring
- [ ] Token usage tracking
- [ ] Cost calculation (OpenAI, Gemini)
- [ ] Usage analytics dashboard
- [ ] Budget alerts

---

## 📚 Documentation Improvements

### High Priority
- [ ] Add GIF demos to README
  - [ ] Streaming response demo
  - [ ] Tool calling demo
  - [ ] Multi-provider switching demo
- [ ] Create video tutorial
  - [ ] Quick start (5 min)
  - [ ] Building custom tools (10 min)
  - [ ] Production deployment (15 min)

### Medium Priority
- [ ] Write blog post about multi-provider architecture
- [ ] Create comparison with other frameworks
  - [ ] LangChain comparison
  - [ ] AutoGen comparison
  - [ ] Semantic Kernel comparison
- [ ] Add more code examples in README
- [ ] Create architecture diagrams
  - [ ] Component diagram
  - [ ] Sequence diagrams
  - [ ] Deployment diagram

### Low Priority
- [ ] Set up documentation website
- [ ] Add interactive playground
- [ ] Create Cookbook with recipes

---

## 🔧 Quality & Infrastructure

### Testing
- [ ] Add benchmarks for all providers
- [ ] Create performance comparison charts
- [ ] Set up continuous benchmarking
- [ ] Add chaos/fuzz testing

### CI/CD
- [ ] GitHub Actions workflow for tests
  - [ ] Run tests on every PR
  - [ ] Test with all Go versions (1.24, 1.25)
  - [ ] Generate coverage reports
- [ ] Set up Ollama in CI for integration tests
- [ ] Automated release process
- [ ] Dependabot configuration

### Code Quality
- [ ] Set up golangci-lint
- [ ] Add pre-commit hooks
- [ ] Code complexity analysis
- [ ] Dependency vulnerability scanning

---

## 🤝 Community & Contribution

### GitHub Setup
- [ ] Create issue templates
  - [ ] Bug report template
  - [ ] Feature request template
  - [ ] Question template
  - [ ] Documentation improvement template
- [ ] Create PR template
  - [ ] Checklist for contributors
  - [ ] Testing requirements
  - [ ] Documentation requirements
- [ ] Add CONTRIBUTING.md (if not exists)
- [ ] Add CODE_OF_CONDUCT.md
- [ ] Set up GitHub Discussions

### Community Building
- [ ] Create examples gallery
- [ ] Collect user stories
- [ ] Set up Discord/Slack community (optional)
- [ ] Regular release cadence

---

## 📊 Metrics & Goals

### v0.2.0 Goals (Current)
- ✅ 3 providers (Ollama, OpenAI, Gemini)
- ✅ Test coverage >= 71.8%
- ✅ 8 working examples
- ✅ Comprehensive documentation
- ⏸️ GitHub release published

### v0.3.0 Goals
- 18 built-in tools (6 categories)
- Test coverage >= 80%
- 100% built-in tools coverage
- 12+ working examples
- Tool usage guide

### v0.4.0 Goals
- Agent builder pattern
- Persistent memory (2 backends)
- Vector database integration
- Multi-agent support
- Cost tracking dashboard

---

## 🎯 Success Criteria

### Project Health
- [ ] Active development (weekly commits)
- [ ] Responsive to issues (< 3 days)
- [ ] Regular releases (monthly)
- [ ] Growing user base
- [ ] Community contributions

### Code Quality
- [ ] Test coverage >= 80%
- [ ] No critical bugs
- [ ] Clean code (linter passing)
- [ ] Up-to-date dependencies
- [ ] Security audit passed

### Documentation
- [ ] Complete API docs
- [ ] Working examples for all features
- [ ] Migration guides for breaking changes
- [ ] Video tutorials available
- [ ] Active Q&A (discussions/issues)

---

**Legend:**  
- ✅ = Completed  
- 🔄 = In Progress  
- ⏸️ = Pending  
- 💡 = Idea/Planned  
- ❌ = Blocked/Cancelled

**Last Updated**: October 27, 2025
