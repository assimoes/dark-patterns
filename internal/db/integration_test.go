//go:build integration

package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run db integration tests")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

func TestUpsertArtifactIsIdempotent(t *testing.T) {
	ctx := context.Background()

	tx, err := testPool(t).Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx)

	q := New(tx)

	sourceID := "test-rec-1"

	args := UpsertArtifactParams{
		Modality:       "text",
		Source:         "steam",
		SourceID:       &sourceID,
		ContentHash:    []byte("hash-1"),
		ScrapedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ExternalGameID: 730,
	}

	id1, err := q.UpsertArtifact(ctx, args)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	id2, err := q.UpsertArtifact(ctx, args)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	if id1 != id2 {
		t.Fatalf("rescrapping should return the same id, got %d vs %d", id1, id2)
	}

}
