package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BlindspotSoftware/rebAIser/internal/ai"
	"github.com/BlindspotSoftware/rebAIser/internal/git"
	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

// TestConflictScenarios tests different types of realistic conflicts
func TestConflictScenarios(t *testing.T) {
	// Skip if no OpenAI API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping conflict scenarios test - set OPENAI_API_KEY environment variable to run")
	}

	scenarios := []struct {
		name        string
		setupFunc   func(t *testing.T, upstreamDir, internalDir string)
		validateFunc func(t *testing.T, resolution string, conflict interfaces.GitConflict)
	}{
		{
			name:        "SimpleStringConflict",
			setupFunc:   setupSimpleStringConflict,
			validateFunc: validateSimpleStringResolution,
		},
		{
			name:        "FunctionSignatureConflict",
			setupFunc:   setupFunctionSignatureConflict,
			validateFunc: validateFunctionSignatureResolution,
		},
		{
			name:        "ImportConflict",
			setupFunc:   setupImportConflict,
			validateFunc: validateImportResolution,
		},
		{
			name:        "ConfigurationConflict",
			setupFunc:   setupConfigurationConflict,
			validateFunc: validateConfigurationResolution,
		},
		{
			name:        "CommentConflict",
			setupFunc:   setupCommentConflict,
			validateFunc: validateCommentResolution,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			testConflictScenario(t, scenario.setupFunc, scenario.validateFunc, apiKey)
		})
	}
}

func testConflictScenario(t *testing.T, setupFunc func(t *testing.T, upstream, internal string), validateFunc func(t *testing.T, resolution string, conflict interfaces.GitConflict), apiKey string) {
	tempDir := t.TempDir()
	
	upstreamDir := filepath.Join(tempDir, "upstream")
	internalDir := filepath.Join(tempDir, "internal")
	workDir := filepath.Join(tempDir, "work")
	
	// Setup base repositories
	createBaseRepo(t, upstreamDir, "upstream")
	createBaseRepo(t, internalDir, "internal")
	
	// Setup specific conflict scenario
	setupFunc(t, upstreamDir, internalDir)
	
	// Initialize services
	gitService := git.NewService()
	aiService := ai.NewService("openai", apiKey, "", "gpt-3.5-turbo", 1000)
	
	ctx := context.Background()
	
	// Clone internal repo to work directory
	internalWorkDir := filepath.Join(workDir, "internal")
	err := gitService.Clone(ctx, internalDir, internalWorkDir)
	require.NoError(t, err)
	
	// Add upstream as remote and fetch
	runGitCommand(t, internalWorkDir, "remote", "add", "upstream", upstreamDir)
	runGitCommand(t, internalWorkDir, "fetch", "upstream")
	
	// Create branch and attempt rebase
	err = gitService.CreateBranch(ctx, internalWorkDir, "test-branch")
	require.NoError(t, err)
	
	err = gitService.Rebase(ctx, internalWorkDir, "upstream/main")
	require.Error(t, err, "Should have conflicts")
	
	// Get and resolve conflicts
	conflicts, err := gitService.GetConflicts(ctx, internalWorkDir)
	require.NoError(t, err)
	require.Greater(t, len(conflicts), 0, "Should have conflicts")
	
	for _, conflict := range conflicts {
		t.Logf("Resolving conflict in: %s", conflict.File)
		
		// Use AI to resolve
		resolution, err := aiService.ResolveConflict(ctx, conflict)
		require.NoError(t, err)
		
		// Validate the resolution
		validateFunc(t, resolution, conflict)
		
		// Apply resolution
		err = gitService.ResolveConflict(ctx, internalWorkDir, conflict.File, resolution)
		require.NoError(t, err)
	}
	
	// Verify clean state
	status, err := gitService.GetStatus(ctx, internalWorkDir)
	require.NoError(t, err)
	assert.True(t, status.IsClean, "Should be clean after resolution")
}

// Scenario setup functions

