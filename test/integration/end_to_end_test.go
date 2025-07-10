package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BlindspotSoftware/rebAIser/internal/ai"
	"github.com/BlindspotSoftware/rebAIser/internal/config"
	"github.com/BlindspotSoftware/rebAIser/internal/git"
	"github.com/BlindspotSoftware/rebAIser/internal/github"
	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
	"github.com/BlindspotSoftware/rebAIser/internal/notify"
	"github.com/BlindspotSoftware/rebAIser/internal/test"
)

// TestEndToEndWorkflow tests the complete AI Rebaser workflow from start to finish
func TestEndToEndWorkflow(t *testing.T) {
	// Skip if no OpenAI API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping end-to-end test - set OPENAI_API_KEY environment variable to run")
	}

	tempDir := t.TempDir()
	
	// Create test scenario
	scenario := createCompleteTestScenario(t, tempDir)
	
	// Test the complete workflow
	testCompleteWorkflow(t, scenario, apiKey)
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	
	t.Run("InvalidGitRepository", func(t *testing.T) {
		testInvalidGitRepository(t, tempDir)
	})
	
	t.Run("NetworkFailure", func(t *testing.T) {
		testNetworkFailure(t, tempDir)
	})
	
	t.Run("AIServiceFailure", func(t *testing.T) {
		testAIServiceFailure(t, tempDir)
	})
	
	t.Run("TestExecutionFailure", func(t *testing.T) {
		testTestExecutionFailure(t, tempDir)
	})
}

// TestScenario represents a complete test scenario
type TestScenario struct {
	UpstreamDir   string
	InternalDir   string
	WorkDir       string
	Config        *config.Config
	ExpectedFiles []string
}

func createCompleteTestScenario(t *testing.T, tempDir string) *TestScenario {
	t.Helper()
	
	upstreamDir := filepath.Join(tempDir, "upstream")
	internalDir := filepath.Join(tempDir, "internal")
	workDir := filepath.Join(tempDir, "work")
	
	// Create upstream repository with a complete Go project
	createUpstreamGoProject(t, upstreamDir)
	
	// Create internal fork with modifications
	createInternalFork(t, internalDir, upstreamDir)
	
	// Create conflicting changes
	createConflictingGoChanges(t, upstreamDir, internalDir)
	
	cfg := &config.Config{
		Git: config.GitConfig{
			InternalRepo: internalDir,
			UpstreamRepo: upstreamDir,
			WorkingDir:   workDir,
			Branch:       "main",
		},
		AI: config.AIConfig{
			OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
			Model:        "gpt-3.5-turbo",
			MaxTokens:    1000,
		},
		Tests: config.TestsConfig{
			Timeout: 5 * time.Minute,
			Commands: []config.TestCommand{
				{
					Name:    "go-mod-tidy",
					Command: "go",
					Args:    []string{"mod", "tidy"},
				},
				{
					Name:    "go-build",
					Command: "go",
					Args:    []string{"build", "."},
				},
				{
					Name:    "go-test",
					Command: "go",
					Args:    []string{"test", "./..."},
				},
			},
		},
	}
	
	return &TestScenario{
		UpstreamDir:   upstreamDir,
		InternalDir:   internalDir,
		WorkDir:       workDir,
		Config:        cfg,
		ExpectedFiles: []string{"main.go", "utils.go", "go.mod"},
	}
}

func createUpstreamGoProject(t *testing.T, dir string) {
	t.Helper()
	
	require.NoError(t, os.MkdirAll(dir, 0755))
	
	// Initialize git repo
	runGitCommand(t, dir, "init")
	runGitCommand(t, dir, "config", "user.name", "Upstream Team")
	runGitCommand(t, dir, "config", "user.email", "upstream@example.com")
	
	// Create go.mod
	writeFile(t, filepath.Join(dir, "go.mod"), `module example.com/upstream

go 1.21
`)
	
	// Create main.go
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Upstream Application v1.0")
	
	result := Calculate(10, 5)
	fmt.Printf("Calculation result: %d\n", result)
	
	log.Println("Application started successfully")
}
`)
	
	// Create utils.go
	writeFile(t, filepath.Join(dir, "utils.go"), `package main

import "fmt"

// Calculate performs a mathematical operation
func Calculate(a, b int) int {
	return a + b
}

