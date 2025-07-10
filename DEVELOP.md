# Development Guide

This document contains development-specific information for contributors and maintainers of AI Rebaser.

## Development

### Running Tests

The project includes comprehensive unit and integration tests:

```bash
# Run all tests (unit tests only, no API calls)
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run specific test suite
go test ./internal/notify -v
go test ./internal/config -v
go test ./internal/git -v
```

#### Integration Tests with Real APIs

Integration tests demonstrate real-world scenarios with actual OpenAI API calls:

```bash
# Set required environment variable
export OPENAI_API_KEY="sk-your-api-key-here"

# Run integration tests (makes real API calls, costs ~$0.10-0.50)
go test ./test/integration -v -timeout 10m

# Run specific integration test scenarios
go test ./test/integration -v -run "TestRealWorldRebaseWorkflow"
go test ./test/integration -v -run "TestConflictScenarios"
go test ./test/integration -v -run "TestEndToEndWorkflow"

# Run error handling tests (no API calls required)
go test ./test/integration -v -run "TestErrorHandling"
```

**Note**: Integration tests require a valid OpenAI API key and will make real API calls. Estimated cost per full test run is $0.10-0.50.

### Test Structure

- **Unit Tests**: Fast, isolated tests using mocks
  - `cmd/ai-rebaser/main_test.go` - Core orchestration logic
  - `internal/config/config_test.go` - Configuration loading
  
- **Integration Tests**: End-to-end workflow demonstrations
  - `cmd/ai-rebaser/integration_test.go` - Full workflow testing
  
- **Mocks**: Complete service mocks for testing
  - `internal/mocks/` - All service interface mocks

### Code Quality

```bash
# Format code
go fmt ./...

# Run linting (if golangci-lint is installed)
golangci-lint run

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Building

```bash
# Build for current platform
go build -o ai-rebaser ./cmd/ai-rebaser

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o ai-rebaser-linux ./cmd/ai-rebaser

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o ai-rebaser.exe ./cmd/ai-rebaser

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o ai-rebaser-mac ./cmd/ai-rebaser
```

## Contributing

### Commit Message Style

This project uses [Conventional Commits](https://www.conventionalcommits.org/) format:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for code style changes
- `refactor:` for code refactoring
- `test:` for adding tests
- `chore:` for maintenance tasks

Examples:
```bash
feat: add AI conflict resolution service
fix: resolve git rebase merge conflicts
docs: update installation instructions
```

### Development Workflow

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature`
3. **Write tests** for your changes
4. **Implement your feature**
5. **Run tests**: `go test ./...`
6. **Commit with conventional format**: `git commit -s -m "feat: add your feature"`
7. **Push and create PR**: `git push origin feature/your-feature`

### Code Standards

- Follow Go best practices and idioms
- Write comprehensive tests for new functionality
- Use structured logging with logrus
- Implement proper error handling
- Document public APIs
- Use dependency injection for testability

## Implementation Status

### ✅ Completed
- **Core orchestration logic and workflow** - Six-phase rebase process
- **Configuration management** - YAML-based with environment variable overrides
- **Service interface definitions** - Clean architecture with dependency injection
- **Git operations** - Hybrid go-git/command approach for comprehensive Git support
- **OpenAI API integration** - Real AI conflict resolution, commit messages, and PR descriptions
- **GitHub API integration** - Full PR management with rebase-based merging
- **Slack webhook notifications** - Rich notifications with attachments and error handling
- **Comprehensive testing** - Unit tests, integration tests, and real-world scenarios
- **Mock-based testing framework** - Complete service mocks for isolated testing
- **CLI interface with Kong** - User-friendly command-line interface
- **Environment variable support** - Secure configuration through environment variables
- **Project structure and documentation** - Production-ready codebase organization

### 🔄 In Progress
- Configuration-driven test commands
- Enhanced retry logic and error handling

### 📋 Planned
- Advanced conflict resolution strategies
- Metrics and monitoring integration
- Docker containerization
- Kubernetes deployment manifests
- CI/CD pipeline examples
- Performance optimizations
- Multi-repository support