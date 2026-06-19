//go:build integration

package reddit

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestFetchLive(t *testing.T) {
	if os.Getenv("REDDIT_LIVE") == "" {
		t.Skip("set REDDIT_LIVE=1 to hit reddit")
	}

	c := New(WithUserAgent("go:dsr-ingest:0.1 (by /u/baalghorn)"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	posts, next, err := c.Fetch(ctx, FetchOpts{Subreddit: "EVE"}, "")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(posts) == 0 {
		t.Fatal("want posts, got none")
	}
	if next == "" {
		t.Fatal("want a next cursor")
	}

	withBody := 0
	for _, p := range posts {
		if p.Body != "" {
			withBody++
		}
	}
	t.Logf("ok: %d posts, %d with body, next=%s, first=%q", len(posts), withBody, next, posts[0].Title)
}
