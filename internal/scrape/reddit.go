package scrape

import (
	"context"

	"github.com/assimoes/dsr/internal/ingest"
	"github.com/assimoes/dsr/internal/reddit"
)

type redditSource struct {
	client *reddit.Client
}

// NewRedditSource wraps a reddit client as a Source.
func NewRedditSource(client *reddit.Client) Source {
	return redditSource{client: client}
}

// Name is "reddit".
func (redditSource) Name() string {
	return "reddit"
}

// Fetch grabs one page of posts for a subreddit and maps them to items. an image post stores the image
// link; the panel downloads it at annotation time, so the row stays small and the scrape makes no extra
// requests.
func (s redditSource) Fetch(ctx context.Context, target, cursor string, params map[string]string) ([]ingest.Item, string, error) {
	posts, next, err := s.client.Fetch(ctx, reddit.FetchOpts{Subreddit: target}, cursor)
	if err != nil {
		return nil, "", err
	}

	items := make([]ingest.Item, 0, len(posts))
	for _, p := range posts {
		items = append(items, ingest.RedditItem(p, p.ImageURL))
	}

	return items, next, nil
}
