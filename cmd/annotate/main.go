package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assimoes/dsr/internal/annotate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

var panelSlugs = []string{
	"openai/gpt-4o-mini",
	"anthropic/claude-3.5-haiku",
	"google/gemini-2.5-flash-lite",
	"deepseek/deepseek-chat-v3.1",
}

func buildRegistry(dryRun bool) (map[string]annotate.Annotator, error) {
	if dryRun {
		return map[string]annotate.Annotator{
			"test-model": annotate.FakeAnnotator{
				Response: json.RawMessage(
					`{"patterns": [{"code": "PM-1", "evidence": "play to win", "explanation": "buys power"}]}`,
				),
				ID: annotate.RunIdentity{Provider: "test", Model: "test-model", ClientVersion: "0.0.1"},
			},
		}, nil
	}

	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		return nil, errors.New("OPENROUTER_API_KEY not set (pass -dry to run the offline dry run panel)")
	}

	httpClient := &http.Client{Timeout: 90 * time.Second}

	registry := make(map[string]annotate.Annotator, len(panelSlugs))

	for _, slug := range panelSlugs {
		registry[slug] = annotate.NewOpenRouterAnnotator(
			key,
			slug,
			annotate.WithHTTPClient(httpClient),
			annotate.WithClientVersion("openrouter/v1"),
			annotate.WithAttribution("https://github.com/assimoes/dark-patterns", "design science research artifact"),
		)
	}

	return registry, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if len(os.Args) < 2 {
		logger.Error("usage: annotate <enqueue | serve> [flags]")
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
		logger.Error("connect to database failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch mode {
	case "enqueue":
		fs := flag.NewFlagSet("enqueue", flag.ExitOnError)
		runID := fs.Int("run", 0, "run id to annotate")
		dry := fs.Bool("dry", false, "use the offline dry run panel instead of openrouter")

		_ = fs.Parse(os.Args[2:])

		if *runID == 0 {
			logger.Error("enqueue required -run <id>")
			os.Exit(2)
		}

		registry, err := buildRegistry(*dry)
		if err != nil {
			logger.Error("build registry", "err", err)
			os.Exit(1)
		}

		client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
		if err != nil {
			logger.Error("river client", "err", err)
			os.Exit(1)
		}

		n, err := annotate.Enqueue(ctx, client, pool, int32(*runID), registry)
		if err != nil {
			logger.Error("enqueue annotate", "err", err)
			os.Exit(1)
		}

		logger.Info("enqueued annotation jobs", "run", *runID, "jobs", n)

	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		dry := fs.Bool("dry", false, "use the offline dry run panel instead of openrouter")
		_ = fs.Parse(os.Args[2:])

		registry, err := buildRegistry(*dry)
		if err != nil {
			logger.Error("build registry", "err", err)
			os.Exit(1)
		}

		loaders := []annotate.Loader{annotate.TextLoader{}}
		workers := river.NewWorkers()
		river.AddWorker(workers, annotate.NewAnnotateWorker(pool, loaders, registry, logger))

		client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
			Queues: map[string]river.QueueConfig{
				"annotate": {MaxWorkers: 10},
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

		logger.Info("serving annotate queue; CTRL+C to stop")

		<-ctx.Done()

		stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := client.Stop(stopCtx); err != nil {
			logger.Error("stop", "err", err)
		}

		logger.Info("stopped; completed annotations are saved, restart serve to resume")

	default:
		logger.Error("unknown mode", "mode", mode)
		os.Exit(2)
	}
}
