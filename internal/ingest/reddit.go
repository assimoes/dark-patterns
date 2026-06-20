package ingest

import "github.com/assimoes/dsr/internal/reddit"

// RedditItem maps a reddit post into a normalised Item. title and self-text are one text body; the
// reddit-only fields ride in source_meta. imageURL is the post image link (or empty), stored as-is and
// downloaded at annotation time; a non-empty link makes the artifact multimodal.
func RedditItem(p reddit.Post, imageURL string) Item {
	body := p.Title
	if p.Body != "" {
		body = p.Title + "\n\n" + p.Body
	}

	return Item{
		SourceID: p.ID,
		Body:     body,
		Lang:     "en",
		Meta: map[string]any{
			"subreddit": p.Subreddit,
			"author":    p.Author,
			"permalink": p.Permalink,
		},
		ImageURI: imageURL,
	}
}
