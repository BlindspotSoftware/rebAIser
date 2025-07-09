package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/BlindspotSoftware/rebAIser/internal/config"
	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
	"github.com/BlindspotSoftware/rebAIser/internal/mocks"
)

func TestInitializeServices(t *testing.T) {
	cfg := &config.Config{
		AI: config.AIConfig{
			OpenAIAPIKey: "test-key",
			Model:        "gpt-4",
			MaxTokens:    2000,
		},
		GitHub: config.GitHubConfig{
			Token: "test-token",
			Owner: "test-owner",
			Repo:  "test-repo",
		},
		Slack: config.SlackConfig{
			WebhookURL: "https://hooks.slack.com/test",
			Channel:    "#test",
			Username:   "test-bot",
		},
	}

	services, err := initializeServices(cfg)
	require.NoError(t, err)
	require.NotNil(t, services)
	assert.NotNil(t, services.Git)
	assert.NotNil(t, services.AI)
	assert.NotNil(t, services.GitHub)
	assert.NotNil(t, services.Notify)
	assert.NotNil(t, services.Test)
}

func TestPerformRebase_Success(t *testing.T) {
	// Setup mocks
	mockGit := &mocks.MockGitService{}
	mockAI := &mocks.MockAIService{}
	mockGitHub := &mocks.MockGitHubService{}
	mockNotify := &mocks.MockNotifyService{}
	mockTest := &mocks.MockTestService{}

	services := &Services{
		Git:    mockGit,
		AI:     mockAI,
		GitHub: mockGitHub,
		Notify: mockNotify,
		Test:   mockTest,
	}

	cfg := &config.Config{
		Git: config.GitConfig{
			WorkingDir:   "/tmp/test",
			InternalRepo: "https://github.com/test/internal.git",
			UpstreamRepo: "https://github.com/test/upstream.git",
			Branch:       "main",
		},
		GitHub: config.GitHubConfig{
			ReviewersTeam: "core-team",
		},
	}

	ctx := context.Background()

	// Mock setup expectations
	mockGit.On("Clone", ctx, cfg.Git.InternalRepo, mock.AnythingOfType("string")).Return(nil)
	mockGit.On("AddRemote", ctx, mock.AnythingOfType("string"), "upstream", cfg.Git.UpstreamRepo).Return(nil)
	mockGit.On("Fetch", ctx, mock.AnythingOfType("string")).Return(nil)
	mockGit.On("CreateBranch", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockGit.On("Rebase", ctx, mock.AnythingOfType("string"), "upstream/main").Return(nil)
	mockGit.On("GetConflicts", ctx, mock.AnythingOfType("string")).Return([]interfaces.GitConflict{}, nil)

	// Mock test expectations
	testResult := &interfaces.TestResult{
		Success:  true,
		Duration: 30 * time.Second,
		Results:  []interfaces.CommandResult{},
	}
	mockTest.On("RunTests", ctx, mock.AnythingOfType("string")).Return(testResult, nil)

	// Mock GitHub expectations
	mockGit.On("Push", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockAI.On("GeneratePRDescription", ctx, []string{}, []interfaces.GitConflict{}).Return("Test PR description", nil)
	
	pr := &interfaces.PullRequest{
		Number:  123,
		HTMLURL: "https://github.com/test/internal/pull/123",
	}
	mockGitHub.On("CreatePullRequest", ctx, mock.AnythingOfType("interfaces.CreatePRRequest")).Return(pr, nil)
	mockGitHub.On("AddReviewers", ctx, 123, []string{"core-team"}).Return(nil)

	// Mock notification expectations
	mockNotify.On("SendMessage", ctx, mock.AnythingOfType("interfaces.NotificationMessage")).Return(nil)

	// Execute
	err := performRebase(ctx, cfg, services)

	// Assert
	assert.NoError(t, err)
	mockGit.AssertExpectations(t)
	mockAI.AssertExpectations(t)
	mockGitHub.AssertExpectations(t)
	mockNotify.AssertExpectations(t)
	mockTest.AssertExpectations(t)
}

func TestPerformRebase_WithConflicts(t *testing.T) {
	// Setup mocks
	mockGit := &mocks.MockGitService{}
	mockAI := &mocks.MockAIService{}
	mockGitHub := &mocks.MockGitHubService{}
	mockNotify := &mocks.MockNotifyService{}
	mockTest := &mocks.MockTestService{}

	services := &Services{
		Git:    mockGit,
		AI:     mockAI,
		GitHub: mockGitHub,
		Notify: mockNotify,
		Test:   mockTest,
	}

	cfg := &config.Config{
		Git: config.GitConfig{
			WorkingDir:   "/tmp/test",
			InternalRepo: "https://github.com/test/internal.git",
			UpstreamRepo: "https://github.com/test/upstream.git",
			Branch:       "main",
		},
		GitHub: config.GitHubConfig{
			ReviewersTeam: "core-team",
		},
	}

	ctx := context.Background()

	// Test conflicts
	conflicts := []interfaces.GitConflict{
		{
			File:    "test.go",
			Content: "conflict content",
			Ours:    "our version",
			Theirs:  "their version",
		},
	}

	// Mock setup expectations
	mockGit.On("Clone", ctx, cfg.Git.InternalRepo, mock.AnythingOfType("string")).Return(nil)
	mockGit.On("AddRemote", ctx, mock.AnythingOfType("string"), "upstream", cfg.Git.UpstreamRepo).Return(nil)
	mockGit.On("Fetch", ctx, mock.AnythingOfType("string")).Return(nil)
	mockGit.On("CreateBranch", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockGit.On("Rebase", ctx, mock.AnythingOfType("string"), "upstream/main").Return(errors.New("rebase conflicts detected"))
	mockGit.On("GetConflicts", ctx, mock.AnythingOfType("string")).Return(conflicts, nil)

	// Mock AI conflict resolution
	mockAI.On("ResolveConflict", ctx, conflicts[0]).Return("resolved content", nil)
	mockGit.On("ResolveConflict", ctx, mock.AnythingOfType("string"), "test.go", "resolved content").Return(nil)
	mockAI.On("GenerateCommitMessage", ctx, []string{"test.go"}).Return("AI: Resolve conflicts in test.go", nil)
	mockGit.On("Commit", ctx, mock.AnythingOfType("string"), "AI: Resolve conflicts in test.go").Return(nil)

	// Mock test expectations
	testResult := &interfaces.TestResult{
		Success:  true,
		Duration: 30 * time.Second,
		Results:  []interfaces.CommandResult{},
	}
	mockTest.On("RunTests", ctx, mock.AnythingOfType("string")).Return(testResult, nil)

	// Mock GitHub expectations
	mockGit.On("Push", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockAI.On("GeneratePRDescription", ctx, []string{}, conflicts).Return("Test PR description with conflicts", nil)
	
	pr := &interfaces.PullRequest{
		Number:  124,
		HTMLURL: "https://github.com/test/internal/pull/124",
	}
	mockGitHub.On("CreatePullRequest", ctx, mock.AnythingOfType("interfaces.CreatePRRequest")).Return(pr, nil)
	mockGitHub.On("AddReviewers", ctx, 124, []string{"core-team"}).Return(nil)

	// Mock notification expectations
	mockNotify.On("SendMessage", ctx, mock.AnythingOfType("interfaces.NotificationMessage")).Return(nil)

	// Execute
	err := performRebase(ctx, cfg, services)

	// Assert
	assert.NoError(t, err)
	mockGit.AssertExpectations(t)
	mockAI.AssertExpectations(t)
	mockGitHub.AssertExpectations(t)
	mockNotify.AssertExpectations(t)
	mockTest.AssertExpectations(t)
}

func TestPerformRebase_TestFailure(t *testing.T) {
	// Setup mocks
	mockGit := &mocks.MockGitService{}
	mockAI := &mocks.MockAIService{}
	mockGitHub := &mocks.MockGitHubService{}
	mockNotify := &mocks.MockNotifyService{}
	mockTest := &mocks.MockTestService{}

	services := &Services{
		Git:    mockGit,
		AI:     mockAI,
		GitHub: mockGitHub,
		Notify: mockNotify,
		Test:   mockTest,
	}

	cfg := &config.Config{
		Git: config.GitConfig{
			WorkingDir:   "/tmp/test",
			InternalRepo: "https://github.com/test/internal.git",
			UpstreamRepo: "https://github.com/test/upstream.git",
			Branch:       "main",
		},
	}

	ctx := context.Background()

	// Mock setup expectations
	mockGit.On("Clone", ctx, cfg.Git.InternalRepo, mock.AnythingOfType("string")).Return(nil)
	mockGit.On("AddRemote", ctx, mock.AnythingOfType("string"), "upstream", cfg.Git.UpstreamRepo).Return(nil)
	mockGit.On("Fetch", ctx, mock.AnythingOfType("string")).Return(nil)
	mockGit.On("CreateBranch", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockGit.On("Rebase", ctx, mock.AnythingOfType("string"), "upstream/main").Return(nil)
	mockGit.On("GetConflicts", ctx, mock.AnythingOfType("string")).Return([]interfaces.GitConflict{}, nil)

	// Mock test failure
	testResult := &interfaces.TestResult{
		Success:     false,
		Duration:    30 * time.Second,
		Results:     []interfaces.CommandResult{},
		FailedTests: []string{"build", "test"},
	}
	mockTest.On("RunTests", ctx, mock.AnythingOfType("string")).Return(testResult, nil)

	// Mock notification for test failure
	mockNotify.On("SendMessage", ctx, mock.AnythingOfType("interfaces.NotificationMessage")).Return(nil)

	// Execute
	err := performRebase(ctx, cfg, services)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tests failed")
	mockGit.AssertExpectations(t)
	mockTest.AssertExpectations(t)
	mockNotify.AssertExpectations(t)
}

func TestSetupWorkingDirectory(t *testing.T) {
	// Setup mocks
	mockGit := &mocks.MockGitService{}
	services := &Services{Git: mockGit}

	cfg := &config.Config{
		Git: config.GitConfig{
			WorkingDir:   "/tmp/test-setup",
			InternalRepo: "https://github.com/test/internal.git",
			UpstreamRepo: "https://github.com/test/upstream.git",
		},
	}

	ctx := context.Background()

	// Mock expectations
	mockGit.On("Clone", ctx, cfg.Git.InternalRepo, mock.AnythingOfType("string")).Return(nil)
	mockGit.On("AddRemote", ctx, mock.AnythingOfType("string"), "upstream", cfg.Git.UpstreamRepo).Return(nil)
	mockGit.On("Fetch", ctx, mock.AnythingOfType("string")).Return(nil)

	// Execute
	err := setupWorkingDirectory(ctx, cfg, services)

	// Assert
	assert.NoError(t, err)
	mockGit.AssertExpectations(t)

	// Cleanup
	os.RemoveAll("/tmp/test-setup")
}

func TestSetupWorkingDirectory_CloneFallsBackToFetch(t *testing.T) {
	// Setup mocks
	mockGit := &mocks.MockGitService{}
	services := &Services{Git: mockGit}

	cfg := &config.Config{
		Git: config.GitConfig{
			WorkingDir:   "/tmp/test-setup-fallback",
			InternalRepo: "https://github.com/test/internal.git",
			UpstreamRepo: "https://github.com/test/upstream.git",
		},
	}

	ctx := context.Background()

	// Mock expectations - clone fails, fetch succeeds
	mockGit.On("Clone", ctx, cfg.Git.InternalRepo, mock.AnythingOfType("string")).Return(errors.New("clone failed"))
	mockGit.On("AddRemote", ctx, mock.AnythingOfType("string"), "upstream", cfg.Git.UpstreamRepo).Return(nil)
	mockGit.On("Fetch", ctx, mock.AnythingOfType("string")).Return(nil)

	// Execute
	err := setupWorkingDirectory(ctx, cfg, services)

	// Assert
	assert.NoError(t, err)
	mockGit.AssertExpectations(t)

	// Cleanup
	os.RemoveAll("/tmp/test-setup-fallback")
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"conflict error", errors.New("rebase conflicts detected"), true},
		{"CONFLICT error", errors.New("CONFLICT (content): merge failed"), true},
		{"other error", errors.New("network timeout"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isConflictError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}