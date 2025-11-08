package domain

import "time"

// Article is a core entity describing metadata fetched from providers.
type Article struct {
	ID          string
	Title       string
	Abstract    string
	URL         string
	Source      string
	PublishedAt time.Time
}

// ArticleReview captures ML scoring and enrichment for prioritization.
type ArticleReview struct {
	Article   Article
	Score     float64
	Topics    []string
	Summary   string
	RankedAt  time.Time
	Processed bool
}

// ProcessingStatus enumerates pipeline milestones.
type ProcessingStatus string

const (
	StatusFetched    ProcessingStatus = "fetched"
	StatusRanked     ProcessingStatus = "ranked"
	StatusSummarized ProcessingStatus = "summarized"
	StatusDelivered  ProcessingStatus = "delivered"
)

// ProcessedArticle persisted to Postgres for deduplication and audit.
type ProcessedArticle struct {
	Article   Article
	Summary   string
	Score     float64
	Status    ProcessingStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
