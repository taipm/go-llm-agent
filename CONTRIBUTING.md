# Contributing to go-llm-agent

Thanks for your interest in contributing! 🎉

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Ollama installed and running
- Git

### Getting Started

```bash
# Clone repository
git clone https://github.com/taipm/go-llm-agent.git
cd go-llm-agent

# Install dependencies
go mod download

# Run tests
make test

# Run examples
make run-simple
```

## Development Workflow

1. **Fork & Clone**: Fork the repository and clone your fork
2. **Branch**: Create a feature branch (`git checkout -b feature/amazing-feature`)
3. **Code**: Make your changes
4. **Test**: Add tests and ensure all tests pass (`make test`)
5. **Commit**: Commit your changes (`git commit -m 'Add amazing feature'`)
6. **Push**: Push to your fork (`git push origin feature/amazing-feature`)
7. **PR**: Open a Pull Request

## Code Guidelines

### Go Best Practices

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting (run `make fmt`)
- Add comments for exported functions
- Write tests for new features

### Testing

- Maintain test coverage >= 70%
- Write unit tests in `*_test.go` files
- Use table-driven tests where appropriate

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test code
        })
    }
}
```

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add streaming support for Ollama provider
fix: handle nil pointer in memory buffer
docs: update README with installation guide
test: add tests for tool registry
```

## What to Contribute

### Good First Issues

- Documentation improvements
- Example programs
- Built-in tools (calculator, file operations, etc.)
- Test coverage improvements

### Feature Contributions

- New LLM providers (OpenAI, Anthropic, etc.)
- Advanced memory strategies
- Tool composition features
- Performance optimizations

### Bug Reports

When reporting bugs, include:

1. Go version
2. Ollama version (if applicable)
3. Minimal reproduction code
4. Expected vs actual behavior
5. Error messages/logs

## Pull Request Process

1. Update documentation if needed
2. Add tests for new features
3. Ensure all tests pass
4. Update CHANGELOG.md (if we have one)
5. Request review from maintainers

## Code of Conduct

- Be respectful and inclusive
- Accept constructive feedback
- Focus on what's best for the project
- Help others learn and grow

## Questions?

- Open an issue for questions
- Check existing issues/PRs first
- Join discussions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🙏
