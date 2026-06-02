package steam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func makeReviews(ids ...string) []Review {
	rs := make([]Review, len(ids))

	for i, id := range ids {
		rs[i] = Review{
			RecommendationID: id,
			Review:           "review-" + id,
		}
	}

	return rs
}

func newTestClient(srv *httptest.Server) *Client {
	c := New(WithHTTPClient(srv.Client()), WithPageDelay(0))
	c.baseURL = srv.URL
	c.backoffBase = time.Millisecond

	return c
}

func TestFetchPagePaginates(t *testing.T) {
	pages := map[string]ReviewResponse{
		"*":  {Success: 1, Cursor: "c1", Reviews: makeReviews("a", "b")},
		"c1": {Success: 1, Cursor: "c1", Reviews: makeReviews("c", "d")},
		"c2": {Success: 1, Cursor: "c2", Reviews: nil},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, ok := pages[r.URL.Query().Get("cursor")]
		if !ok {
			http.Error(w, "unknown cursor", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(resp)
	}))

	defer srv.Close()

	reviews, err := newTestClient(srv).FetchReviews(context.Background(), FetchOpts{AppID: "123"})
	if err != nil {
		t.Fatalf("error during fetch: %v", err)
	}

	if len(reviews) != 4 {
		t.Fatalf("want 4 reviews across two pages, got %d", len(reviews))
	}
}

func TestFetchPageStopsOnRepeatedCursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ReviewResponse{Success: 1, Cursor: "stuck", Reviews: makeReviews("a", "b")})
	}))

	defer srv.Close()

	reviews, err := newTestClient(srv).FetchReviews(context.Background(), FetchOpts{AppID: "123"})
	if err != nil {
		t.Fatalf("error during fetch: %v", err)
	}

	if len(reviews) != 4 {
		t.Fatalf("want 4 then stop, got %d", len(reviews))
	}
}

func TestFetchPageHonoursMaxReviews(t *testing.T) {
	pages := map[string]ReviewResponse{
		"*":  {Success: 1, Cursor: "c1", Reviews: makeReviews("a", "b")},
		"c1": {Success: 1, Cursor: "c1", Reviews: makeReviews("c", "d")},
		"c2": {Success: 1, Cursor: "c2", Reviews: nil},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, ok := pages[r.URL.Query().Get("cursor")]
		if !ok {
			http.Error(w, "unknown cursor", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(resp)
	}))

	reviews, err := newTestClient(srv).FetchReviews(context.Background(), FetchOpts{AppID: "123", MaxReviews: 3})
	if err != nil {
		t.Fatalf("error during fetch: %v", err)
	}

	if len(reviews) != 3 {
		t.Fatalf("want exactly 3 (trimmed), got %d", len(reviews))
	}
}

func TestFetchTerminalOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := newTestClient(srv).FetchReviews(context.Background(), FetchOpts{AppID: "123"}); err == nil {
		t.Fatal("HTTP 404 should be terminal, not retried into success")
	}
}
func TestFetchRetriesThenRecoversOn5xx(t *testing.T) {

	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(ReviewResponse{Success: 1, Reviews: makeReviews("a")})
	}))
	defer srv.Close()

	reviews, err := newTestClient(srv).FetchReviews(context.Background(), FetchOpts{AppID: "123"})
	if err != nil {
		t.Fatalf("should recover after a 500: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf("want 1 review, got %d", len(reviews))
	}
}

func TestFetchRetriesBadBody(t *testing.T) {

	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if calls.Add(1) == 1 {
			_, _ = w.Write([]byte("bad body"))
			return
		}

		_ = json.NewEncoder(w).Encode(ReviewResponse{Success: 1, Reviews: makeReviews("a")})
	}))
	defer srv.Close()

	reviews, err := newTestClient(srv).FetchReviews(context.Background(), FetchOpts{AppID: "123"})
	if err != nil {
		t.Fatalf("should retry on bad body response: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf("want 1 review, got %d", len(reviews))
	}
}

func TestFetchOnceReturnsOnePageAndCursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, _ = w.Write([]byte(`{"success": 1, "cursor" : "NEXT", "reviews": [
			{"recommendationid": "1", "review": "this is so pay to win", "language": "english"}
		]}`))
	}))

	defer srv.Close()

	c := newTestClient(srv)

	reviews, next, err := c.FetchOnce(context.Background(), FetchOpts{AppID: "730"}, "*")

	if err != nil {
		t.Fatalf("FetchOnce: %v", err)
	}

	if len(reviews) != 1 || next != "NEXT" {
		t.Fatalf("want 1 review and cursor NEXT, got %d reviews / %q", len(reviews), next)
	}
}
