package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ArticlesScanner/internal/domain"
	"ArticlesScanner/internal/ports"
)

// Client talks to an external ML service for ranking and summarization.
type Client struct {
	endpoint string
	apiKey   string
	http     *http.Client
}

var _ ports.Analyzer = (*Client)(nil)
var _ ports.Summarizer = (*Client)(nil)

// NewClient creates a reusable HTTP client.
func NewClient(endpoint, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 15 * time.Second},
	}
}

// Rank sends the abstract for scoring and topic detection.
func (c *Client) Rank(ctx context.Context, article domain.Article) (domain.ArticleReview, error) {
	if c.http == nil {
		return domain.ArticleReview{Article: article}, nil
	}

	payload := map[string]any{
		"title":    article.Title,
		"abstract": article.Abstract,
	}

	review := domain.ArticleReview{Article: article}
	if err := c.post(ctx, "/rank", payload, &review); err != nil {
		return domain.ArticleReview{}, err
	}

	return review, nil
}

// Summarize requests a summary for the downloaded article content.
func (c *Client) Summarize(ctx context.Context, article domain.Article, content []byte) (string, error) {
	if c.http == nil {
		return string(content), nil
	}

	payload := map[string]any{
		"title":   article.Title,
		"content": string(content),
	}

	var resp struct {
		Summary string `json:"summary"`
	}

	if err := c.post(ctx, "/summarize", payload, &resp); err != nil {
		return "", err
	}

	return resp.Summary, nil
}

func (c *Client) post(ctx context.Context, path string, payload any, v any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			return fmt.Errorf("unexpected status %s, close body: %v", resp.Status, closeErr)
		}
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	if v == nil {
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("close response body: %w", err)
		}
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		_ = resp.Body.Close()
		return fmt.Errorf("decode response: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("close response body: %w", err)
	}

	return nil
}
