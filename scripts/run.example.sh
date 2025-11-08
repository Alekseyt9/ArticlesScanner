#!/usr/bin/env bash
set -euo pipefail

export ARTICLE_SCANNER_CONFIG=${ARTICLE_SCANNER_CONFIG:-"./configs/config.yaml"}
export DATABASE_DSN=${DATABASE_DSN:-"postgres://user:pass@localhost:5432/articles"}
export CHATGPT_API_KEY=${CHATGPT_API_KEY:-"sk-your-key"}
export CHATGPT_MODEL=${CHATGPT_MODEL:-"gpt-4o-mini"}
export TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-""}
export TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID:-""}

exec go run ./cmd/articlescanner "$@"
