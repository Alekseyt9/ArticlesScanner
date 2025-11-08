BEGIN;

CREATE TABLE IF NOT EXISTS processed_articles (
    external_id TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    summary     TEXT,
    score       DOUBLE PRECISION,
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION set_processed_articles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_processed_articles_updated_at ON processed_articles;
CREATE TRIGGER trg_processed_articles_updated_at
BEFORE UPDATE ON processed_articles
FOR EACH ROW
EXECUTE FUNCTION set_processed_articles_updated_at();

COMMIT;
