package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"

	"github.com/BlindspotSoftware/rebAIser/internal/ai"
	"github.com/BlindspotSoftware/rebAIser/internal/config"
	"github.com/BlindspotSoftware/rebAIser/internal/git"
	"github.com/BlindspotSoftware/rebAIser/internal/github"
	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
	"github.com/BlindspotSoftware/rebAIser/internal/notify"
	"github.com/BlindspotSoftware/rebAIser/internal/test"
	"strings"
)

var CLI struct {
	Config       string `short:"c" help:"Path to configuration file" default:"config.yaml"`
	LogLevel     string `short:"l" help:"Log level (debug, info, warn, error)" default:"info"`
	DryRun       bool   `short:"d" help:"Dry run mode - don't make actual changes"`
	RunOnce      bool   `short:"o" help:"Run once and exit (don't run periodically)"`
	KeepArtifacts bool   `short:"k" help:"Keep temporary working directory artifacts (don't cleanup)"`
	Version      bool   `short:"v" help:"Show version information"`
}

func main() {
	kong.Parse(&CLI)

	if CLI.Version {
		logrus.Info("AI Rebaser v1.0.0")
		return
	}

	// Setup structured logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	
	level, err := logrus.ParseLevel(CLI.LogLevel)
	if err != nil {
		logrus.WithError(err).Fatal("Invalid log level")
	}
	logrus.SetLevel(level)

	log := logrus.WithField("component", "main")
	log.Info("Starting AI Rebaser")

	// Load configuration
	cfg, err := config.LoadConfig(CLI.Config)
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	// Apply CLI overrides
	if CLI.DryRun {
		cfg.DryRun = true
	}
	cfg.KeepArtifacts = CLI.KeepArtifacts

	// Create context for graceful shutdown
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal")
		cancel()
	}()

	// Start the rebaser service
	if err := runRebaser(appCtx, cfg); err != nil {
		log.WithError(err).Fatal("Rebaser failed")
	}

	log.Info("AI Rebaser stopped")
}

func runRebaser(ctx context.Context, cfg *config.Config) error {
	log := logrus.WithField("component", "rebaser")
	
	// Initialize services
	services, err := initializeServices(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}
	
	// Run once if requested
	if CLI.RunOnce {
		log.Info("Running single rebase operation")
		return performRebase(ctx, cfg, services)
	}

	// Create ticker for periodic rebasing
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.WithField("interval", cfg.Interval).Info("Starting rebaser with configured interval")

	// Run initial rebase
	if err := performRebase(ctx, cfg, services); err != nil {
		log.WithError(err).Error("Initial rebase failed")
	}

	// Run periodic rebases
	for {
		select {
		case <-ctx.Done():
			log.Info("Shutting down rebaser")
			return nil
		case <-ticker.C:
			if err := performRebase(ctx, cfg, services); err != nil {
				log.WithError(err).Error("Periodic rebase failed")
			}
		}
	}
}

type Services struct {
	Git    interfaces.GitService
	AI     interfaces.AIService
	GitHub interfaces.GitHubService
	Notify interfaces.NotifyService
	Test   interfaces.TestService
}

func initializeServices(cfg *config.Config) (*Services, error) {
	log := logrus.WithField("component", "services")
	log.Info("Initializing services")

	// Convert config test commands to interface test commands
	testCommands := make([]interfaces.TestCommand, len(cfg.Tests.Commands))
	for i, cmd := range cfg.Tests.Commands {
		testCommands[i] = interfaces.TestCommand{
			Name:        cmd.Name,
			Command:     cmd.Command,
			Args:        cmd.Args,
			WorkingDir:  cmd.WorkingDir,
			Environment: cmd.Environment,
			Timeout:     cfg.Tests.Timeout, // Use global timeout from config
		}
	}

	services := &Services{
		Git:    git.NewService(),
		AI:     ai.NewService(cfg.AI.OpenAIAPIKey, cfg.AI.Model, cfg.AI.MaxTokens),
		GitHub: github.NewService(cfg.GitHub.Token, cfg.GitHub.Owner, cfg.GitHub.Repo),
		Notify: notify.NewService(cfg.Slack.WebhookURL, cfg.Slack.Channel, cfg.Slack.Username),
		Test:   test.NewService(testCommands),
	}

	log.Info("Services initialized successfully")
	return services, nil
}

