# Integration Tests for AI Rebaser

This directory contains comprehensive integration tests that test the AI Rebaser in real-world scenarios with actual Git repositories, AI conflict resolution, and error handling.

## Test Structure

### 📁 Test Files

- **`real_world_test.go`** - Complete end-to-end workflow testing
- **`conflict_scenarios_test.go`** - Different types of realistic conflicts
- **`end_to_end_test.go`** - Full workflow with all services
- **`run_integration_tests.sh`** - Test runner script

### 📁 Test Data

- **`testdata/integration-config.yaml`** - Configuration for integration tests

## Running Integration Tests

### Prerequisites

1. **OpenAI API Key**: Set your OpenAI API key as an environment variable:
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

2. **Git**: Ensure Git is installed and available in your PATH

3. **Go**: Go 1.21 or later

### Running Tests

#### Option 1: Using the Test Runner (Recommended)

```bash
# From the project root directory
./test/integration/run_integration_tests.sh
```

#### Option 2: Manual Test Execution

```bash
# Run all integration tests
go test ./test/integration -v

# Run specific test suites
go test ./test/integration -v -run "TestRealWorldRebaseWorkflow"
go test ./test/integration -v -run "TestConflictScenarios"
go test ./test/integration -v -run "TestEndToEndWorkflow"
go test ./test/integration -v -run "TestErrorHandling"

# Run with longer timeout for complex scenarios
go test ./test/integration -v -timeout 10m
```

#### Option 3: Testing Without OpenAI (Limited)

```bash
# Run error scenarios only (no OpenAI API calls)
go test ./test/integration -v -run "TestErrorHandling"
```

## Test Scenarios

### 🔄 Real-World Workflow Test

**File**: `real_world_test.go`

Tests the complete rebase workflow:

1. **Repository Setup**: Creates upstream and internal repositories
2. **Conflict Creation**: Introduces realistic conflicts
3. **AI Resolution**: Uses OpenAI to resolve conflicts
4. **Validation**: Verifies the resolution quality
5. **End-to-End Flow**: Complete workflow from setup to completion

**What it tests:**
- Git repository cloning and setup
- Branch creation and rebase operations
- Conflict detection and parsing
- AI-powered conflict resolution
- Commit message generation
- PR description generation
- Repository state validation

### 🥊 Conflict Scenarios Test

**File**: `conflict_scenarios_test.go`

Tests various types of realistic conflicts:

1. **Simple String Conflicts** - Version numbers, configuration values
2. **Function Signature Conflicts** - Parameter changes, return types
3. **Import Conflicts** - Different package imports
4. **Configuration Conflicts** - JSON/YAML configuration files
5. **Comment Conflicts** - Documentation and comment changes

**What it tests:**
- Different conflict types and complexities
- AI's ability to understand context
- Resolution quality validation
- Code structure preservation

### 🔧 End-to-End Workflow Test

**File**: `end_to_end_test.go`

Tests the complete application workflow:

1. **Complete Go Project Setup** - Realistic project structure
2. **Service Integration** - All services working together
3. **Test Execution** - Running actual tests after resolution
4. **Build Validation** - Ensuring code compiles
5. **PR Creation** - Mock PR generation
6. **Notifications** - Mock notification sending

**What it tests:**
- Complete service integration
- Test execution framework
- Build validation
- Service orchestration
- Error propagation

### ⚠️ Error Handling Test

**File**: `end_to_end_test.go`

Tests various error scenarios:

1. **Invalid Git Repository** - Non-existent repositories
2. **Network Failures** - Connectivity issues
3. **AI Service Failures** - Invalid API keys, rate limits
4. **Test Execution Failures** - Command failures, timeouts

**What it tests:**
- Error handling robustness
- Graceful failure modes
- Recovery mechanisms
- User feedback quality

## Expected Behavior

### ✅ Successful Test Run

When tests pass, you should see:

- Git repositories created successfully
- Conflicts detected and resolved
- AI-generated content (commit messages, PR descriptions)
- No conflict markers in resolved files
- Tests passing after conflict resolution
- Clean repository state

### ❌ Common Issues

1. **OpenAI API Key Issues**
   ```
   Error: OpenAI API call failed: invalid API key
   ```
   **Solution**: Verify your API key is set correctly

2. **Git Configuration Issues**
   ```
   Error: failed to clone repository
   ```
   **Solution**: Ensure Git is installed and configured

3. **Network Connectivity**
   ```
   Error: timeout connecting to OpenAI
   ```
   **Solution**: Check network connectivity and API status

## Test Configuration

The integration tests use a separate configuration file:

```yaml
# testdata/integration-config.yaml
ai:
  model: "gpt-3.5-turbo"  # Cheaper for testing
  max_tokens: 1000        # Reduced for faster tests
  
tests:
  timeout: 5m             # Reasonable timeout
  commands:               # Basic Go commands
    - go mod tidy
    - go build
    - go test
```

## Customizing Tests

### Adding New Conflict Scenarios

1. Create a new setup function in `conflict_scenarios_test.go`
2. Add corresponding validation function
3. Add to the scenarios slice in `TestConflictScenarios`

### Testing Different Models

Change the model in test configuration:

```go
cfg := &config.Config{
    AI: config.AIConfig{
        Model: "gpt-4", // or "gpt-3.5-turbo"
        MaxTokens: 2000,
    },
}
```

### Adding Custom Test Commands

Modify the test configuration:

```go
Tests: config.TestsConfig{
    Commands: []config.TestCommand{
        {Name: "lint", Command: "golangci-lint", Args: []string{"run"}},
        {Name: "security", Command: "gosec", Args: []string{"./..."}},
    },
}
```

## Performance Considerations

- **OpenAI API Calls**: Each test makes real API calls (~$0.01-0.05 per test)
- **Git Operations**: Repository creation can be slow on some systems
- **Test Timeout**: Default 10 minutes should be sufficient
- **Parallel Execution**: Tests run sequentially to avoid API rate limits

## Troubleshooting

### Verbose Output

Run with verbose logging:

```bash
go test ./test/integration -v -run "TestRealWorldRebaseWorkflow" 2>&1 | tee test-output.log
```

### Debug AI Responses

The tests log AI responses and token usage:

```
INFO[0002] AI conflict resolution completed  component=ai file=main.go tokens_used=245
INFO[0003] AI commit message generated       component=ai message="feat: merge upstream v2.0 features with internal customizations" tokens_used=87
```

### Git Debug

Enable Git debug output:

```bash
GIT_TRACE=1 go test ./test/integration -v -run "TestRealWorldRebaseWorkflow"
```

## Contributing

When adding new integration tests:

1. Follow the existing test structure
2. Use realistic scenarios
3. Validate AI output quality
4. Test both success and failure cases
5. Document expected behavior
6. Keep tests focused and maintainable

## Cost Considerations

Integration tests make real OpenAI API calls:

- **Estimated cost per full test run**: $0.10-0.50
- **Monthly development cost**: $5-20 (depends on usage)
- **Production impact**: Tests use gpt-3.5-turbo to minimize costs

Consider using OpenAI credits or setting up billing alerts for development.