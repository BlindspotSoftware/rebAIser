package interfaces

import "context"

type NotifyService interface {
	SendMessage(ctx context.Context, message NotificationMessage) error
}

type NotificationMessage struct {
	Title   string
	Message string
	URL     string
	Level   NotificationLevel
}

type NotificationLevel string

const (
	NotificationLevelInfo    NotificationLevel = "info"
	NotificationLevelWarning NotificationLevel = "warning"
	NotificationLevelError   NotificationLevel = "error"
	NotificationLevelSuccess NotificationLevel = "success"
)