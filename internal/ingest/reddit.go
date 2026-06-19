package ingest

import "github.com/assimoes/dsr/internal/reddit"

// RedditItem maps a reddit post into a normalised Item. title and self-text are one text body; the
// reddit-only fields ride in source_meta. imageDataURI is set for an image post, empty otherwise, and
// decides whether the artifact is text or multimodal.
func RedditItem(p reddit.Post, imageDataURI, mime string) Item {
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
		ImageURI: imageDataURI,
		MimeType: mime,
	}
}
