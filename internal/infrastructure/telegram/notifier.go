package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ArticlesScanner/internal/ports"
)

// Notifier sends digests to a Telegram chat via bot API.
type Notifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

var _ ports.Notifier = (*Notifier)(nil)

// NewNotifier registers bot token and chat identifier.
func NewNotifier(botToken, chatID string) *Notifier {
	return &Notifier{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

// PublishDigest posts a Markdown message to Telegram.
func (n *Notifier) PublishDigest(ctx context.Context, digest string) error {
	if n.botToken == "" || n.chatID == "" || n.client == nil {
		return fmt.Errorf("telegram notifier misconfigured")
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
	form := url.Values{}
	form.Set("chat_id", n.chatID)
	form.Set("text", digest)
	form.Set("parse_mode", "Markdown")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			return fmt.Errorf("telegram error: %s, close body: %v", resp.Status, closeErr)
		}
		return fmt.Errorf("telegram error: %s", resp.Status)
	}

	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("close telegram response body: %w", err)
	}

	return nil
}