// LogMessage prints a formatted message
func LogMessage(msg string) {
	fmt.Printf("[LOG] %s\n", msg)
}
`)
	
	// Create test file
	writeFile(t, filepath.Join(dir, "utils_test.go"), `package main

import "testing"

func TestCalculate(t *testing.T) {
	result := Calculate(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}
`)
	
	runGitCommand(t, dir, "add", ".")
	runGitCommand(t, dir, "commit", "-m", "Initial upstream project")
}

func createInternalFork(t *testing.T, dir, upstreamDir string) {
	t.Helper()
	
	require.NoError(t, os.MkdirAll(dir, 0755))
	
	// Clone from upstream
	runGitCommand(t, dir, "clone", upstreamDir, ".")
	runGitCommand(t, dir, "config", "user.name", "Internal Team")
	runGitCommand(t, dir, "config", "user.email", "internal@company.com")
	
	// Add internal modifications
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Internal Application v1.0 - Company Edition")
	
	result := Calculate(10, 5)
	fmt.Printf("Internal calculation result: %d\n", result)
	
	InternalFunction()
	
	log.Println("Internal application started successfully")
}

// InternalFunction is a company-specific function
func InternalFunction() {
	fmt.Println("Internal feature enabled")
}
`)
	
	writeFile(t, filepath.Join(dir, "internal_config.go"), `package main

import "fmt"

// InternalConfig holds company-specific configuration
type InternalConfig struct {
	Company string
	Region  string
}

// GetInternalConfig returns the internal configuration
func GetInternalConfig() InternalConfig {
	return InternalConfig{
		Company: "ACME Corp",
		Region:  "US-East",
	}
}

// PrintInternalInfo displays internal information
func PrintInternalInfo() {
	config := GetInternalConfig()
	fmt.Printf("Company: %s, Region: %s\n", config.Company, config.Region)
}
`)
	
	runGitCommand(t, dir, "add", ".")
	runGitCommand(t, dir, "commit", "-m", "Add internal customizations")
}

func createConflictingGoChanges(t *testing.T, upstreamDir, internalDir string) {
	t.Helper()
	
	// Make changes in upstream
	writeFile(t, filepath.Join(upstreamDir, "main.go"), `package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("Upstream Application v2.0 - New Features")
	
	start := time.Now()
	result := Calculate(10, 5)
	fmt.Printf("Calculation result: %d (took %v)\n", result, time.Since(start))
	
	AdvancedFeature()
	
	log.Println("Application v2.0 started successfully")
}

// AdvancedFeature is a new upstream feature
func AdvancedFeature() {
	fmt.Println("Advanced upstream feature enabled")
}
`)
	
	writeFile(t, filepath.Join(upstreamDir, "utils.go"), `package main

import (
	"fmt"
	"math"
)

// Calculate performs advanced mathematical operations
func Calculate(a, b int) int {
	// New upstream logic with validation
	if a < 0 || b < 0 {
		return 0
	}
	return a + b
}

// LogMessage prints a formatted message with timestamp
func LogMessage(msg string) {
	fmt.Printf("[LOG] %s\n", msg)
}

// NewUpstreamFunction is a new utility function
func NewUpstreamFunction(x float64) float64 {
	return math.Sqrt(x)
}
`)
	
	runGitCommand(t, upstreamDir, "add", ".")
	runGitCommand(t, upstreamDir, "commit", "-m", "feat: upgrade to v2.0 with advanced features")
	
	// Make conflicting changes in internal
	writeFile(t, filepath.Join(internalDir, "main.go"), `package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("Internal Application v1.5 - Enhanced Internal Edition")
	
	// Internal environment check
	if env := os.Getenv("INTERNAL_ENV"); env != "" {
		fmt.Printf("Running in environment: %s\n", env)
	}
	
	result := Calculate(10, 5)
	fmt.Printf("Internal calculation result: %d\n", result)
	
	InternalFunction()
	EnhancedInternalFeature()
	
	log.Println("Enhanced internal application started successfully")
}

// InternalFunction is a company-specific function
func InternalFunction() {
	fmt.Println("Internal feature enabled")
}

// EnhancedInternalFeature is a new internal feature
func EnhancedInternalFeature() {
	fmt.Println("Enhanced internal feature v1.5")
}
`)
	
	writeFile(t, filepath.Join(internalDir, "utils.go"), `package main

import (
	"fmt"
	"strings"
)

// Calculate performs internal mathematical operations
func Calculate(a, b int) int {
	// Internal logic with company-specific validation
	if a < 0 || b < 0 {
		return -1 // Different error handling than upstream
	}
	return a + b
}

// LogMessage prints a formatted message with internal prefix
func LogMessage(msg string) {
	fmt.Printf("[INTERNAL-LOG] %s\n", msg)
}

// InternalUtilityFunction is a company-specific utility
func InternalUtilityFunction(s string) string {
	return strings.ToUpper(s)
}
`)
	
	runGitCommand(t, internalDir, "add", ".")
	runGitCommand(t, internalDir, "commit", "-m", "feat: add enhanced internal features v1.5")
}

