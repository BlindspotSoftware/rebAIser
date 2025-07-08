package notify

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/9elements/rebaiser/internal/interfaces"
)

type Service struct {
	webhookURL string
	channel    string
	username   string
	log        *logrus.Entry
}

func NewService(webhookURL, channel, username string) interfaces.NotifyService {
	return &Service{
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
		log:        logrus.WithField("component", "notify"),
	}
}

func (s *Service) SendMessage(ctx context.Context, message interfaces.NotificationMessage) error {
	s.log.WithFields(logrus.Fields{
		"title": message.Title,
		"level": message.Level,
	}).Info("Sending notification")

	// TODO: Implement Slack webhook call
	// For now, just log the message
	s.log.WithFields(logrus.Fields{
		"title":   message.Title,
		"message": message.Message,
		"url":     message.URL,
		"level":   message.Level,
	}).Info("Notification sent")

	return nil
}