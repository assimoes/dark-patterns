// Command api is the http server backing the adjudication frontend, with graceful shutdown.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assimoes/dsr/internal/api"
	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

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

	// Decisions are attributed to a kind=human annotator, same as the adjudication cli.
	adjLabel := getenv("ADJUDICATOR_LABEL", "author")

	auditor, err := db.New(pool).GetAnnotatorByLabel(ctx, adjLabel)
	if err != nil || auditor.Kind != "human" {
		logger.Error("adjudicator must be an existing kind=human annotator", "label", adjLabel, "err", err)
		os.Exit(1)
	}

	sameSite := http.SameSiteLaxMode
	if getenv("SESSION_SAMESITE", "lax") == "none" {
		sameSite = http.SameSiteNoneMode
	}

	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		logger.Error("river client", "err", err)
		os.Exit(1)
	}

	origin := getenv("FRONTEND_ORIGIN", "http://localhost:3000")
	server := api.NewServer(pool, auditor.ID, origin, sameSite, riverClient, logger)

	addr := ":" + getenv("PORT", "8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// listen in the background so main can block on the shutdown signal.
	go func() {
		logger.Info("api listening", "addr", addr, "origin", origin)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve", "err", err)
			stop()
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}

	logger.Info("stopped")
}

// getenv reads an env var, falling back to def when it is unset or empty.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}
