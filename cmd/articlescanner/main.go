package main

import (
	"context"
	"os"

	"ArticlesScanner/internal/app"
	"ArticlesScanner/internal/config"
	"ArticlesScanner/internal/logging"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	logger := logging.New(cfg.Logging.Level)

	application := app.New(cfg, logger)

	if err := application.Run(ctx); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
