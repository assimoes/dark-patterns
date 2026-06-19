package annotate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestAnnotator(srv *httptest.Server, slug string) *OpenRouterAnnotator {
	a := NewOpenRouterAnnotator("test-key", slug, WithHTTPClient(srv.Client()), WithBaseURL(srv.URL+"/api/v1/chat/completions"))
	a.backoffBase = time.Millisecond

	return a
}

func TestOpenRouterAnnotateStripsFenceAndFillsMeta(t *testing.T) {
	const content = "```json\n{\"patterns\":[{\"code\":\"PM-1\",\"evidence\":\"pay to win\",\"explanation\":\"buys power\"}]}\n```"

	res := oraResponse{
		ID:    "gen-abc123",
		Model: "openai/gpt-4o-mini-2026-06-01",
		Choices: []oraChoice{{
			Message:      oraMessage{Role: "assistant", Content: content},
			FinishReason: "stop",
		}},
		Usage: oraUsage{PromptTokens: 412, CompletionTokens: 37, TotalTokens: 449},
	}

	var gotAuth, gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotURL = r.URL.Path

		var req oraRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Model != "openai/gpt-4o-mini" || req.Temperature != 0 || len(req.Messages) != 2 {
			t.Errorf("unexpected request: %+v", req)
		}

		_ = json.NewEncoder(w).Encode(res)
	}))
	defer srv.Close()

	a := newTestAnnotator(srv, "openai/gpt-4o-mini")
	a.referer = ""

	out, err := a.Annotate(context.Background(), Input{
		System: "you label reviews",
		User:   "this game is pay to win",
	})
	if err != nil {
		t.Fatalf("annotate: %v", err)
	}

	resp, perr := Parse(out.Raw)

	if perr != nil {
		t.Fatalf("parse stripped raw: %v (raw=%q)", perr, out.Raw)
	}

	if len(resp.Patterns) != 1 || resp.Patterns[0].Code != "PM-1" {
		t.Fatalf("patterns: %+v", resp.Patterns)
	}

	if out.Meta.ServedModel != "openai/gpt-4o-mini-2026-06-01" {
		t.Fatalf("served_model: got %q", out.Meta.ServedModel)
	}

	if out.Meta.FinishReason != "stop" {
		t.Fatalf("finish_reason: got %q", out.Meta.FinishReason)
	}

	if out.Meta.PromptTokens != 412 || out.Meta.CompletionTokens != 37 {
		t.Fatalf("usage: %+v", out.Meta)
	}

	if out.Meta.RequestID != "gen-abc123" {
		t.Fatalf("request_id: got %q", out.Meta.RequestID)
	}

	if out.Meta.LatencyMS < 0 {
		t.Fatalf("latency should be measured, got %d", out.Meta.LatencyMS)
	}

	if gotAuth != "Bearer test-key" {
		t.Fatalf("auth header: got %q", gotAuth)
	}

	if gotURL != "/api/v1/chat/completions" {
		t.Fatalf("path: got %q", gotURL)
	}
}

func TestOpenRouterRetriesThen200(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		if calls == 1 {
			http.Error(w, "upstream busy", http.StatusServiceUnavailable)
			return
		}

		_ = json.NewEncoder(w).Encode(oraResponse{
			Model:   "deepseek/deepseek-chat",
			Choices: []oraChoice{{Message: oraMessage{Content: `{"patterns":[]}`}, FinishReason: "stop"}},
		})
	}))
	defer srv.Close()

	out, err := newTestAnnotator(srv, "deepseek/deepseek-chat").Annotate(context.Background(), Input{
		System: "s",
		User:   "u",
	})
	if err != nil {
		t.Fatalf("should recover after a 503: %v", err)
	}

	if string(out.Raw) != `{"patterns":[]}` {
		t.Fatalf("raw: %q", out.Raw)
	}

	if calls != 2 {
		t.Fatalf("want 2 calls (one retry), got %d", calls)
	}
}

func TestOpenRouterTerminalOn4xx(t *testing.T) {
	var calls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "no key", http.StatusUnauthorized)
	}))
	defer srv.Close()

	if _, err := newTestAnnotator(srv, "openai/gpt-4o-mini").Annotate(context.Background(), Input{}); err == nil {
		t.Fatal("a 401 should be terminal")
	}

	if calls != 1 {
		t.Fatalf("a 4xx should not be retried, got %d calls", calls)
	}
}

func TestStripJSONFence(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"json fence", "```json\n{\"patterns\":[]}\n```", `{"patterns":[]}`},
		{"bare fence", "```\n{\"patterns\":[]}\n```", `{"patterns":[]}`},
		{"no fence", `{"patterns":[]}`, `{"patterns":[]}`},
		{"surrounding whitespace", "  \n```json\n{\"a\":1}\n```  \n", `{"a":1}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := stripJSONFence(c.in); got != c.want {
				t.Fatalf("in %q: want %q, got %q", c.in, c.want, got)
			}
		})
	}
}

var _ = io.Discard
