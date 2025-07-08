package interfaces

import "context"

type GitService interface {
	Clone(ctx context.Context, repo, dir string) error
	Fetch(ctx context.Context, dir string) error
	Rebase(ctx context.Context, dir, branch string) error
	GetConflicts(ctx context.Context, dir string) ([]GitConflict, error)
	ResolveConflict(ctx context.Context, dir, file, resolution string) error
	Commit(ctx context.Context, dir, message string) error
	Push(ctx context.Context, dir, branch string) error
	CreateBranch(ctx context.Context, dir, branch string) error
	GetStatus(ctx context.Context, dir string) (GitStatus, error)
}

type GitConflict struct {
	File    string
	Content string
	Ours    string
	Theirs  string
}

type GitStatus struct {
	IsClean       bool
	HasConflicts  bool
	ModifiedFiles []string
	ConflictFiles []string
}