func setupSimpleStringConflict(t *testing.T, upstreamDir, internalDir string) {
	// Upstream changes
	writeFile(t, filepath.Join(upstreamDir, "config.go"), `package main

const (
	AppName = "Awesome App"
	Version = "2.0.0"
	Author  = "Upstream Team"
)
`)
	runGitCommand(t, upstreamDir, "add", ".")
	runGitCommand(t, upstreamDir, "commit", "-m", "Update app info to 2.0.0")
	
	// Internal changes
	writeFile(t, filepath.Join(internalDir, "config.go"), `package main

const (
	AppName = "Internal App"
	Version = "1.5.0"
	Author  = "Internal Team"
)
`)
	runGitCommand(t, internalDir, "add", ".")
	runGitCommand(t, internalDir, "commit", "-m", "Update internal app info")
}

func setupFunctionSignatureConflict(t *testing.T, upstreamDir, internalDir string) {
	// Upstream changes - add parameter
	writeFile(t, filepath.Join(upstreamDir, "api.go"), `package main

import "fmt"

func ProcessRequest(id string, userId int, options map[string]string) error {
	fmt.Printf("Processing request %s for user %d with options %v\n", id, userId, options)
	return nil
}

func HandleError(err error, context string) {
	fmt.Printf("Error in %s: %v\n", context, err)
}
`)
	runGitCommand(t, upstreamDir, "add", ".")
	runGitCommand(t, upstreamDir, "commit", "-m", "Add options parameter to ProcessRequest")
	
	// Internal changes - add different parameter
	writeFile(t, filepath.Join(internalDir, "api.go"), `package main

import "fmt"

func ProcessRequest(id string, userId int, priority int) error {
	fmt.Printf("Processing request %s for user %d with priority %d\n", id, userId, priority)
	return nil
}

func HandleError(err error, context string) {
	fmt.Printf("Internal error in %s: %v\n", context, err)
}
`)
	runGitCommand(t, internalDir, "add", ".")
	runGitCommand(t, internalDir, "commit", "-m", "Add priority parameter to ProcessRequest")
}

func setupImportConflict(t *testing.T, upstreamDir, internalDir string) {
	// Upstream changes
	writeFile(t, filepath.Join(upstreamDir, "imports.go"), `package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Starting server...")
	log.Println("Server initialized")
}
`)
	runGitCommand(t, upstreamDir, "add", ".")
	runGitCommand(t, upstreamDir, "commit", "-m", "Add HTTP and time imports")
	
	// Internal changes
	writeFile(t, filepath.Join(internalDir, "imports.go"), `package main

import (
	"fmt"
	"log"
	"database/sql"
	"encoding/json"
)

func main() {
	fmt.Println("Starting internal server...")
	log.Println("Internal server initialized")
}
`)
	runGitCommand(t, internalDir, "add", ".")
	runGitCommand(t, internalDir, "commit", "-m", "Add database and JSON imports")
}

func setupConfigurationConflict(t *testing.T, upstreamDir, internalDir string) {
	// Upstream changes
	writeFile(t, filepath.Join(upstreamDir, "settings.json"), `{
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "upstream_db"
  },
  "redis": {
    "host": "localhost",
    "port": 6379
  },
  "logging": {
    "level": "info"
  }
}
`)
	runGitCommand(t, upstreamDir, "add", ".")
	runGitCommand(t, upstreamDir, "commit", "-m", "Add Redis configuration")
	
	// Internal changes
	writeFile(t, filepath.Join(internalDir, "settings.json"), `{
  "database": {
    "host": "internal-db.company.com",
    "port": 5432,
    "name": "internal_db"
  },
  "auth": {
    "provider": "internal-sso",
    "timeout": 30
  },
  "logging": {
    "level": "debug"
  }
}
`)
	runGitCommand(t, internalDir, "add", ".")
	runGitCommand(t, internalDir, "commit", "-m", "Add internal auth and logging config")
}

