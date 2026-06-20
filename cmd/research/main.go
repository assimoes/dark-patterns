// Command research serves the worker that drains the game-research queue: one web-search-grounded
// OpenRouter call per game, stored as a draft description for human review. enqueuing happens from the
// website when a game is added; this binary only does the background work.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/research"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

const defaultResearchModel = "google/gemini-3-flash-preview"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if len(os.Args) < 2 || os.Args[1] != "serve" {
		logger.Error("usage: research-worker serve")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL not set")
		os.Exit(2)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		logger.Error("OPENROUTER_API_KEY not set")
		os.Exit(2)
	}

	model := os.Getenv("RESEARCH_MODEL")
	if model == "" {
		model = defaultResearchModel
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("connect to postgres failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// the research model must sit outside the annotation panel: panel models never self-research.
	if err := assertNotPanelModel(ctx, pool, model); err != nil {
		logger.Error("research model check", "model", model, "err", err)
		os.Exit(1)
	}

	serve(ctx, logger, pool, apiKey, model)
}

// assertNotPanelModel fails if the research model slug is an active panel model.
func assertNotPanelModel(ctx context.Context, pool *pgxpool.Pool, model string) error {
	models, err := db.New(pool).ListActiveModels(ctx)
	if err != nil {
		return err
	}
	for _, m := range models {
		if m.Slug == model {
			return errors.New("research model is an active panel model; it must be outside the panel")
		}
	}
	return nil
}

func serve(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool, apiKey, model string) {
	client := research.NewClient(apiKey, model,
		research.WithAttribution("https://github.com/assimoes/dsr", "dsr-research"),
	)

	workers := river.NewWorkers()
	river.AddWorker(workers, research.NewWorker(pool, client, logger))

	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"research": {MaxWorkers: 2},
		},
		Workers: workers,
		Logger:  logger,
	})
	if err != nil {
		logger.Error("river client", "err", err)
		os.Exit(1)
	}

	if err := riverClient.Start(ctx); err != nil {
		logger.Error("start", "err", err)
		os.Exit(1)
	}
	logger.Info("serving research queue", "model", model)

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := riverClient.Stop(stopCtx); err != nil {
		logger.Error("stop", "err", err)
	}

	logger.Info("stopped")
}
