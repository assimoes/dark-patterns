// Command scrape serves the worker that drains every source's scrape queue, pulling pages until done.
// enqueuing is done from the website, not here. the cli only does background work.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assimoes/dsr/internal/reddit"
	"github.com/assimoes/dsr/internal/scrape"
	"github.com/assimoes/dsr/internal/steam"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if len(os.Args) < 2 || os.Args[1] != "serve" {
		logger.Error("usage: scrape-worker serve [flags]")
		os.Exit(2)
	}

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

	serve(ctx, logger, pool)
}

// serve registers every source on its own queue and blocks until a shutdown signal. steam and reddit
// are paged, so one worker each, dont parallelize.
func serve(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	delay := fs.Duration("delay", time.Second, "delay between pages")
	_ = fs.Parse(os.Args[2:])

	sc := steam.New(
		steam.WithUserAgent("dsr/1.0 (asvns@iscte-iul.pt)"),
		steam.WithLogger(logger),
	)
	rc := reddit.New(
		reddit.WithUserAgent("go:dsr-ingest:0.1 (by /u/baalghorn)"),
		reddit.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
		reddit.WithLogger(logger),
	)

	sources := map[string]scrape.Source{
		"steam":  scrape.NewSteamSource(sc),
		"reddit": scrape.NewRedditSource(rc),
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, scrape.NewWorker(pool, sources, *delay, logger))

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"steam":  {MaxWorkers: 1},
			"reddit": {MaxWorkers: 1},
		},
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
	logger.Info("serving scrape queues; CTRL+C to stop")

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Stop(stopCtx); err != nil {
		logger.Error("stop", "err", err)
	}

	logger.Info("stopped; committed pages are saved, restart serve to resume")
}