func performRebase(ctx context.Context, cfg *config.Config, services *Services) error {
	log := logrus.WithField("component", "rebase")
	log.Info("Starting rebase operation")

	// Ensure cleanup runs regardless of success or failure
	defer func() {
		if err := cleanupWorkingDirectory(cfg); err != nil {
			log.WithError(err).Warn("Failed to cleanup working directory")
		}
	}()

	// Phase 1: Setup and Git Operations
	if err := setupWorkingDirectory(ctx, cfg, services); err != nil {
		sendErrorNotification(ctx, services, "AI Rebaser - Setup Failed", "Failed to setup working directory", err)
		return fmt.Errorf("setup failed: %w", err)
	}

	// Phase 2: Perform Rebase and Handle Conflicts
	branchName := fmt.Sprintf("ai-rebase-%d", time.Now().Unix())
	conflicts, err := performGitRebase(ctx, cfg, services, branchName)
	if err != nil {
		sendErrorNotification(ctx, services, "AI Rebaser - Git Rebase Failed", "Failed to perform git rebase", err)
		return fmt.Errorf("git rebase failed: %w", err)
	}

	// Phase 3: Resolve Conflicts with AI (if any)
	if len(conflicts) > 0 {
		if err := resolveConflictsWithAI(ctx, cfg, services, conflicts); err != nil {
			sendErrorNotification(ctx, services, "AI Rebaser - Conflict Resolution Failed", 
				fmt.Sprintf("Failed to resolve %d conflicts with AI", len(conflicts)), err)
			return fmt.Errorf("conflict resolution failed: %w", err)
		}
	}

	// Phase 4: Run Tests
	if err := runTests(ctx, cfg, services); err != nil {
		sendErrorNotification(ctx, services, "AI Rebaser - Tests Failed", "Tests failed after rebase", err)
		return fmt.Errorf("tests failed: %w", err)
	}

	// Phase 5: Create PR
	pr, err := createPullRequest(ctx, cfg, services, conflicts, branchName)
	if err != nil {
		sendErrorNotification(ctx, services, "AI Rebaser - PR Creation Failed", "Failed to create pull request", err)
		return fmt.Errorf("PR creation failed: %w", err)
	}

	// Phase 6: Send Notifications
	if err := sendNotifications(ctx, cfg, services, pr, conflicts); err != nil {
		log.WithError(err).Warn("Failed to send notifications")
	}

	log.Info("Rebase operation completed successfully")
	return nil
}