func testCompleteWorkflow(t *testing.T, scenario *TestScenario, apiKey string) {
	t.Helper()
	
	// Initialize all services
	gitService := git.NewService()
	aiService := ai.NewService("openai", apiKey, "", "gpt-3.5-turbo", 1000)
	githubService := github.NewService("test-token", "test-org", "test-repo")
	notifyService := notify.NewService("https://hooks.slack.com/test", "#test", "Test Bot")
	testService := test.NewService([]interfaces.TestCommand{})
	
	ctx := context.Background()
	
	// Phase 1: Setup working directory
	t.Log("=== Phase 1: Setup ===")
	require.NoError(t, os.MkdirAll(scenario.WorkDir, 0755))
	
	internalWorkDir := filepath.Join(scenario.WorkDir, "internal")
	err := gitService.Clone(ctx, scenario.InternalDir, internalWorkDir)
	require.NoError(t, err)
	
	// Add upstream remote
	runGitCommand(t, internalWorkDir, "remote", "add", "upstream", scenario.UpstreamDir)
	runGitCommand(t, internalWorkDir, "fetch", "upstream")
	
	// Phase 2: Create branch and attempt rebase
	t.Log("=== Phase 2: Rebase ===")
	branchName := fmt.Sprintf("ai-rebase-%d", time.Now().Unix())
	err = gitService.CreateBranch(ctx, internalWorkDir, branchName)
	require.NoError(t, err)
	
	err = gitService.Rebase(ctx, internalWorkDir, "upstream/main")
	require.Error(t, err, "Should have conflicts")
	
	// Phase 3: AI conflict resolution
	t.Log("=== Phase 3: AI Conflict Resolution ===")
	conflicts, err := gitService.GetConflicts(ctx, internalWorkDir)
	require.NoError(t, err)
	require.Greater(t, len(conflicts), 0, "Should have conflicts")
	
	t.Logf("Found %d conflicts to resolve", len(conflicts))
	
	for i, conflict := range conflicts {
		t.Logf("Resolving conflict %d/%d: %s", i+1, len(conflicts), conflict.File)
		
		// Resolve with AI
		resolution, err := aiService.ResolveConflict(ctx, conflict)
		require.NoError(t, err)
		
		// Validate resolution
		assert.NotContains(t, resolution, "<<<<<<< HEAD")
		assert.NotContains(t, resolution, "=======")
		assert.NotContains(t, resolution, ">>>>>>> ")
		
		// Apply resolution
		err = gitService.ResolveConflict(ctx, internalWorkDir, conflict.File, resolution)
		require.NoError(t, err)
	}
	
	// Generate and apply commit message
	changedFiles := getConflictFiles(conflicts)
	commitMessage, err := aiService.GenerateCommitMessage(ctx, changedFiles)
	require.NoError(t, err)
	
	err = gitService.Commit(ctx, internalWorkDir, commitMessage)
	require.NoError(t, err)
	
	// Phase 4: Run tests
	t.Log("=== Phase 4: Testing ===")
	
	// Run go mod tidy first
	testResult, err := testService.RunCommand(ctx, interfaces.TestCommand{
		Name:    scenario.Config.Tests.Commands[0].Name,
		Command: scenario.Config.Tests.Commands[0].Command,
		Args:    scenario.Config.Tests.Commands[0].Args,
		WorkingDir: internalWorkDir,
		Environment: scenario.Config.Tests.Commands[0].Environment,
		Timeout: scenario.Config.Tests.Timeout,
	})
	require.NoError(t, err)
	
	// Run build
	testResult, err = testService.RunCommand(ctx, interfaces.TestCommand{
		Name:    scenario.Config.Tests.Commands[1].Name,
		Command: scenario.Config.Tests.Commands[1].Command,
		Args:    scenario.Config.Tests.Commands[1].Args,
		WorkingDir: internalWorkDir,
		Environment: scenario.Config.Tests.Commands[1].Environment,
		Timeout: scenario.Config.Tests.Timeout,
	})
	require.NoError(t, err)
	if !testResult.Success {
		t.Logf("Build output: %s", testResult.Output)
		t.Logf("Build error: %s", testResult.Error)
	}
	assert.True(t, testResult.Success, "Build should succeed after conflict resolution")
	
	// Run tests
	testResult, err = testService.RunCommand(ctx, interfaces.TestCommand{
		Name:    scenario.Config.Tests.Commands[2].Name,
		Command: scenario.Config.Tests.Commands[2].Command,
		Args:    scenario.Config.Tests.Commands[2].Args,
		WorkingDir: internalWorkDir,
		Environment: scenario.Config.Tests.Commands[2].Environment,
		Timeout: scenario.Config.Tests.Timeout,
	})
	require.NoError(t, err)
	if !testResult.Success {
		t.Logf("Test output: %s", testResult.Output)
		t.Logf("Test error: %s", testResult.Error)
	}
	assert.True(t, testResult.Success, "Tests should pass after conflict resolution")
	
	// Phase 5: Create PR (mock)
	t.Log("=== Phase 5: PR Creation ===")
	prDescription, err := aiService.GeneratePRDescription(ctx, []string{commitMessage}, conflicts)
	require.NoError(t, err)
	
	// Mock PR creation
	pr, err := githubService.CreatePullRequest(ctx, interfaces.CreatePRRequest{
		Title: "AI-assisted rebase - " + time.Now().Format("2006-01-02"),
		Body:  prDescription,
		Head:  branchName,
		Base:  "main",
	})
	require.NoError(t, err)
	assert.NotNil(t, pr)
	
	// Phase 6: Send notification (mock)
	t.Log("=== Phase 6: Notifications ===")
	err = notifyService.SendMessage(ctx, interfaces.NotificationMessage{
		Title:   "AI Rebase Completed",
		Message: fmt.Sprintf("Successfully resolved %d conflicts and created PR #%d", len(conflicts), pr.Number),
		URL:     pr.HTMLURL,
		Level:   interfaces.NotificationLevelSuccess,
	})
	require.NoError(t, err)
	
	// Verify final state
	status, err := gitService.GetStatus(ctx, internalWorkDir)
	require.NoError(t, err)
	assert.True(t, status.IsClean, "Repository should be clean")
	
	t.Log("✅ Complete end-to-end workflow successful!")
}

