package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) ResolveConflict(ctx context.Context, conflict interfaces.GitConflict) (string, error) {
	args := m.Called(ctx, conflict)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) GenerateCommitMessage(ctx context.Context, changes []string) (string, error) {
	args := m.Called(ctx, changes)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) GenerateCommitMessageWithConflicts(ctx context.Context, changes []string, conflicts []interfaces.GitConflict) (string, error) {
	args := m.Called(ctx, changes, conflicts)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) GeneratePRDescription(ctx context.Context, commits []string, conflicts []interfaces.GitConflict) (string, error) {
	args := m.Called(ctx, commits, conflicts)
	return args.String(0), args.Error(1)
}