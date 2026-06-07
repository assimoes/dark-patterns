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

type ORAOption func(*OpenRouterAnnotator)

func WithHTTPClient(h *http.Client) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.httpClient = h
	}
}

func WithBaseURL(url string) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.baseURL = url
	}
}

func WithClientVersion(version string) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.clientVersion = version
	}
}

func WithAttribution(referer, title string) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.referer = referer
		a.title = title
	}
}

func WithMaxRetries(maxRetries int) ORAOption {
	return func(a *OpenRouterAnnotator) {
		a.maxRetries = maxRetries
	}
}

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

func (a *OpenRouterAnnotator) Identity() RunIdentity {
	return RunIdentity{
		Provider:      "openrouter-annotator",
		Model:         a.model,
		ClientVersion: a.clientVersion,
	}
}

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
		// I can return the error safely as it is already retried inside "call"
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
		// potential transport error. retryable
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

	// jitter the backoff duration
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

	// strip opening fence
	s = strings.TrimPrefix(s, "```")

	// strips ```json, etc
	if nl := strings.IndexByte(s, '\n'); nl >= 0 {
		s = s[nl+1:]
	}

	// strip closing fence
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")

	return strings.TrimSpace(s)
}
