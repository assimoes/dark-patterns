// Package scrape is the river worker that pulls steam review pages one cursor at a time,
// writes them, then enqueues the next page until done.
package scrape

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/ingest"
	"github.com/assimoes/dsr/internal/steam"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// ScrapeArgs is one page of work: where we are (cursor) and how many reviews still wanted.
type ScrapeArgs struct {
	GameID    int32  `json:"game_id"`
	Filter    string `json:"filter"`
	Language  string `json:"language"`
	Cursor    string `json:"cursor"`
	Remaining int    `json:"remaining"`
}

// Kind is the river job kind for these args.
func (ScrapeArgs) Kind() string {
	return "scrape_page"
}

// InsertOpts pins these to the scrape queue and dedupes by args so the same cursor wont double up.
func (ScrapeArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "scrape",
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

// ScrapeWorker runs one page per job and self-enqueues the next.
type ScrapeWorker struct {
	river.WorkerDefaults[ScrapeArgs]
	pool      *pgxpool.Pool
	steam     *steam.Client
	pageDelay time.Duration
	logger    *slog.Logger
}

// NewScrapeWorker wires up the worker. nil logger falls back to slog.Default.
func NewScrapeWorker(pool *pgxpool.Pool, sc *steam.Client, pageDelay time.Duration, logger *slog.Logger) *ScrapeWorker {
	if logger == nil {
		logger = slog.Default()
	}

	return &ScrapeWorker{
		pool:      pool,
		steam:     sc,
		pageDelay: pageDelay,
		logger:    logger,
	}
}

// Work fetches one page, writes it and the cursor in a tx, then enqueues the next page unless
// were done (no next cursor, repeated cursor, empty page, or hit the remaining cap). page write
// and next-job insert share the tx so a crash never loses or skips a page.
func (w *ScrapeWorker) Work(ctx context.Context, job *river.Job[ScrapeArgs]) error {
	a := job.Args

	cursor := a.Cursor
	if cursor == "" {
		cursor = "*"
	}

	opts := steam.FetchOpts{
		AppID:    strconv.Itoa(int(a.GameID)),
		Language: a.Language,
		Filter:   a.Filter,
	}

	reviews, next, err := w.steam.FetchOnce(ctx, opts, cursor)
	if err != nil {
		return err
	}

	kept := reviews
	if a.Remaining > 0 && len(kept) > a.Remaining {
		kept = kept[:a.Remaining]
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	if err := writePage(ctx, q, a.GameID, a.Filter, a.Language, kept, next); err != nil {
		return err
	}

	remaining := a.Remaining
	if remaining > 0 {
		remaining -= len(kept)
	}

	done := next == "" || next == cursor || len(reviews) == 0 || (a.Remaining > 0 && remaining <= 0)

	if !done {
		client := river.ClientFromContext[pgx.Tx](ctx)

		if _, err := client.InsertTx(ctx, tx, ScrapeArgs{
			GameID:    a.GameID,
			Filter:    a.Filter,
			Language:  a.Language,
			Cursor:    next,
			Remaining: remaining,
		}, &river.InsertOpts{
			Queue:       "scrape",
			ScheduledAt: time.Now().Add(w.pageDelay),
		}); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	w.logger.Info("scraped page", "game", a.GameID, "reviews", len(kept), "more", !done)
	return nil
}

// writePage upserts each review (artifact then text detail) and saves the next cursor so a
// resumed run picks up where this one stopped.
func writePage(ctx context.Context, q *db.Queries, gameID int32,
	filter, language string, reviews []steam.Review, next string) error {

	for _, r := range reviews {
		artifactID, err := q.UpsertArtifact(ctx, ingest.ArtifactParams(
			gameID,
			r,
			time.Now(),
		))

		if err != nil {
			return err
		}

		if err := q.UpsertTextReviewDetail(ctx, ingest.TextReviewParams(artifactID, r)); err != nil {
			return err
		}
	}

	if next == "" {
		return nil
	}

	return q.UpsertScrapeCursor(ctx, db.UpsertScrapeCursorParams{
		ExternalGameID: gameID,
		Filter:         filter,
		Language:       language,
		Cursor:         next,
	})
}

// Enqueue kicks off a scrape for a game by inserting the first page job. resumes from the saved
// cursor if theres one, max is how many reviews to pull (0 = all).
func Enqueue(ctx context.Context, client *river.Client[pgx.Tx],
	pool *pgxpool.Pool, game int32, filter, language string, max int) error {

	cursor, err := loadCursor(ctx, pool, game, filter, language)
	if err != nil {
		return err
	}

	_, err = client.Insert(ctx, ScrapeArgs{
		GameID:    game,
		Filter:    filter,
		Language:  language,
		Cursor:    cursor,
		Remaining: max,
	}, nil)

	return err
}

// loadCursor returns the saved cursor for this game/filter/language, or "*" (start from the top)
// when theres nothing saved yet.
func loadCursor(ctx context.Context, pool *pgxpool.Pool, game int32, filter, language string) (string, error) {
	saved, err := db.New(pool).GetScrapeCursor(ctx, db.GetScrapeCursorParams{
		ExternalGameID: game,
		Filter:         filter,
		Language:       language,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "*", nil
	case err != nil:
		return "", err
	case saved == "":
		return "*", nil
	default:
		return saved, nil
	}
}
