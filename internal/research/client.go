package research

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"

// Client makes the research call to OpenRouter with the web search server tool enabled. it is separate
// from the annotation client on purpose: web search lives only here, never in the annotation loop.
type Client struct {
	httpClient *http.Client
	apiKey     string
	model      string
	baseURL    string
	referer    string
	title      string
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the OpenRouter endpoint (tests).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = u } }

// WithHTTPClient overrides the http client (timeouts, tests).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithAttribution sets the OpenRouter ranking headers.
func WithAttribution(referer, title string) Option {
	return func(c *Client) { c.referer, c.title = referer, title }
}

// NewClient builds a research client for one model slug. web search can be slow, so the default timeout
// is generous.
func NewClient(apiKey, model string, opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 180 * time.Second},
		apiKey:     apiKey,
		model:      model,
		baseURL:    defaultBaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Model is the slug this client researches with, recorded as provenance.
func (c *Client) Model() string { return c.model }

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	Tools       []chatTool    `json:"tools,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatTool carries the OpenRouter web_search server tool. it runs server-side; the model gets the
// results and returns the final answer in one response.
type chatTool struct {
	Type       string         `json:"type"`
	Parameters chatToolParams `json:"parameters"`
}

type chatToolParams struct {
	Engine            string `json:"engine,omitempty"`
	MaxResults        int    `json:"max_results,omitempty"`
	SearchContextSize string `json:"search_context_size,omitempty"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Research runs one web-search-grounded call and returns the models raw text (expected to be the JSON
// object) plus the model id that served it. parsing and validation happen in the worker.
func (c *Client) Research(ctx context.Context, gameName, disambiguation string) (raw, modelUsed string, err error) {
	body, err := json.Marshal(chatRequest{
		Model:       c.model,
		Temperature: 0,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt(gameName, disambiguation)},
		},
		Tools: []chatTool{{
			Type:       "openrouter:web_search",
			Parameters: chatToolParams{Engine: "auto", MaxResults: 8, SearchContextSize: "high"},
		}},
	})
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.referer != "" {
		req.Header.Set("HTTP-Referer", c.referer)
	}
	if c.title != "" {
		req.Header.Set("X-Title", c.title)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("openrouter %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var out chatResponse
	if err := json.Unmarshal(payload, &out); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}
	if out.Error != nil {
		return "", "", fmt.Errorf("openrouter error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", "", fmt.Errorf("openrouter returned no choices")
	}

	return stripJSONFence(out.Choices[0].Message.Content), out.Model, nil
}

// stripJSONFence drops a leading/trailing markdown code fence so a fenced reply still parses, even though
// the prompt asks for none.
func stripJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimPrefix(s, "json")
	if i := strings.LastIndex(s, "```"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
