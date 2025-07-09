package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type Service struct {
	log *logrus.Entry
}

func NewService() interfaces.GitService {
	return &Service{
		log: logrus.WithField("component", "git"),
	}
}

func (s *Service) Clone(ctx context.Context, repo, dir string) error {
	s.log.WithFields(logrus.Fields{
		"repo": repo,
		"dir":  dir,
	}).Info("Cloning repository")

	cmd := exec.CommandContext(ctx, "git", "clone", repo, dir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) Fetch(ctx context.Context, dir string) error {
	s.log.WithField("dir", dir).Info("Fetching updates")

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "fetch", "--all")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) Rebase(ctx context.Context, dir, branch string) error {
	s.log.WithFields(logrus.Fields{
		"dir":    dir,
		"branch": branch,
	}).Info("Starting rebase")

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rebase", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if it's a conflict (expected) or actual error
		if strings.Contains(string(output), "CONFLICT") || strings.Contains(err.Error(), "exit status 1") {
			return fmt.Errorf("rebase conflicts detected: %w\nOutput: %s", err, string(output))
		}
		return fmt.Errorf("failed to rebase: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) GetConflicts(ctx context.Context, dir string) ([]interfaces.GitConflict, error) {
	s.log.WithField("dir", dir).Info("Getting conflicts")

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict files: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	conflicts := make([]interfaces.GitConflict, 0, len(files))

	for _, file := range files {
		if file == "" {
			continue
		}

		conflict, err := s.getConflictContent(dir, file)
		if err != nil {
			s.log.WithError(err).WithField("file", file).Warn("Failed to get conflict content")
			continue
		}

		conflicts = append(conflicts, conflict)
	}

	return conflicts, nil
}

func (s *Service) getConflictContent(dir, file string) (interfaces.GitConflict, error) {
	filePath := fmt.Sprintf("%s/%s", dir, file)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return interfaces.GitConflict{}, fmt.Errorf("failed to read conflict file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var ours, theirs []string
	var inOurs, inTheirs bool

	for _, line := range lines {
		if strings.HasPrefix(line, "<<<<<<< ") {
			inOurs = true
			continue
		}
		if strings.HasPrefix(line, "======= ") {
			inOurs = false
			inTheirs = true
			continue
		}
		if strings.HasPrefix(line, ">>>>>>> ") {
			inTheirs = false
			continue
		}

		if inOurs {
			ours = append(ours, line)
		} else if inTheirs {
			theirs = append(theirs, line)
		}
	}

	return interfaces.GitConflict{
		File:    file,
		Content: string(content),
		Ours:    strings.Join(ours, "\n"),
		Theirs:  strings.Join(theirs, "\n"),
	}, nil
}

func (s *Service) ResolveConflict(ctx context.Context, dir, file, resolution string) error {
	s.log.WithField("file", file).Info("Resolving conflict")

	filePath := fmt.Sprintf("%s/%s", dir, file)
	err := os.WriteFile(filePath, []byte(resolution), 0644)
	if err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "add", file)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add resolved file: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) Commit(ctx context.Context, dir, message string) error {
	s.log.WithField("message", message).Info("Committing changes")

	// Configure git user if not already set
	if err := s.configureGitUser(ctx, dir); err != nil {
		return fmt.Errorf("failed to configure git user: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "commit", "-m", message)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) configureGitUser(ctx context.Context, dir string) error {
	// Check if user.name is already configured
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "config", "user.name")
	if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) != "" {
		return nil // Already configured
	}

	// Set user.name and user.email
	cmd = exec.CommandContext(ctx, "git", "-C", dir, "config", "user.name", "AI Rebaser")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set user.name: %w\nOutput: %s", err, string(output))
	}

	cmd = exec.CommandContext(ctx, "git", "-C", dir, "config", "user.email", "ai-rebaser@example.com")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set user.email: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) Push(ctx context.Context, dir, branch string) error {
	s.log.WithField("branch", branch).Info("Pushing changes")

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "push", "origin", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) CreateBranch(ctx context.Context, dir, branch string) error {
	s.log.WithField("branch", branch).Info("Creating branch")

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "checkout", "-b", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create branch: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) GetStatus(ctx context.Context, dir string) (interfaces.GitStatus, error) {
	s.log.WithField("dir", dir).Info("Getting git status")

	// Get porcelain status
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return interfaces.GitStatus{}, fmt.Errorf("failed to get status: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	gitStatus := interfaces.GitStatus{
		IsClean: len(lines) == 1 && lines[0] == "", // Empty output means clean
	}

	// Check for conflicts specifically
	cmd = exec.CommandContext(ctx, "git", "-C", dir, "diff", "--name-only", "--diff-filter=U")
	conflictOutput, err := cmd.Output()
	if err != nil {
		return interfaces.GitStatus{}, fmt.Errorf("failed to get conflict status: %w", err)
	}

	conflictFiles := strings.Split(strings.TrimSpace(string(conflictOutput)), "\n")
	if len(conflictFiles) > 0 && conflictFiles[0] != "" {
		gitStatus.HasConflicts = true
		gitStatus.ConflictFiles = conflictFiles
	}

	// Parse modified files from porcelain output
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) >= 3 {
			file := line[3:]
			// Check if it's a conflict file
			isConflict := false
			for _, conflictFile := range conflictFiles {
				if file == conflictFile {
					isConflict = true
					break
				}
			}
			if !isConflict {
				gitStatus.ModifiedFiles = append(gitStatus.ModifiedFiles, file)
			}
		}
	}

	return gitStatus, nil
}

// AddRemote adds a remote to the repository
func (s *Service) AddRemote(ctx context.Context, dir, name, url string) error {
	s.log.WithFields(logrus.Fields{
		"dir":  dir,
		"name": name,
		"url":  url,
	}).Info("Adding remote")

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "remote", "add", name, url)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if remote already exists
		if strings.Contains(string(output), "already exists") {
			s.log.WithField("name", name).Info("Remote already exists, skipping")
			return nil
		}
		return fmt.Errorf("failed to add remote: %w\nOutput: %s", err, string(output))
	}

	return nil
}