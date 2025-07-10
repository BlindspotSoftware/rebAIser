package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Store original environment variable
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("OPENAI_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()
	
	// Clear the environment variable for testing
	os.Unsetenv("OPENAI_API_KEY")

	// Create temporary config file
	configContent := `
interval: 4h
dry_run: true

git:
  internal_repo: "https://github.com/test/internal.git"
  upstream_repo: "https://github.com/test/upstream.git"
  working_dir: "/tmp/test"
  branch: "main"

ai:
  openai_api_key: "test-key"
  model: "gpt-4"
  max_tokens: 1000

github:
  token: "test-token"
  owner: "test-owner"
  repo: "test-repo"
  auto_merge_delay: 12h
  pr_template: "test-template.md"
  reviewers_team: "test-team"

slack:
  webhook_url: "https://hooks.slack.com/test"
  channel: "#test"
  username: "test-bot"

tests:
  timeout: 15m
  commands:
    - name: "test-build"
      command: "go"
      args: ["build"]
      working_dir: "/tmp/test"
      environment:
        TEST_ENV: "true"
`

	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	// Verify loaded values
	assert.Equal(t, 4*time.Hour, cfg.Interval)
	assert.True(t, cfg.DryRun)

	assert.Equal(t, "https://github.com/test/internal.git", cfg.Git.InternalRepo)
	assert.Equal(t, "https://github.com/test/upstream.git", cfg.Git.UpstreamRepo)
	assert.Equal(t, "/tmp/test", cfg.Git.WorkingDir)
	assert.Equal(t, "main", cfg.Git.Branch)

	assert.Equal(t, "test-key", cfg.AI.OpenAIAPIKey)
	assert.Equal(t, "gpt-4", cfg.AI.Model)
	assert.Equal(t, 1000, cfg.AI.MaxTokens)

	assert.Equal(t, "test-token", cfg.GitHub.Token)
	assert.Equal(t, "test-owner", cfg.GitHub.Owner)
	assert.Equal(t, "test-repo", cfg.GitHub.Repo)
	assert.Equal(t, 12*time.Hour, cfg.GitHub.AutoMergeDelay)
	assert.Equal(t, "test-template.md", cfg.GitHub.PRTemplate)
	assert.Equal(t, "test-team", cfg.GitHub.ReviewersTeam)

	assert.Equal(t, "https://hooks.slack.com/test", cfg.Slack.WebhookURL)
	assert.Equal(t, "#test", cfg.Slack.Channel)
	assert.Equal(t, "test-bot", cfg.Slack.Username)

	assert.Equal(t, 15*time.Minute, cfg.Tests.Timeout)
	require.Len(t, cfg.Tests.Commands, 1)
	assert.Equal(t, "test-build", cfg.Tests.Commands[0].Name)
	assert.Equal(t, "go", cfg.Tests.Commands[0].Command)
	assert.Equal(t, []string{"build"}, cfg.Tests.Commands[0].Args)
	assert.Equal(t, "/tmp/test", cfg.Tests.Commands[0].WorkingDir)
	assert.Equal(t, "true", cfg.Tests.Commands[0].Environment["TEST_ENV"])
}

func TestLoadConfig_WithDefaults(t *testing.T) {
	// Create minimal config file
	configContent := `
git:
  internal_repo: "https://github.com/test/internal.git"
  upstream_repo: "https://github.com/test/upstream.git"
  working_dir: "/tmp/test"
  branch: "main"

ai:
  openai_api_key: "test-key"

github:
  token: "test-token"
  owner: "test-owner"
  repo: "test-repo"

slack:
  webhook_url: "https://hooks.slack.com/test"
  channel: "#test"
  username: "test-bot"
`

	tmpFile, err := os.CreateTemp("", "config-defaults-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	// Verify defaults are applied
	assert.Equal(t, 8*time.Hour, cfg.Interval)
	assert.Equal(t, "gpt-4", cfg.AI.Model)
	assert.Equal(t, 2000, cfg.AI.MaxTokens)
	assert.Equal(t, 24*time.Hour, cfg.GitHub.AutoMergeDelay)
	assert.Equal(t, 30*time.Minute, cfg.Tests.Timeout)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create invalid YAML file
	configContent := `
invalid: yaml: content:
  - missing closing bracket
`

	tmpFile, err := os.CreateTemp("", "config-invalid-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, cfg)
}