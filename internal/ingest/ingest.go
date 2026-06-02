package ingest

import (
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/steam"
	"github.com/jackc/pgx/v5/pgtype"
)

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
