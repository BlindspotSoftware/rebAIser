package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type Service struct {
	webhookURL string
	channel    string
	username   string
	httpClient *http.Client
	log        *logrus.Entry
}

func NewService(webhookURL, channel, username string) interfaces.NotifyService {
	return &Service{
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: logrus.WithField("component", "notify"),
	}
}

func (s *Service) SendMessage(ctx context.Context, message interfaces.NotificationMessage) error {
	s.log.WithFields(logrus.Fields{
		"title": message.Title,
		"level": message.Level,
	}).Info("Sending Slack notification")

	// Skip sending if webhook URL is not configured
	if s.webhookURL == "" {
		s.log.Info("Slack webhook URL not configured, skipping notification")
		return nil
	}

	// Create Slack message payload
	slackPayload := s.createSlackPayload(message)

	// Send webhook request
	err := s.sendWebhook(ctx, slackPayload)
	if err != nil {
		s.log.WithError(err).Error("Failed to send Slack notification")
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"title":   message.Title,
		"channel": s.channel,
	}).Info("Slack notification sent successfully")

	return nil
}

// SlackPayload represents the structure of a Slack webhook payload
type SlackPayload struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	Text        string            `json:"text,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	TitleLink string       `json:"title_link,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Footer    string       `json:"footer,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// createSlackPayload creates a Slack webhook payload from a notification message
func (s *Service) createSlackPayload(message interfaces.NotificationMessage) SlackPayload {
	// Determine color based on notification level
	color := s.getColorForLevel(message.Level)
	
	// Create attachment
	attachment := SlackAttachment{
		Color:     color,
		Title:     message.Title,
		TitleLink: message.URL,
		Text:      message.Message,
		Footer:    "AI Rebaser",
		Timestamp: time.Now().Unix(),
	}
	
	// Add fields if we have additional context
	if message.URL != "" {
		attachment.Fields = append(attachment.Fields, SlackField{
			Title: "Link",
			Value: message.URL,
			Short: true,
		})
	}
	
	payload := SlackPayload{
		Channel:     s.channel,
		Username:    s.username,
		IconEmoji:   s.getEmojiForLevel(message.Level),
		Attachments: []SlackAttachment{attachment},
	}
	
	return payload
}

// getColorForLevel returns the appropriate color for the notification level
func (s *Service) getColorForLevel(level interfaces.NotificationLevel) string {
	switch level {
	case interfaces.NotificationLevelSuccess:
		return "good"  // Green
	case interfaces.NotificationLevelWarning:
		return "warning"  // Yellow
	case interfaces.NotificationLevelError:
		return "danger"  // Red
	case interfaces.NotificationLevelInfo:
		return "#36a64f"  // Blue
	default:
		return "#36a64f"  // Default to blue
	}
}

// getEmojiForLevel returns the appropriate emoji for the notification level
func (s *Service) getEmojiForLevel(level interfaces.NotificationLevel) string {
	switch level {
	case interfaces.NotificationLevelSuccess:
		return ":white_check_mark:"
	case interfaces.NotificationLevelWarning:
		return ":warning:"
	case interfaces.NotificationLevelError:
		return ":x:"
	case interfaces.NotificationLevelInfo:
		return ":information_source:"
	default:
		return ":robot_face:"
	}
}

// sendWebhook sends the payload to the Slack webhook URL
func (s *Service) sendWebhook(ctx context.Context, payload SlackPayload) error {
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AI-Rebaser/1.0")
	
	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned non-200 status: %d", resp.StatusCode)
	}
	
	return nil
}