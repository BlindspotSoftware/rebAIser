package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"

	"github.com/9elements/rebaiser/internal/interfaces"
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

	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

func (s *Service) Fetch(ctx context.Context, dir string) error {
	s.log.WithField("dir", dir).Info("Fetching updates")

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*"},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

func (s *Service) Rebase(ctx context.Context, dir, branch string) error {
	s.log.WithFields(logrus.Fields{
		"dir":    dir,
		"branch": branch,
	}).Info("Starting rebase")

	// Use git command for rebase since go-git doesn't support it natively
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rebase", branch)
	if err := cmd.Run(); err != nil {
		// Check if it's a conflict (expected) or actual error
		if strings.Contains(err.Error(), "exit status 1") {
			return fmt.Errorf("rebase conflicts detected: %w", err)
		}
		return fmt.Errorf("failed to rebase: %w", err)
	}

	return nil
}

func (s *Service) GetConflicts(ctx context.Context, dir string) ([]interfaces.GitConflict, error) {
	s.log.WithField("dir", dir).Info("Getting conflicts")

	// Use git command to get conflict files
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

	// Use git command to add resolved file
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "add", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add resolved file: %w", err)
	}

	return nil
}

func (s *Service) Commit(ctx context.Context, dir, message string) error {
	s.log.WithField("message", message).Info("Committing changes")

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "AI Rebaser",
			Email: "ai-rebaser@example.com",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func (s *Service) Push(ctx context.Context, dir, branch string) error {
	s.log.WithField("branch", branch).Info("Pushing changes")

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))},
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

func (s *Service) CreateBranch(ctx context.Context, dir, branch string) error {
	s.log.WithField("branch", branch).Info("Creating branch")

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		Create: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

func (s *Service) GetStatus(ctx context.Context, dir string) (interfaces.GitStatus, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return interfaces.GitStatus{}, fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return interfaces.GitStatus{}, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return interfaces.GitStatus{}, fmt.Errorf("failed to get status: %w", err)
	}

	gitStatus := interfaces.GitStatus{
		IsClean: status.IsClean(),
	}

	for file, fileStatus := range status {
		if fileStatus.Staging == git.Unmerged || fileStatus.Worktree == git.Unmerged {
			gitStatus.HasConflicts = true
			gitStatus.ConflictFiles = append(gitStatus.ConflictFiles, file)
		} else {
			gitStatus.ModifiedFiles = append(gitStatus.ModifiedFiles, file)
		}
	}

	return gitStatus, nil
}