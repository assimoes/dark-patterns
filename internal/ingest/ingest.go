// Package ingest maps scraped items into the db upsert rows for an artifact and its detail tables.
package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// Item is one scraped post or review, normalised so every source writes through the same path. Meta
// holds the source-specific fields that ride in source_meta. ImageURI is empty for a text-only item.
type Item struct {
	SourceID  string
	Body      string
	Lang      string
	Meta      map[string]any
	ImageURI  string
	MimeType  string
	ScrapedAt time.Time
}

// WriteItem upserts one item: the artifact, its text detail, and an image detail when the item carries
// an image. modality follows the channels present, text or multimodal.
func WriteItem(ctx context.Context, q *db.Queries, source string, gameID int32, it Item) error {
	modality := "text"
	if it.ImageURI != "" {
		modality = "multimodal"
	}

	artifactID, err := q.UpsertArtifact(ctx, db.UpsertArtifactParams{
		Modality:       modality,
		Source:         source,
		SourceID:       &it.SourceID,
		ContentHash:    hash(it.SourceID, it.Body),
		ScrapedAt:      pgtype.Timestamptz{Time: it.ScrapedAt, Valid: true},
		ExternalGameID: gameID,
	})
	if err != nil {
		return err
	}

	meta, _ := json.Marshal(it.Meta)
	if err := q.UpsertTextReviewDetail(ctx, db.UpsertTextReviewDetailParams{
		ArtifactID: artifactID,
		Body:       it.Body,
		Lang:       it.Lang,
		SourceMeta: meta,
	}); err != nil {
		return err
	}

	if it.ImageURI != "" {
		p := db.UpsertImageDetailParams{ArtifactID: artifactID, ImageUri: it.ImageURI}
		if it.MimeType != "" {
			p.MimeType = &it.MimeType
		}
		if err := q.UpsertImageDetail(ctx, p); err != nil {
			return err
		}
	}

	return nil
}

// hash is the dedupe key for content_hash: source id and body with a null byte between so "ab"+"c"
// and "a"+"bc" dont collide.
func hash(sourceID, body string) []byte {
	h := sha256.New()
	h.Write([]byte(sourceID))
	h.Write([]byte{0})
	h.Write([]byte(body))
	return h.Sum(nil)
}
