@echo off
setlocal ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

if "%ARTICLE_SCANNER_CONFIG%"=="" set ARTICLE_SCANNER_CONFIG=.\config.yaml
if "%DATABASE_DSN%"=="" set DATABASE_DSN=postgres://user:pass@localhost:5432/articles_scanner
if "%CHATGPT_API_KEY%"=="" set CHATGPT_API_KEY=sk-your-key
if "%CHATGPT_MODEL%"=="" set CHATGPT_MODEL=gpt-4o-mini
if "%TELEGRAM_BOT_TOKEN%"=="" set TELEGRAM_BOT_TOKEN=
if "%TELEGRAM_CHAT_ID%"=="" set TELEGRAM_CHAT_ID=

go run ./cmd/articlescanner %*
