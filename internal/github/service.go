package github

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/9elements/rebaiser/internal/interfaces"
)

type Service struct {
	token  string
	owner  string
	repo   string
	log    *logrus.Entry
}

func NewService(token, owner, repo string) interfaces.GitHubService {
	return &Service{
		token: token,
		owner: owner,
		repo:  repo,
		log:   logrus.WithField("component", "github"),
	}
}

func (s *Service) CreatePullRequest(ctx context.Context, req interfaces.CreatePRRequest) (*interfaces.PullRequest, error) {
	s.log.WithFields(logrus.Fields{
		"title": req.Title,
		"head":  req.Head,
		"base":  req.Base,
	}).Info("Creating pull request")

	// TODO: Implement GitHub API call
	// For now, return a mock PR
	pr := &interfaces.PullRequest{
		Number:  123,
		Title:   req.Title,
		Body:    req.Body,
		State:   "open",
		Head:    req.Head,
		Base:    req.Base,
		HTMLURL: fmt.Sprintf("https://github.com/%s/%s/pull/123", s.owner, s.repo),
	}

	return pr, nil
}

func (s *Service) MergePullRequest(ctx context.Context, prNumber int) error {
	s.log.WithField("prNumber", prNumber).Info("Merging pull request")

	// TODO: Implement GitHub API call
	return nil
}

func (s *Service) GetPullRequest(ctx context.Context, prNumber int) (*interfaces.PullRequest, error) {
	s.log.WithField("prNumber", prNumber).Info("Getting pull request")

	// TODO: Implement GitHub API call
	pr := &interfaces.PullRequest{
		Number: prNumber,
		State:  "open",
	}

	return pr, nil
}

func (s *Service) ListPullRequests(ctx context.Context, state string) ([]*interfaces.PullRequest, error) {
	s.log.WithField("state", state).Info("Listing pull requests")

	// TODO: Implement GitHub API call
	return []*interfaces.PullRequest{}, nil
}

func (s *Service) AddReviewers(ctx context.Context, prNumber int, reviewers []string) error {
	s.log.WithFields(logrus.Fields{
		"prNumber":  prNumber,
		"reviewers": reviewers,
	}).Info("Adding reviewers to pull request")

	// TODO: Implement GitHub API call
	return nil
}