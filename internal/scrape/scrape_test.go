//go:build integration

package scrape

import (
	"context"
	"os"
	"testing"

	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/steam"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testGameID = int32(99900001)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")

	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run scrape integration tests")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	t.Cleanup(pool.Close)

	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = pool.Exec(ctx,
			`DELETE FROM text_review_details td USING artifacts a WHERE td.artifact_id = a.id AND a.external_game_id = $1`,
			testGameID)

		_, _ = pool.Exec(ctx, `DELETE from artifacts WHERE external_game_id = $1`, testGameID)
		_, _ = pool.Exec(ctx, `DELETE FROM scrape_cursors WHERE external_game_id = $1`, testGameID)
	})

	return pool
}

func review(id, body string) steam.Review {
	r := steam.Review{RecommendationID: id, Review: body, VotedUp: true, Language: "english"}
	r.Author.PlaytimeForever = 600
	return r
}

func TestWritePageStoresRowsAndCheckpoints(t *testing.T) {

	pool := testPool(t)
	ctx := context.Background()

	tx, err := pool.Begin(ctx)

	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	page := []steam.Review{
		review("rec-1", "pay to win garbage"),
		review("rec-2", "grind wall"),
	}

	if err := writePage(ctx, q, testGameID, "recent", "english", page, "cursor-page-2"); err != nil {
		t.Fatalf("writePage: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	var artifacts, details int

	_ = pool.QueryRow(ctx, `SELECT count(*) FROM artifacts WHERE external_game_id = $1`, testGameID).
		Scan(&artifacts)

	_ = pool.QueryRow(ctx, `SELECT count(*) FROM text_review_details td JOIN artifacts a ON a.id = td.artifact_id WHERE a.external_game_id = $1`, testGameID).
		Scan(&details)

	if artifacts != 2 || details != 2 {
		t.Fatalf("rows: want 2/2, got %d/%d", artifacts, details)
	}

	cur, err := db.New(pool).GetScrapeCursor(ctx, db.GetScrapeCursorParams{
		ExternalGameID: testGameID, Filter: "recent", Language: "english",
	})
	if err != nil {
		t.Fatalf("get cursor: %v", err)
	}

	if cur != "cursor-page-2" {
		t.Fatalf("checkpoint: want cursor-page-2, got %q", cur)
	}
}
