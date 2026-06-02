package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assimoes/dsr/internal/config"
	"github.com/assimoes/dsr/internal/scrape"
	"github.com/assimoes/dsr/internal/steam"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if len(os.Args) < 2 {
		logger.Error("usage: steam-worker <enqueue|serve> [flags]")
		os.Exit(2)
	}

	mode := os.Args[1]

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL not set")
		os.Exit(2)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("connect to postgres failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch mode {
	case "enqueue":
		enqueue(ctx, logger, pool)
	case "serve":
		serve(ctx, logger, pool)
	default:
		logger.Error("unknown mode", "mode", mode)
		os.Exit(2)
	}
}

func serve(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	delay := fs.Duration("delay", 500*time.Millisecond, "delay between pages")
	_ = fs.Parse(os.Args[2:])

	sc := steam.New(
		steam.WithUserAgent("dsr/1.0 (asvns@iscte-iul.pt)"),
		steam.WithLogger(logger),
	)

	workers := river.NewWorkers()

	river.AddWorker(workers, scrape.NewScrapeWorker(pool, sc, *delay, logger))

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:  map[string]river.QueueConfig{"scrape": {MaxWorkers: 1}},
		Workers: workers,
		Logger:  logger,
	})

	if err != nil {
		logger.Error("river client", "err", err)
		os.Exit(1)
	}

	if err := client.Start(ctx); err != nil {
		logger.Error("start", "err", err)
		os.Exit(1)
	}
	logger.Info("serving scrape queue; CTRL+C to stop")

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Stop(stopCtx); err != nil {
		logger.Error("stop", "err", err)
	}

	logger.Info("stopped; committed pages are svaed, restart serve to resume")
}

func enqueue(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) {
	cfg, err := config.Load(os.Args[2:], os.Getenv)
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(2)
	}

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		logger.Error("river client", "err", err)
		os.Exit(1)
	}

	if err := scrape.Enqueue(ctx, client, pool, cfg.GameID, cfg.Filter, cfg.Language, cfg.MaxReviews); err != nil {
		logger.Error("enqueue scrape", "err", err)
		os.Exit(2)
	}
	logger.Info("enqueue scrape", "app", cfg.AppID, "filter", cfg.Filter, "lang", cfg.Language, "max", cfg.MaxReviews)
}
