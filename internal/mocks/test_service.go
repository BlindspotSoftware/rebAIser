package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type MockTestService struct {
	mock.Mock
}

func (m *MockTestService) RunTests(ctx context.Context, workingDir string) (*interfaces.TestResult, error) {
	args := m.Called(ctx, workingDir)
	return args.Get(0).(*interfaces.TestResult), args.Error(1)
}

func (m *MockTestService) RunCommand(ctx context.Context, cmd interfaces.TestCommand) (*interfaces.CommandResult, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*interfaces.CommandResult), args.Error(1)
}