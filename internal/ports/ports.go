package ports

import (
	"context"
	"io"
	"time"

	"ArticlesScanner/internal/domain"
)

// ArticleSource pulls fresh articles from upstream providers.
type ArticleSource interface {
	FetchDaily(ctx context.Context, day time.Time) ([]domain.Article, error)
}

// ArticleRepository persists processed articles for deduplication/history.
type ArticleRepository interface {
	AlreadyProcessed(ctx context.Context, ids []string) (map[string]bool, error)
	SaveProcessed(ctx context.Context, article domain.ProcessedArticle) error
}

// Analyzer pushes abstracts to ML models for scoring and topic extraction.
type Analyzer interface {
	Rank(ctx context.Context, article domain.Article) (domain.ArticleReview, error)
}

// Summarizer generates final summaries of downloaded articles.
type Summarizer interface {
	Summarize(ctx context.Context, article domain.Article, content []byte) (string, error)
}

// Downloader fetches full-text PDFs or HTML payloads.
type Downloader interface {
	Download(ctx context.Context, article domain.Article) (io.ReadCloser, error)
}

// Notifier streams selected digests to Telegram or other channels.
type Notifier interface {
	PublishDigest(ctx context.Context, digest string) error
}

// ChatClient pushes structured digests to LLM APIs (e.g., ChatGPT).
type ChatClient interface {
	SendDigest(ctx context.Context, payload []byte) error
}

// Scheduler controls when pipelines execute.
type Scheduler interface {
	Start(ctx context.Context, job func(time.Time)) error
	Stop(ctx context.Context) error
}