// Error scenario tests

func testInvalidGitRepository(t *testing.T, tempDir string) {
	gitService := git.NewService()
	ctx := context.Background()
	
	err := gitService.Clone(ctx, "/non/existent/repo", filepath.Join(tempDir, "test"))
	assert.Error(t, err, "Should fail with invalid repository")
}

func testNetworkFailure(t *testing.T, tempDir string) {
	// This would simulate network failures
	// For now, we test with invalid URLs
	gitService := git.NewService()
	ctx := context.Background()
	
	err := gitService.Clone(ctx, "https://invalid-url-that-does-not-exist.com/repo.git", filepath.Join(tempDir, "test"))
	assert.Error(t, err, "Should fail with network error")
}

func testAIServiceFailure(t *testing.T, tempDir string) {
	// Test with invalid API key
	aiService := ai.NewService("openai", "invalid-key", "", "gpt-3.5-turbo", 1000)
	ctx := context.Background()
	
	conflict := interfaces.GitConflict{
		File:    "test.go",
		Content: "conflict content",
		Ours:    "our version",
		Theirs:  "their version",
	}
	
	_, err := aiService.ResolveConflict(ctx, conflict)
	assert.Error(t, err, "Should fail with invalid API key")
}

func testTestExecutionFailure(t *testing.T, tempDir string) {
	testService := test.NewService([]interfaces.TestCommand{})
	ctx := context.Background()
	
	// Test with invalid command
	result, err := testService.RunCommand(ctx, interfaces.TestCommand{
		Name:    "invalid-test",
		Command: "invalid-command-that-does-not-exist",
		Args:    []string{"arg1"},
	})
	
	require.NoError(t, err, "Should return result even if command fails")
	assert.False(t, result.Success, "Command should fail")
	// Exit code may be 0 in some cases, so we'll just check that the command failed
	assert.NotEmpty(t, result.Error, "Should have error message")
}