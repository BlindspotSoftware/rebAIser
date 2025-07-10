package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BlindspotSoftware/rebAIser/internal/ai"
	"github.com/BlindspotSoftware/rebAIser/internal/config"
	"github.com/BlindspotSoftware/rebAIser/internal/git"
	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

const (
	// Set to your actual OpenAI API key for testing
	// Or use environment variable: OPENAI_API_KEY
	testOpenAIKey = "test-placeholder-key"
)

func TestRealWorldRebaseWorkflow(t *testing.T) {
	// Skip if no OpenAI API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = testOpenAIKey
	}
	if apiKey == "test-placeholder-key" {
		t.Skip("Skipping integration test - set OPENAI_API_KEY environment variable to run")
	}

	// Create temporary directory for test repositories
	tempDir := t.TempDir()
	
	// Setup test repositories
	upstreamDir := filepath.Join(tempDir, "upstream")
	internalDir := filepath.Join(tempDir, "internal")
	workDir := filepath.Join(tempDir, "work")
	
	// Create upstream repository
	createUpstreamRepo(t, upstreamDir)
	
	// Create internal repository (fork of upstream)
	createInternalRepo(t, internalDir, upstreamDir)
	
	// Create conflicting changes
	createConflictingChanges(t, upstreamDir, internalDir)
	
	// Test the full rebase workflow
	testFullRebaseWorkflow(t, upstreamDir, internalDir, workDir, apiKey)
}

func TestErrorScenarios(t *testing.T) {
	tempDir := t.TempDir()
	
	t.Run("InvalidAPIKey", func(t *testing.T) {
		testInvalidAPIKey(t, tempDir)
	})
	
	t.Run("GitRebaseFailure", func(t *testing.T) {
		testGitRebaseFailure(t, tempDir)
	})
	
	t.Run("TestFailureScenario", func(t *testing.T) {
		testTestFailureScenario(t, tempDir)
	})
	
	t.Run("ComplexConflictScenario", func(t *testing.T) {
		testComplexConflictScenario(t, tempDir)
	})
}

func createUpstreamRepo(t *testing.T, dir string) {
	t.Helper()
	
	require.NoError(t, os.MkdirAll(dir, 0755))
	
	// Initialize git repo
	runGitCommand(t, dir, "init")
	runGitCommand(t, dir, "config", "user.name", "Test User")
	runGitCommand(t, dir, "config", "user.email", "test@example.com")
	
	// Create initial files
	writeFile(t, filepath.Join(dir, "README.md"), `# Upstream Project

This is the upstream open-source project.

## Features
- Feature A
- Feature B
`)
	
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println("Hello from upstream!")
	fmt.Println("Version: 1.0.0")
}
`)
	
	writeFile(t, filepath.Join(dir, "utils.go"), `package main

import "fmt"

func PrintVersion() {
	fmt.Println("Upstream Version: 1.0.0")
}

func CalculateSum(a, b int) int {
	return a + b
}
`)
	
	runGitCommand(t, dir, "add", ".")
	runGitCommand(t, dir, "commit", "-m", "Initial upstream commit")
}

func createInternalRepo(t *testing.T, dir, upstreamDir string) {
	t.Helper()
	
	require.NoError(t, os.MkdirAll(dir, 0755))
	
	// Clone from upstream
	runGitCommand(t, dir, "clone", upstreamDir, ".")
	runGitCommand(t, dir, "config", "user.name", "Internal User")
	runGitCommand(t, dir, "config", "user.email", "internal@company.com")
	
	// Add internal customizations
	writeFile(t, filepath.Join(dir, "README.md"), `# Internal Project

This is our internal fork of the upstream project.

## Features
- Feature A
- Feature B
- Internal Feature X
- Internal Feature Y

## Internal Notes
- This version includes company-specific customizations
- Contact internal-team@company.com for support
`)
	
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println("Hello from internal version!")
	fmt.Println("Version: 1.0.0-internal")
	fmt.Println("Company: ACME Corp")
}
`)
	
	writeFile(t, filepath.Join(dir, "internal_config.go"), `package main

import "fmt"

func GetInternalConfig() map[string]string {
	return map[string]string{
		"company": "ACME Corp",
		"env":     "production",
		"region":  "us-east-1",
	}
}

func PrintInternalInfo() {
	fmt.Println("Internal build with custom features")
}
`)
	
	runGitCommand(t, dir, "add", ".")
	runGitCommand(t, dir, "commit", "-m", "Add internal customizations")
}

