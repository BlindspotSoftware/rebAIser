package interfaces

import "context"

type GitHubService interface {
	CreatePullRequest(ctx context.Context, req CreatePRRequest) (*PullRequest, error)
	MergePullRequest(ctx context.Context, prNumber int) error
	GetPullRequest(ctx context.Context, prNumber int) (*PullRequest, error)
	ListPullRequests(ctx context.Context, state string) ([]*PullRequest, error)
	AddReviewers(ctx context.Context, prNumber int, reviewers []string) error
}

type CreatePRRequest struct {
	Title       string
	Body        string
	Head        string
	Base        string
	Draft       bool
	Maintainer  bool
}

type PullRequest struct {
	Number    int
	Title     string
	Body      string
	State     string
	Head      string
	Base      string
	HTMLURL   string
	Mergeable bool
	Draft     bool
	CreatedAt string
	UpdatedAt string
}