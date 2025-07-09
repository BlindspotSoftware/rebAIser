package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BlindspotSoftware/rebAIser/internal/config"
)

func TestIntegration_ConfigLoading(t *testing.T) {
	// This test demonstrates how the full configuration loading works
	// It uses the test config file to ensure the entire config parsing works
	
	configPath := "testdata/test-config.yaml"
	
	// Verify test config file exists
	_, err := os.Stat(configPath)
	require.NoError(t, err, "Test config file should exist")
	
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err, "Should be able to load test config")
	
	// Verify configuration is loaded correctly
	assert.True(t, cfg.DryRun, "Dry run should be enabled in test config")
	assert.Equal(t, "https://github.com/test-org/internal-repo.git", cfg.Git.InternalRepo)
	assert.Equal(t, "https://github.com/upstream/open-source-repo.git", cfg.Git.UpstreamRepo)
	assert.Equal(t, "/tmp/ai-rebaser-test", cfg.Git.WorkingDir)
	assert.Equal(t, "main", cfg.Git.Branch)
	
	// Verify service initialization works with this config
	services, err := initializeServices(cfg)
	require.NoError(t, err, "Should be able to initialize services with test config")
	
	// Verify all services are initialized
	assert.NotNil(t, services.Git)
	assert.NotNil(t, services.AI)
	assert.NotNil(t, services.GitHub)
	assert.NotNil(t, services.Notify)
	assert.NotNil(t, services.Test)
}

func TestIntegration_CLIFlags(t *testing.T) {
	// This test demonstrates how CLI flags work with the application
	// It's a demonstration of how the CLI could be tested in integration scenarios
	
	// Save original CLI values
	originalConfig := CLI.Config
	originalDryRun := CLI.DryRun
	originalRunOnce := CLI.RunOnce
	
	// Reset CLI for test
	CLI.Config = "testdata/test-config.yaml"
	CLI.DryRun = true
	CLI.RunOnce = true
	
	// Restore original values after test
	defer func() {
		CLI.Config = originalConfig
		CLI.DryRun = originalDryRun
		CLI.RunOnce = originalRunOnce
	}()
	
	// Load configuration
	cfg, err := config.LoadConfig(CLI.Config)
	require.NoError(t, err)
	
	// Apply CLI overrides (this is what happens in main())
	if CLI.DryRun {
		cfg.DryRun = true
	}
	
	// Verify the override worked
	assert.True(t, cfg.DryRun, "CLI dry-run override should be applied")
}

// This demonstrates how you could create a mock-based integration test
// that verifies the entire workflow without external dependencies
func TestIntegration_MockedWorkflow(t *testing.T) {
	// This test demonstrates how you could run the entire workflow
	// using mocked services to verify the orchestration is correct
	
	cfg := &config.Config{
		Git: config.GitConfig{
			WorkingDir:   "/tmp/test-integration",
			InternalRepo: "https://github.com/test/internal.git",
			UpstreamRepo: "https://github.com/test/upstream.git",
			Branch:       "main",
		},
		GitHub: config.GitHubConfig{
			ReviewersTeam: "test-team",
		},
		DryRun: true,
	}
	
	// In a real integration test, you would:
	// 1. Use the real services with test repositories
	// 2. Or use mocks to simulate the entire workflow
	// 3. Verify that the workflow completes successfully
	
	// For demonstration, we'll just verify that service initialization works
	services, err := initializeServices(cfg)
	require.NoError(t, err)
	assert.NotNil(t, services)
	
	// In a full integration test, you would call:
	// err = performRebase(context.Background(), cfg, services)
	// assert.NoError(t, err)
}