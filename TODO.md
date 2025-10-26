# TODO List

## ✅ v0.1.0 Released (Oct 26, 2025)

- [x] Test with actual Ollama instance
- [x] Verify all examples work end-to-end
- [x] Create GitHub repository (https://github.com/taipm/go-llm-agent)
- [x] Tag v0.1.0 release
- [x] Publish to public GitHub
- [x] Verify go get works from external projects
- [x] Implement streaming support (bonus feature!)
- [x] Create 5 working examples
- [x] Comprehensive documentation suite

## Immediate Next Steps (Priority Order)

### 1. Documentation Updates for Streaming
**Priority**: High | **Effort**: Medium | **Impact**: High

- [ ] Add streaming section to SPEC.md
  - Document StreamChunk type structure
  - Explain StreamHandler callback pattern
  - Show error handling in streaming
  - Document memory integration during streaming
- [ ] Update QUICKSTART.md with streaming example
  - Basic streaming example with explanation
  - Advanced streaming with metrics
  - Best practices for handler implementation
- [ ] Add streaming API reference to README.md
  - ChatStream() method signature
  - StreamHandler callback interface
  - Real-time output patterns

**Acceptance Criteria**:
- Users can understand streaming API without reading code
- Clear examples of both basic and advanced usage
- Error handling patterns documented

---

### 2. Streaming Unit Tests ✅
**Priority**: High | **Effort**: Medium | **Impact**: High | **Status**: COMPLETED

- [x] Create pkg/provider/ollama/ollama_test.go
  - Test Stream() with real Ollama instance
  - Test JSON streaming responses
  - Test error handling (cancellation, handler errors, timeouts)
  - Test partial chunk handling
  - Test done flag handling
- [x] Add streaming tests to pkg/agent/agent_test.go
  - Test ChatStream() with real provider
  - Test handler callback invocation
  - Test response accumulation for memory
  - Test error propagation to handler
  - Test concurrent streaming calls

**Acceptance Criteria**: ✅
- All streaming paths covered
- Edge cases tested (connection drops, cancellation, errors)
- Test coverage: 71.8% overall
  - memory: 92.1%
  - tool: 93.2%
  - ollama: 66.7%
  - agent: 61.0%

**Notes**: Integration tests with real Ollama (qwen3:1.7b). 23 tests total, all passing.

---

### 3. Integration Test with Real Ollama
**Priority**: High | **Effort**: Low | **Impact**: High

- [ ] Create tests/integration_test.go
  - Test Chat() with real Ollama instance
  - Test ChatStream() with real streaming
  - Test tool calling with function-capable models
  - Test memory persistence across calls
  - Test error scenarios (Ollama not running, wrong model)
- [ ] Add GitHub Actions workflow
  - Set up Ollama in CI environment
  - Run integration tests automatically
  - Generate coverage reports
- [ ] Document how to run integration tests locally

**Acceptance Criteria**:
- Integration tests pass with real Ollama
- CI pipeline runs tests on every PR
- Clear instructions for local testing

---

### 4. Test in go-youtube-channel Project ✅
**Priority**: Medium | **Effort**: Low | **Impact**: High | **Status**: COMPLETED

- [x] Integrate go-llm-agent into go-agent service
  - Replace existing LLM integration with go-llm-agent
  - Update go.mod to use published module
  - Remove duplicate agent code
- [x] Create agent for content planning
  - Use Calculator tool for metrics
  - Use Weather tool for seasonal content ideas
  - Implement custom tools (YouTube API, analytics)
- [x] Test streaming for real-time feedback
  - Stream content ideas to user
  - Show progress during long operations
- [x] Document integration patterns
  - How to create custom tools
  - Best practices for production use
  - Error handling and retry logic

**Acceptance Criteria**: ✅
- go-youtube-channel uses published module successfully
- Custom tools work correctly
- Production-ready error handling implemented

**Notes**: Completed by user, integrated successfully into go-youtube-channel project.

---

### 5. GitHub Release Notes & Examples
**Priority**: Medium | **Effort**: Low | **Impact**: Medium

- [ ] Create comprehensive v0.1.0 release notes
  - Feature highlights with code examples
  - Breaking changes (none for initial release)
  - Known limitations
  - Upgrade guide (N/A for v0.1.0)
- [ ] Add GIF demos to README
  - Streaming response in action
  - Tool calling example
  - Multi-turn conversation
- [ ] Create GitHub Release page
  - Attach release notes
  - Link to documentation
  - Show installation instructions
  - Add changelog
- [ ] Set up issue templates
  - Bug report template
  - Feature request template
  - Question template
- [ ] Create PR template
  - Checklist for contributors
  - Testing requirements
  - Documentation requirements

**Acceptance Criteria**:
- Release page looks professional
- Visual demos show key features
- Contributors know how to file issues/PRs

---

## Documentation

- [ ] Add more code examples in README
- [ ] Create API documentation site (pkg.go.dev)
- [ ] Add architecture diagrams
- [ ] Record demo video
- [ ] Write blog post about the project

## Testing

- [ ] Add integration tests with Ollama
- [ ] Add tests for agent.go
- [ ] Add tests for ollama provider
- [ ] Increase test coverage to 80%+
- [ ] Add benchmarks

## v0.2 Preparation - Multi-Provider Support

**FOCUS**: Add OpenAI and Google Gemini providers while maintaining API uniformity

**Reference**: See INTEGRATION_PLAN.md for detailed implementation strategy

### Sprint 1: OpenAI Provider (Week 1)
**Priority**: CRITICAL | **Effort**: 5 days | **Impact**: HIGH

- [ ] Research & design (1 day) - **See INTEGRATION_PLAN.md**
- [ ] OpenAI provider implementation (2 days)
  - Create `pkg/provider/openai/openai.go` (~300 lines)
  - Pin SDK version: `github.com/openai/openai-go/v3 v3.6.1`
  - Implement `types.LLMProvider` interface (Chat, Stream)
  - Constructor: `New(apiKey, model string) *Provider`
- [ ] Message/tool conversion layer (1 day)
  - Create `pkg/provider/openai/converter.go` (~200 lines)
  - Convert `types.Message` ↔ `openai.ChatCompletionMessageParamUnion`
  - Convert `types.ToolDefinition` ↔ `openai.ChatCompletionToolParam`
  - Handle streaming chunks conversion
- [ ] Tests & examples (1 day)
  - Create `pkg/provider/openai/openai_test.go` (~400 lines)
  - Integration tests with real OpenAI API
  - Target: 70%+ coverage
  - Create `examples/openai_chat/main.go` (~100 lines)
  - Mirror structure of `examples/simple_chat`
- [ ] Documentation (0.5 day)
  - Update README.md with OpenAI examples
  - Add to QUICKSTART.md
  - Document API key setup

**Acceptance Criteria**:
- ✅ API identical to Ollama provider (switch = 1 line change)
- ✅ All tests pass with real OpenAI API
- ✅ Example works out of the box
- ✅ Documentation shows provider switching

---

### Sprint 2: Google Gemini Provider (Week 2)
**Priority**: CRITICAL | **Effort**: 5 days | **Impact**: HIGH

- [ ] Gemini provider implementation (2 days)
  - Create `pkg/provider/gemini/gemini.go` (~250 lines)
  - Pin SDK version: `google.golang.org/genai v1.32.0`
  - Implement `types.LLMProvider` interface
  - Constructor: `New(apiKey, model string) (*Provider, error)`
  - Support both Gemini API and Vertex AI backends
- [ ] Multimodal support extension (1 day)
  - Create `pkg/provider/gemini/converter.go` (~150 lines)
  - Extend `types.Message` with optional `Parts` field for images
  - Support inline data (base64 images)
  - Maintain backward compatibility (text-only still works)
- [ ] Tests & examples (1 day)
  - Create `pkg/provider/gemini/gemini_test.go` (~300 lines)
  - Integration tests with real Gemini API
  - Create `examples/gemini_chat/main.go` - basic text chat
  - Create `examples/gemini_multimodal/main.go` - vision example
- [ ] Documentation (0.5 day)
  - Update README with Gemini examples
  - Document multimodal usage
  - Add provider comparison table

**Acceptance Criteria**:
- ✅ API identical to OpenAI/Ollama for basic chat
- ✅ Multimodal is opt-in extension
- ✅ Tests pass with real Gemini API
- ✅ Examples show both text and vision

---

### Sprint 3: Integration & Polish (Week 3)
**Priority**: HIGH | **Effort**: 5 days | **Impact**: MEDIUM

- [ ] Factory pattern (1 day)
  - Create `pkg/provider/factory.go` (~150 lines)
  - `New(config Config) (types.LLMProvider, error)`
  - `FromEnv()` - auto-detect from environment variables
  - Support provider types: ollama, openai, gemini
- [ ] Unified configuration (1 day)
  - Environment variable support (PROVIDER, *_API_KEY, etc.)
  - Config struct with validation
  - Examples: `examples/multi_provider/main.go`
- [ ] Comprehensive tests (1 day)
  - Factory tests: `pkg/provider/factory_test.go`
  - Cross-provider compatibility tests
  - Ensure identical behavior where possible
- [ ] Examples for all providers (1 day)
  - Update all existing examples to show provider switching
  - Add README per example explaining differences
  - Provider fallback example
- [ ] Final documentation (1 day)
  - Complete README overhaul with multi-provider
  - Update SPEC.md with provider section
  - Provider migration guide
  - API uniformity guarantees

**Acceptance Criteria**:
- ✅ Switch provider via environment variable
- ✅ All providers work identically from user code
- ✅ Documentation clear and comprehensive
- ✅ Ready for v0.2.0 release

---

### Dependencies Update
**Priority**: CRITICAL | **Status**: Planned

Update `go.mod`:
```go
module github.com/taipm/go-llm-agent

go 1.22 // Required by both openai-go and go-genai

require (
    github.com/openai/openai-go/v3 v3.6.1  // Pinned for stability
    google.golang.org/genai v1.32.0        // Pinned for stability
)
```

**Version Pin Strategy**:
- Review SDK changelogs quarterly
- Test in dev branch before updating pins
- Document breaking changes in CHANGELOG.md

---

### Quality Gates (All Providers Must Pass)

- [ ] Implements `types.LLMProvider` interface exactly
- [ ] Constructor follows pattern: `New(apiKey, model string)`
- [ ] Chat() signature: `Chat(ctx, messages, options) (*Response, error)`
- [ ] Stream() signature: `Stream(ctx, messages, options, handler) error`
- [ ] Error handling returns `types.ProviderError`
- [ ] Example code structure mirrors `examples/simple_chat`
- [ ] Tests achieve 70%+ coverage
- [ ] Documentation shows 1-line provider switch

---

### Streaming Support ✅

- [x] Design streaming API
- [x] Implement streaming in Ollama provider
- [x] Add streaming examples (basic + advanced)
- [ ] Document streaming usage in SPEC.md
- [x] Add streaming unit tests (71.8% coverage)
- [ ] Add streaming integration tests

---

### Built-in Tools (Post Multi-Provider)
- [ ] File operations (read, write, list, delete)
- [ ] HTTP requests (GET, POST, PUT, DELETE)
- [ ] JSON/YAML parser
- [ ] Shell command execution
- [ ] Time/date operations
- [ ] String manipulation
- [ ] Web scraping
- [ ] Database queries (SQL)
- [ ] CSV processing
- [ ] Math operations (extended)

### Configuration
- [ ] Config file support (YAML/JSON)
- [ ] Environment variables
- [ ] Logging framework integration
- [ ] Metrics collection
- [ ] Timeout policies
- [ ] Retry strategies

### Memory Enhancements
- [ ] Summary memory
- [ ] Token counting
- [ ] Smart truncation
- [ ] Conversation branching
- [ ] Export/import conversations

## v0.3 and Beyond

### Additional Provider Support
- [ ] Azure OpenAI provider (separate from OpenAI)
- [ ] Anthropic Claude provider
- [ ] Cohere provider
- [ ] AI21 Labs provider
- [ ] Generic HTTP provider (custom endpoints)
- [ ] Provider fallback/retry mechanism
- [ ] Provider pooling/load balancing

### Persistent Storage
- [ ] SQLite backend
- [ ] PostgreSQL backend
- [ ] Vector database integration (Qdrant, Weaviate)
- [ ] Conversation export formats

### Advanced Features
- [ ] ReAct pattern
- [ ] Chain of thought
- [ ] Self-correction
- [ ] Planning & execution separation
- [ ] Multi-agent collaboration
- [ ] Agent delegation

### Production Features
- [ ] Metrics & monitoring
- [ ] Rate limiting
- [ ] Cost tracking
- [ ] Circuit breaker
- [ ] Health checks
- [ ] Graceful shutdown

### Developer Experience
- [ ] CLI tool for testing
- [ ] Web UI for debugging
- [ ] Prompt template system
- [ ] Hot reload for development
- [ ] Better error messages

## Community & Ecosystem

- [ ] Set up GitHub discussions
- [ ] Create Discord/Slack community
- [ ] Add code of conduct
- [ ] Create issue templates
- [ ] Add PR templates
- [ ] Set up CI/CD (GitHub Actions)
- [ ] Automated releases
- [ ] Changelog automation

## Performance

- [ ] Profile memory usage
- [ ] Optimize hot paths
- [ ] Add connection pooling
- [ ] Implement caching where appropriate
- [ ] Benchmark vs competitors

## Security

- [ ] Security audit
- [ ] Input validation
- [ ] API key management
- [ ] Rate limiting
- [ ] Sandboxing for tool execution

## Nice to Have

- [ ] VS Code extension
- [ ] Web playground
- [ ] Docker images
- [ ] Kubernetes manifests
- [ ] Example deployments
- [ ] Integration with popular frameworks
- [ ] Plugin system for extensions

---

**Priority Legend**:

- Must have: Required for version completion
- Should have: Important but not blocking
- Nice to have: Would be great but optional

**Current Focus**: v0.1.0 released! Next: **Multi-Provider Support (OpenAI + Gemini)**

**Timeline**: 3 weeks to complete v0.2.0
- Week 1: OpenAI provider (v3.6.1)
- Week 2: Gemini provider (v1.32.0)
- Week 3: Integration + factory pattern

**Release Info**:

- v0.1.0: Oct 26, 2025 - Initial release with Ollama + streaming support
- v0.2.0: Mid-Nov 2025 (planned) - Multi-provider (Ollama + OpenAI + Gemini)
- Repository: <https://github.com/taipm/go-llm-agent>
- Go Module: `go get github.com/taipm/go-llm-agent@v0.1.0`

**See Also**: INTEGRATION_PLAN.md for detailed multi-provider architecture and implementation guide
