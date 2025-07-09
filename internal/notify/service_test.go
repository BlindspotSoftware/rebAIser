package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

func TestService_SendMessage(t *testing.T) {
	// Create mock Slack server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "AI-Rebaser/1.0", r.Header.Get("User-Agent"))
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	// Create service with test webhook URL
	service := NewService(server.URL, "#test-channel", "Test Bot")

	// Test sending a notification
	message := interfaces.NotificationMessage{
		Title:   "Test Notification",
		Message: "This is a test message",
		URL:     "https://github.com/test/repo/pull/123",
		Level:   interfaces.NotificationLevelSuccess,
	}

	err := service.SendMessage(context.Background(), message)
	require.NoError(t, err)
}

func TestService_SendMessage_NoWebhookURL(t *testing.T) {
	// Create service without webhook URL
	service := NewService("", "#test-channel", "Test Bot")

	message := interfaces.NotificationMessage{
		Title:   "Test Notification",
		Message: "This is a test message",
		Level:   interfaces.NotificationLevelInfo,
	}

	// Should not error when no webhook URL is configured
	err := service.SendMessage(context.Background(), message)
	require.NoError(t, err)
}

func TestService_createSlackPayload(t *testing.T) {
	service := &Service{
		channel:  "#test-channel",
		username: "Test Bot",
	}

	message := interfaces.NotificationMessage{
		Title:   "Test Notification",
		Message: "This is a test message",
		URL:     "https://github.com/test/repo/pull/123",
		Level:   interfaces.NotificationLevelSuccess,
	}

	payload := service.createSlackPayload(message)

	assert.Equal(t, "#test-channel", payload.Channel)
	assert.Equal(t, "Test Bot", payload.Username)
	assert.Equal(t, ":white_check_mark:", payload.IconEmoji)
	assert.Len(t, payload.Attachments, 1)

	attachment := payload.Attachments[0]
	assert.Equal(t, "good", attachment.Color)
	assert.Equal(t, "Test Notification", attachment.Title)
	assert.Equal(t, "https://github.com/test/repo/pull/123", attachment.TitleLink)
	assert.Equal(t, "This is a test message", attachment.Text)
	assert.Equal(t, "AI Rebaser", attachment.Footer)
	assert.Greater(t, attachment.Timestamp, int64(0))
}

func TestService_getColorForLevel(t *testing.T) {
	service := &Service{}

	tests := []struct {
		level    interfaces.NotificationLevel
		expected string
	}{
		{interfaces.NotificationLevelSuccess, "good"},
		{interfaces.NotificationLevelWarning, "warning"},
		{interfaces.NotificationLevelError, "danger"},
		{interfaces.NotificationLevelInfo, "#36a64f"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			color := service.getColorForLevel(tt.level)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestService_getEmojiForLevel(t *testing.T) {
	service := &Service{}

	tests := []struct {
		level    interfaces.NotificationLevel
		expected string
	}{
		{interfaces.NotificationLevelSuccess, ":white_check_mark:"},
		{interfaces.NotificationLevelWarning, ":warning:"},
		{interfaces.NotificationLevelError, ":x:"},
		{interfaces.NotificationLevelInfo, ":information_source:"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			emoji := service.getEmojiForLevel(tt.level)
			assert.Equal(t, tt.expected, emoji)
		})
	}
}

func TestService_sendWebhook_ErrorHandling(t *testing.T) {
	// Test with server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error"))
	}))
	defer server.Close()

	service := &Service{
		webhookURL: server.URL,
		httpClient: &http.Client{Timeout: 1 * time.Second},
	}

	payload := SlackPayload{
		Channel:  "#test",
		Username: "Test Bot",
		Text:     "Test message",
	}

	err := service.sendWebhook(context.Background(), payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-200 status: 500")
}

func TestService_sendWebhook_Timeout(t *testing.T) {
	// Test with server that times out
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := &Service{
		webhookURL: server.URL,
		httpClient: &http.Client{Timeout: 500 * time.Millisecond},
	}

	payload := SlackPayload{
		Channel:  "#test",
		Username: "Test Bot",
		Text:     "Test message",
	}

	err := service.sendWebhook(context.Background(), payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeded")
}