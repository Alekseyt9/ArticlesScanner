package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ArticlesScanner/internal/config"
	"ArticlesScanner/internal/ports"
)

// ChatGPTClient implements ports.ChatClient backed by OpenAI-compatible APIs.
type ChatGPTClient struct {
	endpoint     string
	model        string
	apiKey       string
	systemPrompt string
	httpClient   *http.Client
}

var _ ports.ChatClient = (*ChatGPTClient)(nil)

// NewChatGPTClient builds a client from configuration.
func NewChatGPTClient(cfg config.ChatGPTConfig) *ChatGPTClient {
	return &ChatGPTClient{
		endpoint:     cfg.Endpoint,
		model:        cfg.Model,
		apiKey:       cfg.APIKey,
		systemPrompt: cfg.SystemPrompt,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// SendDigest posts the JSON payload as a user message to ChatGPT.
func (c *ChatGPTClient) SendDigest(ctx context.Context, payload []byte) error {
	if c == nil {
		return fmt.Errorf("chatgpt client is nil")
	}
	if c.apiKey == "" || c.endpoint == "" || c.model == "" {
		return fmt.Errorf("chatgpt client misconfigured")
	}

	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": safePrompt(c.systemPrompt)},
			{"role": "user", "content": string(payload)},
		},
	})
	if err != nil {
		return fmt.Errorf("marshal chatgpt payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send digest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("chatgpt error %s: %s", resp.Status, strings.TrimSpace(string(payload)))
	}

	return nil
}

func safePrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "You are a helpful assistant that receives article digests."
	}
	return prompt
}
