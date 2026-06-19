package ingest

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/assimoes/dsr/internal/steam"
)

func TestToNumeric(t *testing.T) {

	n, err := toNumeric("")
	if err != nil {
		t.Fatalf("empty: unexpected error %v", err)
	}

	if n.Valid {
		t.Fatal("empty: want invalid (NULL), got valid")
	}

	n, err = toNumeric("0.5")
	if err != nil {
		t.Fatalf("0.5: unexpected error %v", err)
	}

	if !n.Valid {
		t.Fatal("0.5: want a valid numeric")
	}

	if _, err := toNumeric("not-a-number"); err == nil {
		t.Fatal("nan: want an error, got nil")
	}
}

func TestHash(t *testing.T) {
	a := steam.Review{RecommendationID: "1", Review: "great game"}
	same := steam.Review{RecommendationID: "1", Review: "great game"}
	sameBody := steam.Review{RecommendationID: "2", Review: "great game"}

	if !bytes.Equal(hash(a), hash(same)) {
		t.Fatal("identical reviews should hash the same")
	}

	if bytes.Equal(hash(a), hash(sameBody)) {
		t.Fatal("different recommendation id should change the hash even with an identical body")
	}

	if len(hash(a)) != 32 {
		t.Fatalf("want a 32-byte sha256, got %d", len(hash(a)))
	}
}

func TestArtifactParams(t *testing.T) {
	when := time.Unix(1700000000, 0)

	p := ArtifactParams(730, steam.Review{RecommendationID: "abc"}, when)

	if p.Modality != "text" || p.Source != "steam" {
		t.Fatalf("modality/source mismatch: %+v", p)
	}

	if p.SourceID == nil || *p.SourceID != "abc" {
		t.Fatalf("source id: %v", p.SourceID)
	}

	if p.ExternalGameID != 730 {
		t.Fatalf("game id: want 730, got %d", p.ExternalGameID)
	}

	if !p.ScrapedAt.Valid || !p.ScrapedAt.Time.Equal(when) {
		t.Fatalf("scrapt at: %v", p.ScrapedAt)
	}

	if len(p.ContentHash) != 32 {
		t.Fatalf("content hash: want 32 bytes, got %d", len(p.ContentHash))
	}
}

func TestTextReviewParams(t *testing.T) {
	r := steam.Review{
		Review:             "great",
		VotedUp:            true,
		Language:           "english",
		WeightedVotedScore: "0.6",
		Author:             steam.Author{PlaytimeForever: 150},
	}

	p := TextReviewParams(42, r)

	if p.ArtifactID != 42 {
		t.Fatalf("artifact id: want 42, got %d", p.ArtifactID)
	}

	if p.Body != "great" || p.Lang != "english" {
		t.Fatalf("field mapping mismatch: %v", p)
	}

	var meta map[string]any
	if err := json.Unmarshal(p.SourceMeta, &meta); err != nil {
		t.Fatalf("source_meta should be valid json: %v", err)
	}
	if meta["voted_up"] != true {
		t.Fatalf("voted_up: want true, got %v", meta["voted_up"])
	}
	if meta["hours_played"].(float64) != 2 {
		t.Fatalf("hours: want 2 (truncated from 150 minutes), got %v", meta["hours_played"])
	}
	if meta["weighted_vote_score"] != "0.6" {
		t.Fatalf("weighted_vote_score: want \"0.6\", got %v", meta["weighted_vote_score"])
	}

	empty := TextReviewParams(1, steam.Review{Author: steam.Author{PlaytimeForever: 59}})

	var emptyMeta map[string]any
	if err := json.Unmarshal(empty.SourceMeta, &emptyMeta); err != nil {
		t.Fatalf("source_meta should be valid json: %v", err)
	}
	if emptyMeta["hours_played"].(float64) != 0 {
		t.Fatalf("59 minutes should truncate to 0 hours, got %v", emptyMeta["hours_played"])
	}
	if emptyMeta["weighted_vote_score"] != "" {
		t.Fatalf("empty score should store as empty string, got %v", emptyMeta["weighted_vote_score"])
	}
}
