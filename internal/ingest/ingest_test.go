package ingest

import (
	"bytes"
	"testing"

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
	a := hash("1", "great game")
	same := hash("1", "great game")
	sameBody := hash("2", "great game")

	if !bytes.Equal(a, same) {
		t.Fatal("identical input should hash the same")
	}

	if bytes.Equal(a, sameBody) {
		t.Fatal("a different source id should change the hash even with an identical body")
	}

	if len(a) != 32 {
		t.Fatalf("want a 32-byte sha256, got %d", len(a))
	}
}

func TestSteamItem(t *testing.T) {
	r := steam.Review{
		Review:             "great",
		VotedUp:            true,
		Language:           "english",
		WeightedVotedScore: "0.6",
		Author:             steam.Author{PlaytimeForever: 150},
	}

	it := SteamItem(r)

	if it.SourceID != r.RecommendationID || it.Body != "great" || it.Lang != "english" {
		t.Fatalf("field mapping mismatch: %+v", it)
	}

	if it.Meta["voted_up"] != true {
		t.Fatalf("voted_up: want true, got %v", it.Meta["voted_up"])
	}
	if it.Meta["hours_played"].(int32) != 2 {
		t.Fatalf("hours: want 2 (from 150 minutes), got %v", it.Meta["hours_played"])
	}
	if it.Meta["weighted_vote_score"] != "0.6" {
		t.Fatalf("weighted_vote_score: want 0.6, got %v", it.Meta["weighted_vote_score"])
	}

	if it.ImageURI != "" {
		t.Fatalf("a steam item carries no image, got %q", it.ImageURI)
	}
}