// Phase 1: Setup working directory and clone repositories
func setupWorkingDirectory(ctx context.Context, cfg *config.Config, services *Services) error {
	log := logrus.WithField("component", "setup")
	log.Info("Setting up working directory")

	// Create temporary directory with random name
	tempDir, err := os.MkdirTemp("", "ai-rebaser-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	
	// Store the actual working directory in config
	cfg.ActualWorkingDir = tempDir
	log.WithField("temp_dir", tempDir).Info("Created temporary working directory")

	// Clone internal repository
	internalDir := fmt.Sprintf("%s/internal", cfg.ActualWorkingDir)
	if err := services.Git.Clone(ctx, cfg.Git.InternalRepo, internalDir); err != nil {
		// If clone fails, try to fetch (repo might already exist)
		log.WithError(err).Info("Clone failed, attempting to fetch instead")
		if err := services.Git.Fetch(ctx, internalDir); err != nil {
			return fmt.Errorf("failed to clone or fetch internal repo: %w", err)
		}
	}

	// Add upstream remote and fetch
	if err := services.Git.AddRemote(ctx, internalDir, "upstream", cfg.Git.UpstreamRepo); err != nil {
		return fmt.Errorf("failed to add upstream remote: %w", err)
	}
	
	if err := services.Git.Fetch(ctx, internalDir); err != nil {
		return fmt.Errorf("failed to fetch from repositories: %w", err)
	}

	log.Info("Working directory setup completed")
	return nil
}

// Phase 2: Perform git rebase and detect conflicts
func performGitRebase(ctx context.Context, cfg *config.Config, services *Services, branchName string) ([]interfaces.GitConflict, error) {
	log := logrus.WithField("component", "git-rebase")
	log.Info("Starting git rebase operation")

	internalDir := fmt.Sprintf("%s/internal", cfg.ActualWorkingDir)
	
	// Create a new branch for the rebase
	if err := services.Git.CreateBranch(ctx, internalDir, branchName); err != nil {
		return nil, fmt.Errorf("failed to create rebase branch: %w", err)
	}

	// Attempt rebase against upstream
	upstreamBranch := fmt.Sprintf("upstream/%s", cfg.Git.Branch)
	err := services.Git.Rebase(ctx, internalDir, upstreamBranch)
	if err != nil {
		// Check if it's a conflict error (expected) or actual failure
		if !isConflictError(err) {
			return nil, fmt.Errorf("unexpected rebase error: %w", err)
		}
		log.WithError(err).Info("Rebase conflicts detected, proceeding with conflict resolution")
	}

	// Get conflicts if any
	conflicts, err := services.Git.GetConflicts(ctx, internalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicts: %w", err)
	}

	log.WithField("conflicts", len(conflicts)).Info("Git rebase completed")
	return conflicts, nil
}

// Phase 3: Resolve conflicts using AI
func resolveConflictsWithAI(ctx context.Context, cfg *config.Config, services *Services, conflicts []interfaces.GitConflict) error {
	log := logrus.WithField("component", "conflict-resolution")
	log.WithField("conflicts", len(conflicts)).Info("Resolving conflicts with AI")

	internalDir := fmt.Sprintf("%s/internal", cfg.ActualWorkingDir)

	for _, conflict := range conflicts {
		log.WithField("file", conflict.File).Info("Resolving conflict")
		
		// Use AI to resolve the conflict
		resolution, err := services.AI.ResolveConflict(ctx, conflict)
		if err != nil {
			return fmt.Errorf("AI failed to resolve conflict in %s: %w", conflict.File, err)
		}

		// Apply the resolution
		if err := services.Git.ResolveConflict(ctx, internalDir, conflict.File, resolution); err != nil {
			return fmt.Errorf("failed to apply resolution for %s: %w", conflict.File, err)
		}
	}

	// Generate commit message for the resolved conflicts
	changes := make([]string, len(conflicts))
	for i, conflict := range conflicts {
		changes[i] = conflict.File
	}
	
	commitMessage, err := services.AI.GenerateCommitMessage(ctx, changes)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Commit the resolved conflicts
	if err := services.Git.Commit(ctx, internalDir, commitMessage); err != nil {
		return fmt.Errorf("failed to commit resolved conflicts: %w", err)
	}

	log.Info("All conflicts resolved successfully")
	return nil
}

// Phase 4: Run tests to validate the rebase
func runTests(ctx context.Context, cfg *config.Config, services *Services) error {
	log := logrus.WithField("component", "testing")
	log.Info("Running tests")

	internalDir := fmt.Sprintf("%s/internal", cfg.ActualWorkingDir)
	
	// Run the test suite
	result, err := services.Test.RunTests(ctx, internalDir)
	if err != nil {
		return fmt.Errorf("failed to run tests: %w", err)
	}

	if !result.Success {
		log.WithField("failed_tests", result.FailedTests).Error("Tests failed")
		return fmt.Errorf("tests failed: %v", result.FailedTests)
	}

	log.WithField("duration", result.Duration).Info("All tests passed")
	return nil
}

// Phase 5: Create pull request
func createPullRequest(ctx context.Context, cfg *config.Config, services *Services, conflicts []interfaces.GitConflict, branchName string) (*interfaces.PullRequest, error) {
	log := logrus.WithField("component", "pr-creation")
	log.Info("Creating pull request")

	internalDir := fmt.Sprintf("%s/internal", cfg.ActualWorkingDir)

	// Push the branch to GitHub
	if err := services.Git.Push(ctx, internalDir, branchName); err != nil {
		return nil, fmt.Errorf("failed to push branch: %w", err)
	}

	// Generate PR description with AI
	commits := []string{} // TODO: Get actual commit messages
	prDescription, err := services.AI.GeneratePRDescription(ctx, commits, conflicts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PR description: %w", err)
	}

	// Create the PR
	prTitle := fmt.Sprintf("AI-assisted rebase - %s", time.Now().Format("2006-01-02"))
	prRequest := interfaces.CreatePRRequest{
		Title: prTitle,
		Body:  prDescription,
		Head:  branchName,
		Base:  cfg.Git.Branch,
		Draft: false,
	}

	pr, err := services.GitHub.CreatePullRequest(ctx, prRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	// Add reviewers if configured
	if cfg.GitHub.ReviewersTeam != "" {
		if err := services.GitHub.AddReviewers(ctx, pr.Number, []string{cfg.GitHub.ReviewersTeam}); err != nil {
			log.WithError(err).Warn("Failed to add reviewers")
		}
	}

	log.WithField("pr_number", pr.Number).Info("Pull request created successfully")
	return pr, nil
}

// Phase 6: Send notifications
func sendNotifications(ctx context.Context, cfg *config.Config, services *Services, pr *interfaces.PullRequest, conflicts []interfaces.GitConflict) error {
	log := logrus.WithField("component", "notifications")
	log.Info("Sending notifications")

	// Create detailed message based on conflicts
	var messageText string
	if len(conflicts) == 0 {
		messageText = fmt.Sprintf("✅ Rebase completed successfully with no conflicts. PR #%d created and ready for review.", pr.Number)
	} else {
		conflictFiles := make([]string, len(conflicts))
		for i, conflict := range conflicts {
			conflictFiles[i] = conflict.File
		}
		messageText = fmt.Sprintf("🤖 AI-assisted rebase completed! Resolved %d conflicts in files: %s. PR #%d created and ready for review.", 
			len(conflicts), 
			strings.Join(conflictFiles, ", "), 
			pr.Number)
	}

	message := interfaces.NotificationMessage{
		Title:   "AI Rebaser - Rebase Completed",
		Message: messageText,
		URL:     pr.HTMLURL,
		Level:   interfaces.NotificationLevelSuccess,
	}

	if err := services.Notify.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	log.Info("Notifications sent successfully")
	return nil
}

// Helper function to send error notifications
func sendErrorNotification(ctx context.Context, services *Services, title, message string, err error) {
	log := logrus.WithField("component", "notifications")
	
	notification := interfaces.NotificationMessage{
		Title:   title,
		Message: fmt.Sprintf("❌ %s\n\nError: %s", message, err.Error()),
		Level:   interfaces.NotificationLevelError,
	}
	
	if notifyErr := services.Notify.SendMessage(ctx, notification); notifyErr != nil {
		log.WithError(notifyErr).Error("Failed to send error notification")
	}
}

// Helper function to check if an error is a conflict error
func isConflictError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "conflict") || strings.Contains(err.Error(), "CONFLICT"))
}

// Cleanup working directory unless artifacts should be kept
func cleanupWorkingDirectory(cfg *config.Config) error {
	if cfg.KeepArtifacts {
		log := logrus.WithField("component", "cleanup")
		log.WithField("temp_dir", cfg.ActualWorkingDir).Info("Keeping artifacts, skipping cleanup")
		return nil
	}

	if cfg.ActualWorkingDir == "" {
		return nil // Nothing to cleanup
	}

	log := logrus.WithField("component", "cleanup")
	log.WithField("temp_dir", cfg.ActualWorkingDir).Info("Cleaning up temporary working directory")

	if err := os.RemoveAll(cfg.ActualWorkingDir); err != nil {
		return fmt.Errorf("failed to remove temporary directory: %w", err)
	}

	log.Info("Cleanup completed successfully")
	return nil
}