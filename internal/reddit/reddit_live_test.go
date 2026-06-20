//go:build integration

package reddit

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestFetchLive(t *testing.T) {
	if os.Getenv("REDDIT_LIVE") == "" {
		t.Skip("set REDDIT_LIVE=1 to hit reddit")
	}

	c := New(WithUserAgent("go:dsr-ingest:0.1 (by /u/baalghorn)"), WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	after := ""
	for page := 1; page <= 2; page++ {
		start := time.Now()
		posts, next, err := c.Fetch(ctx, FetchOpts{Subreddit: "EVE"}, after)
		if err != nil {
			t.Fatalf("page %d fetch: %v", page, err)
		}
		if len(posts) == 0 || next == "" {
			t.Fatalf("page %d: %d posts, next=%q", page, len(posts), next)
		}
		t.Logf("page %d: %d posts in %s, next=%s", page, len(posts), time.Since(start).Round(time.Second), next)
		after = next
	}
}
