package scrape

import (
	"context"

	"github.com/assimoes/dsr/internal/ingest"
	"github.com/assimoes/dsr/internal/steam"
)

type steamSource struct {
	client *steam.Client
}

// NewSteamSource wraps a steam client as a Source.
func NewSteamSource(client *steam.Client) Source {
	return steamSource{client: client}
}

// Name is "steam".
func (steamSource) Name() string {
	return "steam"
}

// Fetch grabs one page of reviews for an app id and maps them to items. params carry the steam filter
// and language; an empty cursor starts from the top.
func (s steamSource) Fetch(ctx context.Context, target, cursor string, params map[string]string) ([]ingest.Item, string, error) {
	if cursor == "" {
		cursor = "*"
	}

	filter := params["filter"]
	if filter == "" {
		filter = "recent"
	}
	language := params["language"]
	if language == "" {
		language = "english"
	}

	reviews, next, err := s.client.FetchOnce(ctx, steam.FetchOpts{
		AppID:    target,
		Filter:   filter,
		Language: language,
	}, cursor)
	if err != nil {
		return nil, "", err
	}

	items := make([]ingest.Item, 0, len(reviews))
	for _, r := range reviews {
		items = append(items, ingest.SteamItem(r))
	}

	return items, next, nil
}
