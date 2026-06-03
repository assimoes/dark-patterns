package steam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://store.steampowered.com"
	maxPerPage     = 100
	maxBackoff     = 30 * time.Second
)

type Client struct {
	httpClient  *http.Client
	baseURL     string
	userAgent   string
	pageDelay   time.Duration
	maxRetries  int
	backoffBase time.Duration
	logger      *slog.Logger
}

type Options func(*Client)

func WithHTTPClient(h *http.Client) Options {
	return func(c *Client) {
		c.httpClient = h
	}
}

func WithUserAgent(ua string) Options {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithPageDelay(delay time.Duration) Options {
	return func(c *Client) {
		c.pageDelay = delay
	}
}

func WithMaxRetries(retries int) Options {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

func WithLogger(l *slog.Logger) Options {
	return func(c *Client) {
		c.logger = l
	}
}

func New(opts ...Options) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     defaultBaseURL,
		userAgent:   "dsr-scraper/1.0",
		pageDelay:   500 * time.Millisecond,
		maxRetries:  3,
		backoffBase: time.Second,
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) FetchOnce(ctx context.Context, opts FetchOpts, cursor string) (reviews []Review, next string, err error) {
	opts.normalize()
	if cursor == "" {
		cursor = "*"
	}

	page, err := c.fetchPage(ctx, opts, cursor)
	if err != nil {
		return nil, "", err
	}

	return page.Reviews, page.Cursor, nil
}

func (c *Client) FetchReviews(ctx context.Context, opts FetchOpts) ([]Review, error) {
	var all []Review

	err := c.FetchPage(ctx, opts, "*", func(page []Review, _ string) error {
		all = append(all, page...)
		return nil
	})

	return all, err
}

func (c *Client) FetchPage(ctx context.Context,
	opts FetchOpts, startCursor string, fn func(reviews []Review, next string) error) error {

	opts.normalize()

	if startCursor == "" {
		startCursor = "*"
	}

	cursor := startCursor
	seen := map[string]bool{}

	fetched := 0

	for {
		perPage := opts.MaxPerPage

		if opts.MaxReviews > 0 {
			remaining := opts.MaxReviews - fetched
			if remaining <= 0 {
				return nil
			}

			if remaining < perPage {
				perPage = remaining
			}
		}

		pageOpts := opts
		pageOpts.MaxPerPage = perPage

		page, err := c.fetchPage(ctx, pageOpts, cursor)
		if err != nil {
			return err
		}

		if len(page.Reviews) == 0 {
			return nil
		}

		reviews := page.Reviews

		if opts.MaxReviews > 0 && fetched+len(reviews) > opts.MaxReviews {
			reviews = reviews[:opts.MaxReviews-fetched]
		}

		fetched += len(reviews)
		next := page.Cursor

		if err := fn(reviews, next); err != nil {
			return err
		}

		if opts.MaxReviews > 0 && fetched >= opts.MaxReviews {
			return nil
		}

		// sometimes steam returns the same cursor instead of a new one.
		// this prevents an infinite loop fetching reviews
		if next == "" || next == cursor || seen[next] {
			return nil
		}

		seen[next] = true
		cursor = next

		if c.pageDelay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.pageDelay):
			}
			c.logger.Info("fetched reviews", "count", fetched)
		}
	}
}

func (c *Client) fetchPage(ctx context.Context, opts FetchOpts, cursor string) (*ReviewResponse, error) {
	q := url.Values{}
	q.Set("json", "1")
	q.Set("filter", opts.Filter)
	q.Set("language", opts.Language)
	q.Set("num_per_page", strconv.Itoa(opts.MaxPerPage))
	q.Set("cursor", cursor)

	endpoint := fmt.Sprintf("%s/appreviews/%s?%s", c.baseURL, url.PathEscape(opts.AppID), q.Encode())

	var lastErr error
	for attempt := 0; ; attempt++ {
		out, retryAfter, err := c.try(ctx, endpoint)
		if err == nil {
			return out, nil
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		var re retryable
		if !errors.As(err, &re) {
			return nil, err
		}

		lastErr = err

		if attempt >= c.maxRetries {
			return nil, fmt.Errorf("exhausted %d retries: %w", c.maxRetries, lastErr)
		}

		wait := c.backoff(attempt, retryAfter)

		c.logger.Debug("retrying steam request", "attempt", attempt+1, "wait", wait, "err", err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}

func (c *Client) try(ctx context.Context, endpoint string) (*ReviewResponse, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, retryable{err}
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		return nil, parseRetryAfter(
				res.Header.Get("Retry-After")),
			retryable{fmt.Errorf("steam api status %d", res.StatusCode)}
	case res.StatusCode >= 500:
		return nil, 0, retryable{fmt.Errorf("steam api status %d", res.StatusCode)}
	case res.StatusCode != http.StatusOK:
		return nil, 0, fmt.Errorf("steam api status %d", res.StatusCode)
	}

	var out ReviewResponse

	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, 0, retryable{fmt.Errorf("decode error: %w", err)}
	}

	if out.Success != 1 {
		return nil, 0, retryable{fmt.Errorf("steam api success response=%d", out.Success)}
	}

	return &out, 0, nil
}

func (c *Client) backoff(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}

	base := min(c.backoffBase<<attempt, maxBackoff)
	half := base / 2

	if half <= 0 {
		return base
	}

	return half + time.Duration(rand.Int64N(int64(half)))
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}

	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}

	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}

	return 0
}
