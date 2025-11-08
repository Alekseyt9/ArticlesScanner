package main

import (
	"context"
	"log"

	"ArticlesScanner/internal/app"
	"ArticlesScanner/internal/config"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	application := app.New(cfg)

	if err := application.Run(ctx); err != nil {
		log.Fatalf("application stopped: %v", err)
	}
}
