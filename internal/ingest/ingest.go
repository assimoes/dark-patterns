// Package ingest maps a steam review into the db upsert params, artifact row and the text detail.
package ingest

import (
	"encoding/json"
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/steam"
	"github.com/jackc/pgx/v5/pgtype"
)

// ArtifactParams builds the artifact upsert row for a review. content hash dedupes re-scrapes.
func ArtifactParams(gameID int32, r steam.Review, scrapedAt time.Time) db.UpsertArtifactParams {
	return db.UpsertArtifactParams{
		Modality:       "text",
		Source:         "steam",
		SourceID:       ptr(r.RecommendationID),
		ContentHash:    hash(r),
		ScrapedAt:      pgtype.Timestamptz{Time: scrapedAt, Valid: true},
		ExternalGameID: gameID,
	}
}

// TextReviewParams builds the text detail row for a review. body and lang are columns; the steam fields
// (voted_up, hours, score) live in source_meta. playtime comes in minutes, divide by 60 for hours.
func TextReviewParams(artifactID int64, r steam.Review) db.UpsertTextReviewDetailParams {
	meta, _ := json.Marshal(map[string]any{
		"voted_up":            r.VotedUp,
		"hours_played":        int32(r.Author.PlaytimeForever / 60),
		"weighted_vote_score": string(r.WeightedVotedScore),
	})

	return db.UpsertTextReviewDetailParams{
		ArtifactID: artifactID,
		Body:       r.Review,
		Lang:       r.Language,
		SourceMeta: meta,
	}
}
