// Package scrape runs the river worker that pulls a source one page at a time, writes the items, then
// enqueues the next page until done. it holds a registry of sources and never names a concrete client.
package scrape

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/ingest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// Args is one page of work: the source and its target, where we are (cursor), how many items still
// wanted, and any source-specific knobs.
type Args struct {
	Source    string            `json:"source"`
	GameID    int32             `json:"game_id"`
	Target    string            `json:"target"`
	Cursor    string            `json:"cursor"`
	Remaining int               `json:"remaining"`
	Params    map[string]string `json:"params"`
}

// Kind is the river job kind.
func (Args) Kind() string {
	return "scrape_page"
}

// InsertOpts pins each job to its source's queue and dedupes by args so the same cursor wont double up.
func (a Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: a.Source,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

// Worker runs one page per job and self-enqueues the next, dispatching to a source by name.
type Worker struct {
	river.WorkerDefaults[Args]
	pool      *pgxpool.Pool
	sources   map[string]Source
	pageDelay time.Duration
	logger    *slog.Logger
}

// NewWorker wires up the worker over a registry of sources. nil logger falls back to slog.Default.
func NewWorker(pool *pgxpool.Pool, sources map[string]Source, pageDelay time.Duration, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}

	return &Worker{
		pool:      pool,
		sources:   sources,
		pageDelay: pageDelay,
		logger:    logger,
	}
}

// Work fetches one page, writes it and the cursor in a tx, then enqueues the next page unless done (no
// next cursor, repeated cursor, empty page, or hit the cap). the page write and next-job insert share
// the tx so a crash never loses or skips a page.
func (w *Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	a := job.Args

	src, ok := w.sources[a.Source]
	if !ok {
		return fmt.Errorf("no source registered for %q", a.Source)
	}

	items, next, err := src.Fetch(ctx, a.Target, a.Cursor, a.Params)
	if err != nil {
		return err
	}

	kept := items
	if a.Remaining > 0 && len(kept) > a.Remaining {
		kept = kept[:a.Remaining]
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	now := time.Now()
	for _, it := range kept {
		it.ScrapedAt = now
		if err := ingest.WriteItem(ctx, q, a.Source, a.GameID, it); err != nil {
			return err
		}
	}

	if next != "" {
		if err := saveCursor(ctx, q, a, next); err != nil {
			return err
		}
	}

	remaining := a.Remaining
	if remaining > 0 {
		remaining -= len(kept)
	}

	done := next == "" || next == a.Cursor || len(items) == 0 || (a.Remaining > 0 && remaining <= 0)

	if !done {
		client := river.ClientFromContext[pgx.Tx](ctx)

		nextArgs := a
		nextArgs.Cursor = next
		nextArgs.Remaining = remaining

		if _, err := client.InsertTx(ctx, tx, nextArgs, &river.InsertOpts{
			Queue:       a.Source,
			ScheduledAt: time.Now().Add(w.pageDelay),
		}); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	w.logger.Info("scraped page", "source", a.Source, "game", a.GameID, "items", len(kept), "more", !done)
	return nil
}

// Enqueue kicks off a scrape by inserting the first page job. it resumes from the saved cursor if
// theres one. max is how many items to pull (0 = all).
func Enqueue(ctx context.Context, client *river.Client[pgx.Tx], pool *pgxpool.Pool,
	source string, gameID int32, target string, params map[string]string, max int) error {

	cursor, err := loadCursor(ctx, pool, source, gameID, params)
	if err != nil {
		return err
	}

	_, err = client.Insert(ctx, Args{
		Source:    source,
		GameID:    gameID,
		Target:    target,
		Cursor:    cursor,
		Remaining: max,
		Params:    params,
	}, nil)

	return err
}

// cursorKey splits the per-source cursor knobs. steam pages per filter and language; reddit pages by
// subreddit alone, so its variant is empty.
func cursorKey(params map[string]string) (filter, language string) {
	return params["filter"], params["language"]
}

func saveCursor(ctx context.Context, q *db.Queries, a Args, next string) error {
	filter, language := cursorKey(a.Params)
	return q.UpsertScrapeCursor(ctx, db.UpsertScrapeCursorParams{
		ExternalGameID: a.GameID,
		Source:         a.Source,
		Filter:         filter,
		Language:       language,
		Cursor:         next,
	})
}

// loadCursor returns the saved cursor for this game/source, or empty (start from the top) when theres
// nothing saved yet.
func loadCursor(ctx context.Context, pool *pgxpool.Pool, source string, gameID int32, params map[string]string) (string, error) {
	filter, language := cursorKey(params)
	saved, err := db.New(pool).GetScrapeCursor(ctx, db.GetScrapeCursorParams{
		ExternalGameID: gameID,
		Source:         source,
		Filter:         filter,
		Language:       language,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", nil
	case err != nil:
		return "", err
	default:
		return saved, nil
	}
}