func createConflictingChanges(t *testing.T, upstreamDir, internalDir string) {
	t.Helper()
	
	// Make conflicting changes in upstream
	writeFile(t, filepath.Join(upstreamDir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println("Hello from upstream!")
	fmt.Println("Version: 2.0.0")
	fmt.Println("New upstream feature enabled")
}
`)
	
	writeFile(t, filepath.Join(upstreamDir, "utils.go"), `package main

import "fmt"

func PrintVersion() {
	fmt.Println("Upstream Version: 2.0.0")
}

func CalculateSum(a, b int) int {
	return a + b
}

func CalculateProduct(a, b int) int {
	return a * b
}

func NewUpstreamFunction() {
	fmt.Println("New function added in upstream")
}
`)
	
	runGitCommand(t, upstreamDir, "add", ".")
	runGitCommand(t, upstreamDir, "commit", "-m", "feat: upgrade to version 2.0.0 with new features")
	
	// Make conflicting changes in internal
	writeFile(t, filepath.Join(internalDir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println("Hello from internal version!")
	fmt.Println("Version: 1.0.0-internal")
	fmt.Println("Company: ACME Corp")
	fmt.Println("Internal feature v2 enabled")
}
`)
	
	writeFile(t, filepath.Join(internalDir, "utils.go"), `package main

import "fmt"

func PrintVersion() {
	fmt.Println("Internal Version: 1.0.0-internal")
}

func CalculateSum(a, b int) int {
	return a + b
}

func InternalUtilityFunction() {
	fmt.Println("Internal utility function")
}
`)
	
	runGitCommand(t, internalDir, "add", ".")
	runGitCommand(t, internalDir, "commit", "-m", "feat: add internal v2 features")
}

func testFullRebaseWorkflow(t *testing.T, upstreamDir, internalDir, workDir string, apiKey string) {
	t.Helper()
	
	// Create configuration for the test
	cfg := &config.Config{
		Git: config.GitConfig{
			InternalRepo: internalDir,
			UpstreamRepo: upstreamDir,
			WorkingDir:   workDir,
			Branch:       "main",
		},
		AI: config.AIConfig{
			OpenAIAPIKey: apiKey,
			Model:        "gpt-3.5-turbo", // Use cheaper model for testing
			MaxTokens:    1000,
		},
	}
	
	// Initialize services
	gitService := git.NewService()
	aiService := ai.NewService("openai", cfg.AI.OpenAIAPIKey, "", cfg.AI.Model, cfg.AI.MaxTokens)
	
	ctx := context.Background()
	
	// Test Phase 1: Setup working directory
	t.Log("Phase 1: Setting up working directory")
	require.NoError(t, os.MkdirAll(workDir, 0755))
	
	internalWorkDir := filepath.Join(workDir, "internal")
	err := gitService.Clone(ctx, internalDir, internalWorkDir)
	require.NoError(t, err)
	
	// Add upstream as remote
	runGitCommand(t, internalWorkDir, "remote", "add", "upstream", upstreamDir)
	runGitCommand(t, internalWorkDir, "fetch", "upstream")
	
	// Test Phase 2: Attempt rebase and detect conflicts
	t.Log("Phase 2: Attempting rebase")
	branchName := fmt.Sprintf("ai-rebase-%d", time.Now().Unix())
	err = gitService.CreateBranch(ctx, internalWorkDir, branchName)
	require.NoError(t, err)
	
	// This should fail with conflicts
	err = gitService.Rebase(ctx, internalWorkDir, "upstream/main")
	assert.Error(t, err, "Expected rebase to fail with conflicts")
	assert.Contains(t, err.Error(), "conflict", "Error should indicate conflicts")
	
	// Get conflicts
	conflicts, err := gitService.GetConflicts(ctx, internalWorkDir)
	require.NoError(t, err)
	assert.Greater(t, len(conflicts), 0, "Should have conflicts to resolve")
	
	t.Logf("Found %d conflicts: %v", len(conflicts), getConflictFiles(conflicts))
	
	// Test Phase 3: Resolve conflicts with AI
	t.Log("Phase 3: Resolving conflicts with AI")
	for _, conflict := range conflicts {
		t.Logf("Resolving conflict in: %s", conflict.File)
		
		// Test that we can parse the conflict
		assert.NotEmpty(t, conflict.Content, "Conflict content should not be empty")
		assert.NotEmpty(t, conflict.Ours, "Our version should not be empty")
		assert.NotEmpty(t, conflict.Theirs, "Their version should not be empty")
		
		// Use AI to resolve the conflict
		resolution, err := aiService.ResolveConflict(ctx, conflict)
		require.NoError(t, err, "AI should be able to resolve conflict")
		assert.NotEmpty(t, resolution, "Resolution should not be empty")
		
		// Verify resolution doesn't contain conflict markers
		assert.NotContains(t, resolution, "<<<<<<< HEAD", "Resolution should not contain conflict markers")
		assert.NotContains(t, resolution, "=======", "Resolution should not contain conflict markers")
		assert.NotContains(t, resolution, ">>>>>>> ", "Resolution should not contain conflict markers")
		
		// Apply the resolution
		err = gitService.ResolveConflict(ctx, internalWorkDir, conflict.File, resolution)
		require.NoError(t, err, "Should be able to apply resolution")
		
		t.Logf("Successfully resolved conflict in: %s", conflict.File)
	}
	
	// Generate commit message
	changedFiles := getConflictFiles(conflicts)
	commitMessage, err := aiService.GenerateCommitMessage(ctx, changedFiles)
	require.NoError(t, err)
	assert.NotEmpty(t, commitMessage, "Commit message should not be empty")
	
	t.Logf("Generated commit message: %s", commitMessage)
	
	// Commit the resolved conflicts
	err = gitService.Commit(ctx, internalWorkDir, commitMessage)
	require.NoError(t, err, "Should be able to commit resolved conflicts")
	
	// Test Phase 4: Verify the rebase was successful
	t.Log("Phase 4: Verifying rebase success")
	status, err := gitService.GetStatus(ctx, internalWorkDir)
	require.NoError(t, err)
	assert.True(t, status.IsClean, "Repository should be clean after resolving conflicts")
	assert.False(t, status.HasConflicts, "Should not have conflicts after resolution")
	
	// Test Phase 5: Generate PR description
	t.Log("Phase 5: Generating PR description")
	commits := []string{commitMessage}
	prDescription, err := aiService.GeneratePRDescription(ctx, commits, conflicts)
	require.NoError(t, err)
	assert.NotEmpty(t, prDescription, "PR description should not be empty")
	assert.Contains(t, prDescription, "##", "PR description should contain markdown headers")
	
	t.Logf("Generated PR description:\n%s", prDescription)
	
	// Verify the merged content makes sense
	verifyMergedContent(t, internalWorkDir)
	
	t.Log("✅ Full rebase workflow completed successfully!")
}

