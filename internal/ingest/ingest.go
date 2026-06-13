// Package ingest maps a steam review into the db upsert params, artifact row and the text detail.
package ingest

import (
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

// TextReviewParams builds the text detail row tied to an artifact. playtime comes in minutes so
// divide by 60 for hours, and a bad weighted score just falls back to empty numeric.
func TextReviewParams(artifactID int64, r steam.Review) db.UpsertTextReviewDetailParams {
	weightedScore, err := toNumeric(string(r.WeightedVotedScore))
	if err != nil {
		weightedScore = pgtype.Numeric{}
	}

	return db.UpsertTextReviewDetailParams{
		ArtifactID:        artifactID,
		Body:              r.Review,
		VotedUp:           r.VotedUp,
		HoursPlayed:       int32(r.Author.PlaytimeForever / 60),
		Lang:              r.Language,
		WeightedVoteScore: weightedScore,
	}
}
