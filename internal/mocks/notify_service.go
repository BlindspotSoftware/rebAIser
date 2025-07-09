package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type MockNotifyService struct {
	mock.Mock
}

func (m *MockNotifyService) SendMessage(ctx context.Context, message interfaces.NotificationMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}