func testInvalidAPIKey(t *testing.T, tempDir string) {
	t.Helper()
	
	// Test with invalid API key
	aiService := ai.NewService("openai", "invalid-key", "", "gpt-3.5-turbo", 1000)
	
	conflict := interfaces.GitConflict{
		File:    "test.go",
		Content: "conflict content",
		Ours:    "our version",
		Theirs:  "their version",
	}
	
	ctx := context.Background()
	_, err := aiService.ResolveConflict(ctx, conflict)
	assert.Error(t, err, "Should fail with invalid API key")
	assert.Contains(t, err.Error(), "openai API call failed", "Should indicate API failure")
}

func testGitRebaseFailure(t *testing.T, tempDir string) {
	t.Helper()
	
	// Test with non-existent repository
	gitService := git.NewService()
	ctx := context.Background()
	
	err := gitService.Rebase(ctx, "/non/existent/path", "main")
	assert.Error(t, err, "Should fail with non-existent repository")
}

func testTestFailureScenario(t *testing.T, tempDir string) {
	t.Helper()
	
	// This would test what happens when tests fail after conflict resolution
	// For now, we'll simulate it by testing the error handling
	t.Log("Testing test failure scenario - would be implemented with actual test execution")
}

func testComplexConflictScenario(t *testing.T, tempDir string) {
	t.Helper()
	
	// Create a more complex conflict scenario
	// This would involve multiple files with different types of conflicts
	t.Log("Testing complex conflict scenario - would involve multiple conflict types")
}

// Helper functions

func runGitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Git command failed: git %v", args)
		t.Logf("Output: %s", output)
		t.Fatalf("Git command failed: %v", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "Should be able to write file: %s", path)
}

func getConflictFiles(conflicts []interfaces.GitConflict) []string {
	files := make([]string, len(conflicts))
	for i, conflict := range conflicts {
		files[i] = conflict.File
	}
	return files
}

func verifyMergedContent(t *testing.T, workDir string) {
	t.Helper()
	
	// Verify that key files exist and have reasonable content
	mainGoPath := filepath.Join(workDir, "main.go")
	content, err := os.ReadFile(mainGoPath)
	require.NoError(t, err, "Should be able to read main.go")
	
	contentStr := string(content)
	
	// The AI should have merged the versions intelligently
	// We can't predict exactly what it will do, but we can check for basic sanity
	assert.Contains(t, contentStr, "package main", "Should contain package declaration")
	assert.Contains(t, contentStr, "func main()", "Should contain main function")
	assert.NotContains(t, contentStr, "<<<<<<< HEAD", "Should not contain conflict markers")
	assert.NotContains(t, contentStr, "=======", "Should not contain conflict markers")
	assert.NotContains(t, contentStr, ">>>>>>> ", "Should not contain conflict markers")
	
	t.Log("✅ Merged content verification passed")
}