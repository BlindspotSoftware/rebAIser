package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type MockGitHubService struct {
	mock.Mock
}

func (m *MockGitHubService) CreatePullRequest(ctx context.Context, req interfaces.CreatePRRequest) (*interfaces.PullRequest, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*interfaces.PullRequest), args.Error(1)
}

func (m *MockGitHubService) MergePullRequest(ctx context.Context, prNumber int) error {
	args := m.Called(ctx, prNumber)
	return args.Error(0)
}

func (m *MockGitHubService) GetPullRequest(ctx context.Context, prNumber int) (*interfaces.PullRequest, error) {
	args := m.Called(ctx, prNumber)
	return args.Get(0).(*interfaces.PullRequest), args.Error(1)
}

func (m *MockGitHubService) ListPullRequests(ctx context.Context, state string) ([]*interfaces.PullRequest, error) {
	args := m.Called(ctx, state)
	return args.Get(0).([]*interfaces.PullRequest), args.Error(1)
}

func (m *MockGitHubService) AddReviewers(ctx context.Context, prNumber int, reviewers []string) error {
	args := m.Called(ctx, prNumber, reviewers)
	return args.Error(0)
}