// Package reddit is a small client for reddit's public Atom feeds, /r/<sub>/new/.rss, with cursor
// paging and 429 backoff. it needs no API key: the JSON endpoints are blocked but the feeds are served.
package reddit

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const maxBackoff = 60 * time.Second

// Client reads subreddit listings from reddit's Atom feeds.
type Client struct {
	httpClient *http.Client
	userAgent  string
	maxRetries int
	logger     *slog.Logger
}

// Options tweaks a Client at construction, pass them to New.
type Options func(*Client)

// WithHTTPClient swaps the http client, handy for tests with a stub transport.
func WithHTTPClient(h *http.Client) Options {
	return func(c *Client) {
		c.httpClient = h
	}
}

// WithUserAgent sets the User-Agent reddit sees. reddit wants a real one like "go:dsr-ingest:0.1 (by /u/name)".
func WithUserAgent(ua string) Options {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithMaxRetries sets how many times a rate-limited request is retried before giving up.
func WithMaxRetries(retries int) Options {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithLogger sets the logger. default discards everything.
func WithLogger(l *slog.Logger) Options {
	return func(c *Client) {
		c.logger = l
	}
}

// New builds a Client with sane defaults then applies opts.
func New(opts ...Options) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "go:dsr-ingest:0.1",
		maxRetries: 4,
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// FetchOpts is the target of one fetch: which subreddit to read.
type FetchOpts struct {
	Subreddit string
}

// Post is one submission flattened to the fields the feed carries. score and comment count are not in
// the feed. ImageURL is empty for a text-only post.
type Post struct {
	ID        string
	Subreddit string
	Author    string
	Title     string
	Body      string
	Permalink string
	Published time.Time
	ImageURL  string
}

type atomFeed struct {
	Entries []struct {
		ID      string `xml:"id"`
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		Author  struct {
			Name string `xml:"name"`
		} `xml:"author"`
		Category struct {
			Term string `xml:"term,attr"`
		} `xml:"category"`
		Link struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Content string `xml:"content"`
	} `xml:"entry"`
}

var (
	bodyRe = regexp.MustCompile(`(?s)<!-- SC_OFF -->(.*?)<!-- SC_ON -->`)
	imgRe  = regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	hrefRe = regexp.MustCompile(`href="([^"]+\.(?:jpg|jpeg|png|gif))"`)
	tagRe  = regexp.MustCompile(`<[^>]+>`)
	wsRe   = regexp.MustCompile(`[ \t]*\n[ \t]*`)
)

// Fetch grabs one page of /new for a subreddit. cursor is the fullname to read after (empty for the
// first page). it hands back the posts and the next cursor; the scrape worker pages itself, one page
// per job. an empty next cursor means the feed ended.
func (c *Client) Fetch(ctx context.Context, opts FetchOpts, cursor string) (posts []Post, next string, err error) {
	sub := strings.TrimPrefix(strings.TrimPrefix(opts.Subreddit, "r/"), "/r/")
	q := url.Values{}
	q.Set("limit", "100")
	if cursor != "" {
		q.Set("after", cursor)
	}
	endpoint := fmt.Sprintf("https://www.reddit.com/r/%s/new/.rss?%s", sub, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("reddit r/%s rss: status %d", sub, resp.StatusCode)
	}

	var feed atomFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, "", err
	}

	for _, e := range feed.Entries {
		pub, _ := time.Parse(time.RFC3339, e.Updated)
		posts = append(posts, Post{
			ID:        e.ID,
			Subreddit: e.Category.Term,
			Author:    strings.TrimPrefix(e.Author.Name, "/u/"),
			Title:     html.UnescapeString(e.Title),
			Body:      cleanBody(e.Content),
			Permalink: e.Link.Href,
			Published: pub,
			ImageURL:  imageURL(e.Content),
		})
	}

	if len(posts) > 0 {
		next = posts[len(posts)-1].ID
	}
	return posts, next, nil
}

// FetchImage downloads an image and returns its bytes and mime type, so an image post can be inlined
// the same way uploads are stored.
func (c *Client) FetchImage(ctx context.Context, imageURL string) (data []byte, mime string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch image: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return body, resp.Header.Get("Content-Type"), nil
}

// do runs the request and retries on 429, honouring Retry-After. reddit rate-limits unauthenticated
// feeds, so a background pull backs off rather than failing the page.
func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	for try := 0; ; try++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests || try == c.maxRetries {
			return resp, nil
		}

		wait := time.Duration(try+1) * 5 * time.Second
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, e := time.ParseDuration(ra + "s"); e == nil {
				wait = secs
			}
		}
		if wait > maxBackoff {
			wait = maxBackoff
		}
		resp.Body.Close()
		c.logger.Info("reddit rate limited, backing off", "wait", wait)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}

// cleanBody pulls the self-text out of the entry content, between the SC_OFF and SC_ON markers reddit
// wraps the body in, strips the HTML and trims. a link or image post with no self-text returns empty.
func cleanBody(content string) string {
	m := bodyRe.FindStringSubmatch(content)
	if m == nil {
		return ""
	}
	t := html.UnescapeString(m[1])
	t = strings.ReplaceAll(t, "</p>", "\n\n")
	t = strings.ReplaceAll(t, "<br/>", "\n")
	t = tagRe.ReplaceAllString(t, "")
	t = html.UnescapeString(t)
	t = wsRe.ReplaceAllString(t, "\n")
	return strings.TrimSpace(t)
}

// imageURL returns the post image when the entry embeds one, else empty.
func imageURL(content string) string {
	c := html.UnescapeString(content)
	if m := hrefRe.FindStringSubmatch(c); m != nil {
		return m[1]
	}
	for _, m := range imgRe.FindAllStringSubmatch(c, -1) {
		src := m[1]
		if strings.Contains(src, "i.redd.it") || strings.Contains(src, "preview.redd.it") ||
			strings.Contains(src, "external-preview") {
			return src
		}
	}
	return ""
}
