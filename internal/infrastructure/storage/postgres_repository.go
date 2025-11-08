package storage

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

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

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// AlreadyProcessed returns a map with IDs that already exist in storage.
func (r *PostgresRepository) AlreadyProcessed(ctx context.Context, ids []string) (map[string]bool, error) {
	if r.db == nil || len(ids) == 0 {
		return map[string]bool{}, nil
	}

	query, args, err := psql.
		Select("external_id").
		From("processed_articles").
		Where(sq.Eq{"external_id": ids}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build processed query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
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

	query, args, err := psql.
		Insert("processed_articles").
		Columns("external_id", "title", "summary", "score", "status").
		Values(
			article.Article.ID,
			article.Article.Title,
			article.Summary,
			article.Score,
			article.Status,
		).
		Suffix("ON CONFLICT (external_id) DO UPDATE SET summary = EXCLUDED.summary, score = EXCLUDED.score, status = EXCLUDED.status, updated_at = NOW()").
		ToSql()
	if err != nil {
		return fmt.Errorf("build upsert processed: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("upsert processed: %w", err)
	}

	return nil
}
