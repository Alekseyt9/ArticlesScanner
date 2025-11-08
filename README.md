# ArticlesScanner

Scientific article parser built with Go 1.25.4 and clean architecture.

- Scheduler triggers the ingestion pipeline once per day.
- Scanner strategies crawl configured sites (Arxiv implementation included) and return daily summaries.
- Processed items are deduplicated in Postgres and pushed as JSON into ChatGPT while Telegram can still receive human‑readable digests.
- Run scripts for Linux/Windows export the required environment variables so secrets stay out of Git.

## Directory structure

```
cmd/articlescanner     # application entry point (main)
internal/app           # application wiring and lifecycle
internal/config        # configuration structs and loader
internal/domain        # entities and business rules
internal/ports         # inbound/outbound interfaces
internal/scanner       # strategy registry abstractions
internal/usecase       # orchestration logic (pipeline, scheduler)
internal/infrastructure# adapters (parser strategies, storage, ml, llm, scheduler, telegram)
internal/logging       # slog helper wiring
configs/               # YAML configuration (real file gitignored, example tracked)
scripts/               # run script templates for Linux/Windows
```

## Configuration & scripts

1. Copy `configs/config.example.yaml` to `configs/config.yaml` and adjust:
   - Add/rename sites, choose scanner name (`arxiv`) and list category URLs (each URL may point to a filtered Arxiv listing).
   - Provide ChatGPT endpoint/model/key plus optional system prompt.
   - Pick `logging.level` (`debug`, `info`, `warn`, `error`) — default is verbose debug logging.
2. Copy run script templates:
   - `cp scripts/run.example.sh scripts/run.sh` (Linux/macOS) or `copy scripts\run.example.cmd scripts\run.cmd` (Windows).
   - Edit the new files to export real secrets (DSN, API keys, config path). Actual files are ignored by Git.
3. Execute the desired script; it sets the env vars and runs `go run ./cmd/articlescanner`.

Important env vars:

- `ARTICLE_SCANNER_CONFIG` – path to the YAML config (defaults to `./configs/config.yaml`).
- `DATABASE_DSN`, `CHATGPT_API_KEY`, `CHATGPT_MODEL`, `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`.

## Tooling

- `make lint` — runs `gofmt`, `golangci-lint`, and `go vet ./...`.
- `make test` — runs `go test ./...`.

## Next steps

- Implement concrete storage, downloader, analyzer, and summarizer adapters plus migrations/tests.
- Replace the toy ticker with a cron library (e.g., `github.com/robfig/cron/v3`) and wire dependency injection/container logic.
- Extend scanner registry with more strategies (e.g., IEEE, PubMed) and add resilient throttling/retry logic.
