package app

import (
	"context"
	"log/slog"
	"time"

	"ArticlesScanner/internal/config"
	"ArticlesScanner/internal/infrastructure/llm"
	"ArticlesScanner/internal/infrastructure/parser"
	"ArticlesScanner/internal/logging"
	"ArticlesScanner/internal/ports"
	"ArticlesScanner/internal/scanner"
	"ArticlesScanner/internal/usecase"
)

// Application wires configs to use cases and lifecycle orchestration.
type Application struct {
	cfg      config.Config
	pipeline *usecase.Pipeline
}

// New builds a minimal runnable application instance.
func New(cfg config.Config, baseLogger *slog.Logger) *Application {
	if baseLogger == nil {
		baseLogger = logging.New(cfg.Logging.Level)
	}

	registry := scanner.NewRegistry()
	registry.Register(parser.NewArxivScanner(nil, baseLogger.With("component", "scanner.arxiv")))

	source := parser.NewStrategySource(registry, cfg.Sites, baseLogger.With("component", "source"))

	var chatClient ports.ChatClient
	if cfg.ChatGPT.APIKey != "" {
		chatClient = llm.NewChatGPTClient(cfg.ChatGPT)
	}

	pipeline := usecase.NewPipeline(usecase.PipelineDeps{
		Source:     source,
		ChatClient: chatClient,
		Logger:     baseLogger.With("component", "pipeline"),
	})
	return &Application{cfg: cfg, pipeline: pipeline}
}

// Run performs a single pipeline execution placeholder; later plug scheduler.
func (a *Application) Run(ctx context.Context) error {
	if a.pipeline == nil {
		return nil
	}

	now := time.Now().In(a.cfg.Scheduler.Location())
	return a.pipeline.ProcessDay(ctx, now)
}
