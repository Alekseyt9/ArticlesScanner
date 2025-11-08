package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"ArticlesScanner/internal/domain"
	"ArticlesScanner/internal/ports"
)

// PipelineDeps wires all driven adapters into the orchestration pipeline.
type PipelineDeps struct {
	Source     ports.ArticleSource
	Repository ports.ArticleRepository
	Analyzer   ports.Analyzer
	Summarizer ports.Summarizer
	Downloader ports.Downloader
	Notifier   ports.Notifier
	ChatClient ports.ChatClient
	Logger     *slog.Logger
}

// Pipeline implements the article-ingestion workflow.
type Pipeline struct {
	source     ports.ArticleSource
	repository ports.ArticleRepository
	analyzer   ports.Analyzer
	summarizer ports.Summarizer
	downloader ports.Downloader
	notifier   ports.Notifier
	chatClient ports.ChatClient
	logger     *slog.Logger
}

// NewPipeline constructs the orchestration component.
func NewPipeline(deps PipelineDeps) *Pipeline {
	return &Pipeline{
		source:     deps.Source,
		repository: deps.Repository,
		analyzer:   deps.Analyzer,
		summarizer: deps.Summarizer,
		downloader: deps.Downloader,
		notifier:   deps.Notifier,
		chatClient: deps.ChatClient,
		logger:     deps.Logger,
	}
}

// ProcessDay orchestrates fetching, ranking, summarizing, and notifying.
func (p *Pipeline) ProcessDay(ctx context.Context, day time.Time) error {
	if p.source == nil {
		return nil
	}

	p.debug("starting pipeline", "day", day.Format("2006-01-02"))

	articles, err := p.source.FetchDaily(ctx, day)
	if err != nil {
		return fmt.Errorf("fetch daily: %w", err)
	}
	p.debug("source returned articles", "count", len(articles))

	ids := make([]string, len(articles))
	for i, art := range articles {
		ids[i] = art.ID
	}

	skip := map[string]bool{}
	if p.repository != nil && len(ids) > 0 {
		skip, err = p.repository.AlreadyProcessed(ctx, ids)
		if err != nil {
			return fmt.Errorf("load processed: %w", err)
		}
	}

	var digest []domain.ArticleReview
	for _, article := range articles {
		if skip[article.ID] {
			p.debug("skip article (already processed)", "article_id", article.ID)
			continue
		}

		p.debug("processing article", "article_id", article.ID)

		review := domain.ArticleReview{
			Article: article,
			Summary: article.Abstract,
		}

		if p.analyzer != nil {
			review, err = p.analyzer.Rank(ctx, article)
			if err != nil {
				return fmt.Errorf("rank article %s: %w", article.ID, err)
			}
		}

		var payload []byte
		if p.downloader != nil {
			reader, dErr := p.downloader.Download(ctx, article)
			if dErr != nil {
				return fmt.Errorf("download article %s: %w", article.ID, dErr)
			}
			if reader != nil {
				data, readErr := io.ReadAll(reader)
				closeErr := reader.Close()
				if readErr != nil {
					return fmt.Errorf("read article %s: %w", article.ID, readErr)
				}
				if closeErr != nil {
					return fmt.Errorf("close article %s: %w", article.ID, closeErr)
				}
				payload = data
			}
		}

		if p.summarizer != nil {
			summary, sErr := p.summarizer.Summarize(ctx, article, payload)
			if sErr != nil {
				return fmt.Errorf("summarize article %s: %w", article.ID, sErr)
			}
			review.Summary = summary
		}

		digest = append(digest, review)

		if p.repository != nil {
			err = p.repository.SaveProcessed(ctx, domain.ProcessedArticle{
				Article: review.Article,
				Summary: review.Summary,
				Score:   review.Score,
				Status:  domain.StatusDelivered,
			})
			if err != nil {
				return fmt.Errorf("persist article %s: %w", article.ID, err)
			}
		}
	}

	if len(digest) == 0 {
		p.debug("no articles processed", "day", day.Format("2006-01-02"))
		return nil
	}

	if p.chatClient != nil {
		payload, err := buildDigestJSON(digest)
		if err != nil {
			return fmt.Errorf("build chatgpt payload: %w", err)
		}
		if err := p.chatClient.SendDigest(ctx, payload); err != nil {
			return fmt.Errorf("send digest to chatgpt: %w", err)
		}
		p.debug("sent articles to chatgpt", "count", len(digest))
	}

	if p.notifier == nil {
		return nil
	}

	message := buildDigestMessage(digest)
	p.debug("publishing digest to notifier", "bytes", len(message))
	return p.notifier.PublishDigest(ctx, message)
}

func buildDigestMessage(reviews []domain.ArticleReview) string {
	if len(reviews) == 0 {
		return ""
	}

	var formatted string
	for _, review := range reviews {
		formatted += fmt.Sprintf("- %s\nScore: %.2f\n%s\n%s\n\n",
			review.Article.Title,
			review.Score,
			review.Summary,
			review.Article.URL)
	}

	return formatted
}

func buildDigestJSON(reviews []domain.ArticleReview) ([]byte, error) {
	type item struct {
		ID      string `json:"id"`
		URL     string `json:"url"`
		Summary string `json:"summary"`
		Source  string `json:"source"`
		Title   string `json:"title"`
	}

	payload := make([]item, 0, len(reviews))
	for _, review := range reviews {
		payload = append(payload, item{
			ID:      review.Article.ID,
			URL:     review.Article.URL,
			Summary: review.Summary,
			Source:  review.Article.Source,
			Title:   review.Article.Title,
		})
	}

	return json.Marshal(payload)
}

func (p *Pipeline) debug(msg string, args ...interface{}) {
	if p.logger != nil {
		p.logger.Debug(msg, args...)
	}
}
