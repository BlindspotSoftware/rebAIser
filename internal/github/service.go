package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type Service struct {
	client *github.Client
	owner  string
	repo   string
	log    *logrus.Entry
}

func NewService(token, owner, repo string) interfaces.GitHubService {
	// Create OAuth2 token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	
	// Create GitHub client
	client := github.NewClient(tc)
	
	return &Service{
		client: client,
		owner:  owner,
		repo:   repo,
		log:    logrus.WithField("component", "github"),
	}
}

func (s *Service) CreatePullRequest(ctx context.Context, req interfaces.CreatePRRequest) (*interfaces.PullRequest, error) {
	s.log.WithFields(logrus.Fields{
		"title": req.Title,
		"head":  req.Head,
		"base":  req.Base,
	}).Info("Creating pull request")

	// Create GitHub pull request
	prRequest := &github.NewPullRequest{
		Title: github.String(req.Title),
		Head:  github.String(req.Head),
		Base:  github.String(req.Base),
		Body:  github.String(req.Body),
	}

	if req.Draft {
		prRequest.Draft = github.Bool(true)
	}

	ghPR, _, err := s.client.PullRequests.Create(ctx, s.owner, s.repo, prRequest)
	if err != nil {
		s.log.WithError(err).Error("Failed to create pull request")
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	pr := &interfaces.PullRequest{
		Number:    *ghPR.Number,
		Title:     *ghPR.Title,
		Body:      getStringValue(ghPR.Body),
		State:     *ghPR.State,
		Head:      *ghPR.Head.Ref,
		Base:      *ghPR.Base.Ref,
		HTMLURL:   *ghPR.HTMLURL,
		Mergeable: getBoolValue(ghPR.Mergeable),
		Draft:     getBoolValue(ghPR.Draft),
	}

	s.log.WithFields(logrus.Fields{
		"prNumber": pr.Number,
		"url":      pr.HTMLURL,
	}).Info("Pull request created successfully")

	return pr, nil
}

func (s *Service) MergePullRequest(ctx context.Context, prNumber int) error {
	s.log.WithField("prNumber", prNumber).Info("Merging pull request with rebase method")

	// First check if PR is mergeable
	pr, _, err := s.client.PullRequests.Get(ctx, s.owner, s.repo, prNumber)
	if err != nil {
		s.log.WithError(err).Error("Failed to get pull request")
		return fmt.Errorf("failed to get pull request: %w", err)
	}

	if pr.Mergeable != nil && !*pr.Mergeable {
		return fmt.Errorf("pull request #%d is not mergeable", prNumber)
	}

	if *pr.State != "open" {
		return fmt.Errorf("pull request #%d is not open (state: %s)", prNumber, *pr.State)
	}

	// Merge the pull request using rebase method
	commitMessage := fmt.Sprintf("Rebase pull request #%d", prNumber)
	mergeOptions := &github.PullRequestOptions{
		CommitTitle: commitMessage,
		MergeMethod: "rebase", // Use rebase method for AI-assisted rebasing tool
	}

	mergeResult, _, err := s.client.PullRequests.Merge(ctx, s.owner, s.repo, prNumber, "", mergeOptions)
	if err != nil {
		s.log.WithError(err).Error("Failed to merge pull request")
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	if !*mergeResult.Merged {
		return fmt.Errorf("pull request #%d was not merged: %s", prNumber, getStringValue(mergeResult.Message))
	}

	s.log.WithFields(logrus.Fields{
		"prNumber": prNumber,
		"sha":      getStringValue(mergeResult.SHA),
	}).Info("Pull request rebased and merged successfully")

	return nil
}

func (s *Service) GetPullRequest(ctx context.Context, prNumber int) (*interfaces.PullRequest, error) {
	s.log.WithField("prNumber", prNumber).Info("Getting pull request")

	ghPR, _, err := s.client.PullRequests.Get(ctx, s.owner, s.repo, prNumber)
	if err != nil {
		s.log.WithError(err).Error("Failed to get pull request")
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	pr := &interfaces.PullRequest{
		Number:    *ghPR.Number,
		Title:     *ghPR.Title,
		Body:      getStringValue(ghPR.Body),
		State:     *ghPR.State,
		Head:      *ghPR.Head.Ref,
		Base:      *ghPR.Base.Ref,
		HTMLURL:   *ghPR.HTMLURL,
		Mergeable: getBoolValue(ghPR.Mergeable),
		Draft:     getBoolValue(ghPR.Draft),
	}

	return pr, nil
}

func (s *Service) ListPullRequests(ctx context.Context, state string) ([]*interfaces.PullRequest, error) {
	s.log.WithField("state", state).Info("Listing pull requests")

	// Validate state parameter
	validStates := map[string]bool{
		"open":   true,
		"closed": true,
		"all":    true,
	}
	if !validStates[state] {
		return nil, fmt.Errorf("invalid state '%s', must be 'open', 'closed', or 'all'", state)
	}

	// List pull requests
	listOptions := &github.PullRequestListOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100, // Max per page
		},
	}

	var allPRs []*interfaces.PullRequest
	for {
		ghPRs, resp, err := s.client.PullRequests.List(ctx, s.owner, s.repo, listOptions)
		if err != nil {
			s.log.WithError(err).Error("Failed to list pull requests")
			return nil, fmt.Errorf("failed to list pull requests: %w", err)
		}

		// Convert GitHub PRs to interface PRs
		for _, ghPR := range ghPRs {
			pr := &interfaces.PullRequest{
				Number:    *ghPR.Number,
				Title:     *ghPR.Title,
				Body:      getStringValue(ghPR.Body),
				State:     *ghPR.State,
				Head:      *ghPR.Head.Ref,
				Base:      *ghPR.Base.Ref,
				HTMLURL:   *ghPR.HTMLURL,
				Mergeable: getBoolValue(ghPR.Mergeable),
				Draft:     getBoolValue(ghPR.Draft),
			}
			allPRs = append(allPRs, pr)
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}

	s.log.WithField("count", len(allPRs)).Info("Listed pull requests")
	return allPRs, nil
}

func (s *Service) AddReviewers(ctx context.Context, prNumber int, reviewers []string) error {
	s.log.WithFields(logrus.Fields{
		"prNumber":  prNumber,
		"reviewers": reviewers,
	}).Info("Adding reviewers to pull request")

	if len(reviewers) == 0 {
		return nil // No reviewers to add
	}

	// Split reviewers into individual users and teams
	var users, teams []string
	for _, reviewer := range reviewers {
		// Teams are prefixed with @ or contain /
		if strings.HasPrefix(reviewer, "@") || strings.Contains(reviewer, "/") {
			// Remove @ prefix if present
			team := strings.TrimPrefix(reviewer, "@")
			teams = append(teams, team)
		} else {
			users = append(users, reviewer)
		}
	}

	// Create review request
	reviewRequest := github.ReviewersRequest{}
	if len(users) > 0 {
		reviewRequest.Reviewers = users
	}
	if len(teams) > 0 {
		reviewRequest.TeamReviewers = teams
	}

	_, _, err := s.client.PullRequests.RequestReviewers(ctx, s.owner, s.repo, prNumber, reviewRequest)
	if err != nil {
		s.log.WithError(err).Error("Failed to add reviewers")
		return fmt.Errorf("failed to add reviewers: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"prNumber": prNumber,
		"users":    users,
		"teams":    teams,
	}).Info("Reviewers added successfully")

	return nil
}

// Helper functions for safe pointer dereferencing

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}