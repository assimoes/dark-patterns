package scrape

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

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

// Fetch grabs one page of posts for a subreddit and maps them to items. an image post has its picture
// downloaded and inlined as a data uri, the same shape uploads use. a failed image download falls back
// to a text-only item.
func (s redditSource) Fetch(ctx context.Context, target, cursor string, params map[string]string) ([]ingest.Item, string, error) {
	posts, next, err := s.client.Fetch(ctx, reddit.FetchOpts{Subreddit: target}, cursor)
	if err != nil {
		return nil, "", err
	}

	items := make([]ingest.Item, 0, len(posts))
	for _, p := range posts {
		dataURI, mime := s.inlineImage(ctx, p.ImageURL)
		items = append(items, ingest.RedditItem(p, dataURI, mime))
	}

	return items, next, nil
}

// inlineImage downloads an image and returns it as a data uri plus its mime type. an empty url or a
// failed download returns empty, leaving the post text-only.
func (s redditSource) inlineImage(ctx context.Context, imageURL string) (dataURI, mime string) {
	if imageURL == "" {
		return "", ""
	}

	data, mime, err := s.client.FetchImage(ctx, imageURL)
	if err != nil || len(data) == 0 {
		return "", ""
	}
	if mime == "" {
		mime = http.DetectContentType(data)
	}

	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data)), mime
}
