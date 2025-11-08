package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"ArticlesScanner/internal/domain"
	"ArticlesScanner/internal/ports"
)

// PostgresRepository persists processed articles into Postgres.
type PostgresRepository struct {
	db *sql.DB
}

var _ ports.ArticleRepository = (*PostgresRepository)(nil)

// NewPostgresRepository wires a sql.DB implementation.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// AlreadyProcessed returns a map with IDs that already exist in storage.
func (r *PostgresRepository) AlreadyProcessed(ctx context.Context, ids []string) (map[string]bool, error) {
	if r.db == nil || len(ids) == 0 {
		return map[string]bool{}, nil
	}

	query := `SELECT external_id FROM processed_articles WHERE external_id = ANY($1)`

	rows, err := r.db.QueryContext(ctx, query, pq.StringArray(ids))
	if err != nil {
		return nil, fmt.Errorf("query processed: %w", err)
	}

	result := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan id: %w", err)
		}
		result[id] = true
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("rows iteration: %w", rowsErr)
	}

	if closeErr := rows.Close(); closeErr != nil {
		return nil, fmt.Errorf("close rows: %w", closeErr)
	}

	return result, nil
}

// SaveProcessed upserts the processed article snapshot.
func (r *PostgresRepository) SaveProcessed(ctx context.Context, article domain.ProcessedArticle) error {
	if r.db == nil {
		return nil
	}

	query := `INSERT INTO processed_articles (external_id, title, summary, score, status)
              VALUES ($1, $2, $3, $4, $5)
              ON CONFLICT (external_id) DO UPDATE
              SET summary = EXCLUDED.summary,
                  score = EXCLUDED.score,
                  status = EXCLUDED.status,
                  updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		article.Article.ID,
		article.Article.Title,
		article.Summary,
		article.Score,
		article.Status,
	)
	if err != nil {
		return fmt.Errorf("upsert processed: %w", err)
	}

	return nil
}
