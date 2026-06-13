//go:build integration

package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func seedReview(
	t *testing.T, ctx context.Context, q *Queries,
	srcID string, game int32, scrapedAt time.Time, hours int32, body string) int64 {

	t.Helper()
	id, err := q.UpsertArtifact(ctx, UpsertArtifactParams{
		Modality:       "text",
		Source:         "steam",
		SourceID:       &srcID,
		ContentHash:    []byte("hash-" + srcID),
		ScrapedAt:      pgtype.Timestamptz{Time: scrapedAt, Valid: true},
		ExternalGameID: game,
	})
	if err != nil {
		t.Fatalf("seed artifact %s: %v", srcID, err)
	}

	if err := q.UpsertTextReviewDetail(ctx, UpsertTextReviewDetailParams{
		ArtifactID:        id,
		Body:              body,
		VotedUp:           true,
		HoursPlayed:       hours,
		Lang:              "english",
		WeightedVoteScore: pgtype.Numeric{},
	}); err != nil {
		t.Fatalf("seed review %s: %v", srcID, err)
	}

	return id
}

func TestStratifiedPopulation(t *testing.T) {
	ctx := context.Background()
	tx, err := testPool(t).Begin(ctx)

	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx)

	q := New(tx)

	// freeze scans the whole artifacts table, so clear it first — this test should
	// only see its own seeds. Rolled back with the tx.
	if _, err := tx.Exec(ctx, "TRUNCATE artifacts CASCADE"); err != nil {
		t.Fatalf("truncate artifacts: %v", err)
	}

	cutoff := time.Now()
	// eligible
	before := cutoff.Add(-time.Hour)
	// exclude
	after := cutoff.Add(time.Hour)

	longBody := "this games gates progress behind a paywall. Classic pay to win with very predatory monetisation indeed"

	seedReview(t, ctx, q, "c-100-a", 100, before, 5, longBody)
	seedReview(t, ctx, q, "c-100-b", 100, before, 5, longBody)
	seedReview(t, ctx, q, "c-100-c", 100, before, 5, longBody)

	seedReview(t, ctx, q, "c-200-a", 200, before, 5, longBody)

	seedReview(t, ctx, q, "c-late", 200, after, 5, longBody)

	seedReview(t, ctx, q, "c-no-hours", 200, before, 0, longBody)

	popID, err := q.CreatePopulation(ctx, CreatePopulationParams{
		Modality:    "text",
		Description: ptr("test population"),
		Criteria: []byte(`
			{"min_hours_played": 1, "per_game_cap": 2}
		`),
		ArtifactsCutoff: pgtype.Timestamptz{Time: cutoff, Valid: true},
	})

	if err != nil {
		t.Fatalf("create population: %v", err)
	}

	inserted, err := q.FreezeStratifiedPopulation(ctx, FreezeStratifiedPopulationParams{
		PopulationID:    popID,
		ArtifactsCutoff: pgtype.Timestamptz{Time: cutoff, Valid: true},
		MinHoursPlayed:  1,
		PerGameCap:      2,
	})
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if inserted != 3 {
		t.Fatalf("freeze should insert 3 (cap 2 of game100 + 1 of game200), got %d", inserted)
	}

	count, err := q.CountIndividuals(ctx, popID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}

	if count != 3 {
		t.Fatalf("population should hold 3 individuals, got %d", count)
	}

	// refreeze should be idempotent

	again, err := q.FreezeStratifiedPopulation(ctx, FreezeStratifiedPopulationParams{
		PopulationID:    popID,
		ArtifactsCutoff: pgtype.Timestamptz{Time: cutoff, Valid: true},
		MinHoursPlayed:  1,
		PerGameCap:      2,
	})

	if err != nil {
		t.Fatalf("refreeze: %v", err)
	}

	if again != 0 {
		t.Fatalf("refreezing should insert 0, got %d", again)
	}

	total, _ := q.CountIndividuals(ctx, popID)

	if total != 3 {
		t.Fatalf("refreezing must not change the population total: was 3, now %d", total)
	}
}

func ptr(s string) *string { return &s }
