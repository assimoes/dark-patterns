package annotate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

const (
	openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	maxBackoff    = 30 * time.Second
)

// OpenRouterAnnotator talks to the OpenRouter chat API. One per model slug. Retries on 429/5xx.
type OpenRouterAnnotator struct {
	httpClient    *http.Client
	apiKey        string
	model         string
	baseURL       string
	clientVersion string
	referer       string
	title         string
	maxRetries    int
	backoffBase   time.Duration
}

// ORAOption tweaks an OpenRouterAnnotator at construction.
type ORAOption func(*OpenRouterAnnotator)

// WithHTTPClient swaps the http client, handy for tests.
func WithHTTPClient(h *http.Client) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.httpClient = h
	}
}

// WithBaseURL points at a different endpoint, e.g. a stub server in tests.
func WithBaseURL(url string) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.baseURL = url
	}
}

// WithClientVersion sets what gets recorded as the client version in provenance.
func WithClientVersion(version string) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.clientVersion = version
	}
}

// WithAttribution sets the HTTP-Referer and X-Title headers OpenRouter uses for attribution.
func WithAttribution(referer, title string) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.referer = referer
		a.title = title
	}
}

// WithMaxRetries caps how many times a retryable call gets re-tried.
func WithMaxRetries(maxRetries int) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.maxRetries = maxRetries
	}
}

// NewOpenRouterAnnotator builds one with sane defaults, then applies opts.
func NewOpenRouterAnnotator(apiKey, slug string, opts ...ORAOption) *OpenRouterAnnotator {
	a := &OpenRouterAnnotator{
		httpClient:    &http.Client{Timeout: 90 * time.Second},
		apiKey:        apiKey,
		model:         slug,
		baseURL:       openRouterURL,
		clientVersion: "openrouter/v1",
		maxRetries:    3,
		backoffBase:   time.Second,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Identity is the provenance for this annotator: provider, model slug, client version.
func (a *OpenRouterAnnotator) Identity() RunIdentity {
	return RunIdentity{
		Provider:      "openrouter-annotator",
		Model:         a.model,
		ClientVersion: a.clientVersion,
	}
}

// Annotate sends the system and user messages at temperature 0, strips any json fence off the
// reply, and returns the raw body. Retries live in call.
func (a *OpenRouterAnnotator) Annotate(ctx context.Context, in Input) (Output, error) {

	body := oraRequest{
		Model:       a.model,
		Temperature: 0,
		Messages: []oraMessage{
			{Role: "system", Content: in.System},
			{Role: "user", Content: in.User},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return Output{}, err
	}

	res, meta, err := a.call(ctx, payload)
	if err != nil {
		// already retried inside call
		return Output{}, err
	}

	if res.Error != nil {
		return Output{}, fmt.Errorf("openrouter error for %s: %s", a.model, res.Error.Message)
	}

	if len(res.Choices) == 0 {
		return Output{}, fmt.Errorf("openrouter returned no choices for %s", a.model)
	}

	raw := stripJSONFence(res.Choices[0].Message.Content)

	meta.FinishReason = res.Choices[0].FinishReason

	return Output{
		Raw:  json.RawMessage(raw),
		Meta: meta,
	}, nil
}

func (a *OpenRouterAnnotator) call(ctx context.Context, payload []byte) (oraResponse, ResponseMeta, error) {
	var lastErr error
	for attempt := 0; ; attempt++ {
		res, meta, retry, err := a.try(ctx, payload)
		if err == nil {
			return res, meta, nil
		}

		if ctx.Err() != nil {
			return oraResponse{}, ResponseMeta{}, ctx.Err()
		}
		if !retry {
			return oraResponse{}, ResponseMeta{}, err
		}

		lastErr = err

		if attempt >= a.maxRetries {
			return oraResponse{}, ResponseMeta{}, fmt.Errorf("exhausted %d retries: %w", a.maxRetries, lastErr)
		}

		select {
		case <-ctx.Done():
			return oraResponse{}, ResponseMeta{}, ctx.Err()
		case <-time.After(a.backoff(attempt)):
		}
	}
}

func (a *OpenRouterAnnotator) try(ctx context.Context, payload []byte) (oraResponse, ResponseMeta, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL, bytes.NewReader(payload))
	if err != nil {
		return oraResponse{}, ResponseMeta{}, false, err
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	if a.referer != "" {
		req.Header.Set("HTTP-Referer", a.referer)
	}
	if a.title != "" {
		req.Header.Set("X-Title", a.title)
	}

	start := time.Now()
	res, err := a.httpClient.Do(req)
	if err != nil {
		// transport error, retryable
		return oraResponse{}, ResponseMeta{}, true, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	latency := time.Since(start).Milliseconds()

	switch {
	case res.StatusCode == http.StatusTooManyRequests, res.StatusCode >= 500:
		return oraResponse{}, ResponseMeta{}, true, fmt.Errorf("openrouter status %d: %s", res.StatusCode, snippet(body))
	case res.StatusCode != http.StatusOK:
		return oraResponse{}, ResponseMeta{}, false, fmt.Errorf("openrouter status %d: %s", res.StatusCode, snippet(body))
	}

	var response oraResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return oraResponse{}, ResponseMeta{}, true, fmt.Errorf("decode openrouter response: %w", err)
	}

	meta := ResponseMeta{
		ServedModel:      response.Model,
		PromptTokens:     response.Usage.PromptTokens,
		CompletionTokens: response.Usage.CompletionTokens,
		RequestID:        response.ID,
		LatencyMS:        latency,
	}

	return response, meta, false, nil
}

func (a *OpenRouterAnnotator) backoff(attempt int) time.Duration {
	// 1s, 2s, 4s...30s
	base := min(a.backoffBase<<attempt, maxBackoff)

	half := base / 2

	if half <= 0 {
		return base
	}

	// jitter
	return half + time.Duration(rand.Int64N(int64(half)))
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))

	if len(s) > 200 {
		return s[:200] + "..."
	}

	return s
}

func stripJSONFence(s string) string {
	s = strings.TrimSpace(s)

	// no fence, return as-is
	if !strings.HasPrefix(s, "```") {
		return s
	}

	s = strings.TrimPrefix(s, "```")

	// drop the ```json (or bare ```) info line
	if nl := strings.IndexByte(s, '\n'); nl >= 0 {
		s = s[nl+1:]
	}

	s = strings.TrimSuffix(strings.TrimSpace(s), "```")

	return strings.TrimSpace(s)
}