func setupCommentConflict(t *testing.T, upstreamDir, internalDir string) {
	// Upstream changes
	writeFile(t, filepath.Join(upstreamDir, "documented.go"), `package main

// Calculator provides mathematical operations
// Updated in version 2.0 with improved algorithms
type Calculator struct {
	precision int
}

// Add performs addition with high precision
// Returns the sum of two numbers
func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}
`)
	runGitCommand(t, upstreamDir, "add", ".")
	runGitCommand(t, upstreamDir, "commit", "-m", "Improve documentation")
	
	// Internal changes
	writeFile(t, filepath.Join(internalDir, "documented.go"), `package main

// Calculator provides mathematical operations
// Internal version with company-specific features
type Calculator struct {
	precision int
}

// Add performs addition with validation
// Internal implementation with error checking
func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}
`)
	runGitCommand(t, internalDir, "add", ".")
	runGitCommand(t, internalDir, "commit", "-m", "Add internal documentation")
}

// Validation functions

func validateSimpleStringResolution(t *testing.T, resolution string, conflict interfaces.GitConflict) {
	// Should contain package declaration
	assert.Contains(t, resolution, "package main")
	assert.Contains(t, resolution, "const (")
	
	// Should not contain conflict markers
	assert.NotContains(t, resolution, "<<<<<<< HEAD")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>> ")
	
	// Should be valid Go code structure
	assert.Contains(t, resolution, "AppName")
	assert.Contains(t, resolution, "Version")
	assert.Contains(t, resolution, "Author")
}

func validateFunctionSignatureResolution(t *testing.T, resolution string, conflict interfaces.GitConflict) {
	assert.Contains(t, resolution, "package main")
	assert.Contains(t, resolution, "func ProcessRequest(")
	assert.Contains(t, resolution, "func HandleError(")
	
	// Should not contain conflict markers
	assert.NotContains(t, resolution, "<<<<<<< HEAD")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>> ")
	
	// Should have reasonable function signature
	assert.Contains(t, resolution, "id string")
	assert.Contains(t, resolution, "userId int")
}

func validateImportResolution(t *testing.T, resolution string, conflict interfaces.GitConflict) {
	assert.Contains(t, resolution, "package main")
	assert.Contains(t, resolution, "import (")
	assert.Contains(t, resolution, "func main()")
	
	// Should not contain conflict markers
	assert.NotContains(t, resolution, "<<<<<<< HEAD")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>> ")
	
	// Should have basic imports
	assert.Contains(t, resolution, "\"fmt\"")
	assert.Contains(t, resolution, "\"log\"")
}

func validateConfigurationResolution(t *testing.T, resolution string, conflict interfaces.GitConflict) {
	// Should be valid JSON structure
	assert.Contains(t, resolution, "{")
	assert.Contains(t, resolution, "}")
	assert.Contains(t, resolution, "\"database\"")
	assert.Contains(t, resolution, "\"logging\"")
	
	// Should not contain conflict markers
	assert.NotContains(t, resolution, "<<<<<<< HEAD")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>> ")
}

func validateCommentResolution(t *testing.T, resolution string, conflict interfaces.GitConflict) {
	assert.Contains(t, resolution, "package main")
	assert.Contains(t, resolution, "// Calculator")
	assert.Contains(t, resolution, "type Calculator struct")
	assert.Contains(t, resolution, "func (c *Calculator) Add(")
	
	// Should not contain conflict markers
	assert.NotContains(t, resolution, "<<<<<<< HEAD")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>> ")
}

// Helper function to create base repository
func createBaseRepo(t *testing.T, dir, name string) {
	t.Helper()
	
	require.NoError(t, os.MkdirAll(dir, 0755))
	
	runGitCommand(t, dir, "init")
	runGitCommand(t, dir, "config", "user.name", "Test User")
	runGitCommand(t, dir, "config", "user.email", "test@example.com")
	
	// Create initial file
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println("Hello from `+name+`!")
}
`)
	
	runGitCommand(t, dir, "add", ".")
	runGitCommand(t, dir, "commit", "-m", "Initial commit")
}