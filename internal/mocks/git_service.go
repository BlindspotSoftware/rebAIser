package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type MockGitService struct {
	mock.Mock
}

func (m *MockGitService) Clone(ctx context.Context, repo, dir string) error {
	args := m.Called(ctx, repo, dir)
	return args.Error(0)
}

func (m *MockGitService) Fetch(ctx context.Context, dir string) error {
	args := m.Called(ctx, dir)
	return args.Error(0)
}

func (m *MockGitService) Rebase(ctx context.Context, dir, branch string) error {
	args := m.Called(ctx, dir, branch)
	return args.Error(0)
}

func (m *MockGitService) GetConflicts(ctx context.Context, dir string) ([]interfaces.GitConflict, error) {
	args := m.Called(ctx, dir)
	return args.Get(0).([]interfaces.GitConflict), args.Error(1)
}

func (m *MockGitService) ResolveConflict(ctx context.Context, dir, file, resolution string) error {
	args := m.Called(ctx, dir, file, resolution)
	return args.Error(0)
}

func (m *MockGitService) Commit(ctx context.Context, dir, message string) error {
	args := m.Called(ctx, dir, message)
	return args.Error(0)
}

func (m *MockGitService) Push(ctx context.Context, dir, branch string) error {
	args := m.Called(ctx, dir, branch)
	return args.Error(0)
}

func (m *MockGitService) CreateBranch(ctx context.Context, dir, branch string) error {
	args := m.Called(ctx, dir, branch)
	return args.Error(0)
}

func (m *MockGitService) GetStatus(ctx context.Context, dir string) (interfaces.GitStatus, error) {
	args := m.Called(ctx, dir)
	return args.Get(0).(interfaces.GitStatus), args.Error(1)
}

func (m *MockGitService) AddRemote(ctx context.Context, dir, name, url string) error {
	args := m.Called(ctx, dir, name, url)
	return args.Error(0)
}