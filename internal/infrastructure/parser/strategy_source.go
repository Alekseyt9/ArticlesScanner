package parser

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"ArticlesScanner/internal/config"
	"ArticlesScanner/internal/domain"
	"ArticlesScanner/internal/ports"
	"ArticlesScanner/internal/scanner"
)

// StrategySource implements ArticleSource via registered scanner strategies.
type StrategySource struct {
	registry *scanner.Registry
	sites    []config.SiteConfig
	logger   *slog.Logger
}

var _ ports.ArticleSource = (*StrategySource)(nil)

// NewStrategySource wires scanner registry with config-defined sites.
func NewStrategySource(reg *scanner.Registry, sites []config.SiteConfig, log *slog.Logger) *StrategySource {
	return &StrategySource{
		registry: reg,
		sites:    sites,
		logger:   log,
	}
}

// FetchDaily iterates over configured sites and executes their scanners.
func (s *StrategySource) FetchDaily(ctx context.Context, day time.Time) ([]domain.Article, error) {
	if s.registry == nil {
		return nil, fmt.Errorf("scanner registry is not configured")
	}

	s.debug("fetch daily", "sites", len(s.sites), "day", day.Format("2006-01-02"))

	var aggregated []domain.Article
	for _, site := range s.sites {
		s.debug("process site", "site", site.Name, "scanner", site.Scanner, "categories", len(site.Categories))
		strategy, err := s.registry.Resolve(site.Scanner)
		if err != nil {
			return nil, fmt.Errorf("site %s: %w", site.Name, err)
		}

		req := scanner.Request{
			Day:        day,
			SiteName:   site.Name,
			Options:    site.Options,
			Categories: toScannerCategories(site.Categories),
		}

		results, err := strategy.Scan(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("scan site %s: %w", site.Name, err)
		}

		for i := range results {
			if results[i].Source == "" {
				results[i].Source = site.Name
			}
		}
		s.debug("site produced articles", "site", site.Name, "count", len(results))
		aggregated = append(aggregated, results...)
	}

	s.debug("strategy source done", "total_articles", len(aggregated))
	return aggregated, nil
}

func toScannerCategories(cfg []config.CategoryConfig) []scanner.Category {
	categories := make([]scanner.Category, 0, len(cfg))
	for _, cat := range cfg {
		categories = append(categories, scanner.Category{
			Name: cat.Name,
			URL:  cat.URL,
		})
	}
	return categories
}

func (s *StrategySource) debug(msg string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Debug(msg, args...)
	}
}
