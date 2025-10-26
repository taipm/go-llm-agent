# TODO List

## Immediate (Before First Use)

- [ ] Test with actual Ollama instance
- [ ] Verify all 3 examples work end-to-end
- [ ] Check tool calling with Ollama (may need model that supports function calling)
- [ ] Add more error handling edge cases
- [ ] Create GitHub repository
- [ ] Tag v0.1.0 release

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

## v0.2 Preparation

### Streaming Support
- [ ] Design streaming API
- [ ] Implement streaming in Ollama provider
- [ ] Add streaming examples
- [ ] Document streaming usage

### Built-in Tools
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

### Multi-Provider Support
- [ ] OpenAI provider
- [ ] Azure OpenAI provider
- [ ] Anthropic Claude provider
- [ ] Generic HTTP provider
- [ ] Provider fallback mechanism

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

**Current Focus**: Test v0.1 with real Ollama → Release → Start v0.2
