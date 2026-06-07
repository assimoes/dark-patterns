package annotate

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestFakeAnnotator(t *testing.T) {
	var a Annotator = FakeAnnotator{
		Response: json.RawMessage(`{"patterns": []}`),
		ID:       RunIdentity{Provider: "test", Model: "test-model", ClientVersion: "0.0.1"},
	}

	out, err := a.Annotate(context.Background(), Input{System: "sys", User: "user"})
	if err != nil {
		t.Fatalf("annotate: %v", err)
	}

	if string(out.Raw) != `{"patterns": []}` {
		t.Fatalf("raw: got %s", out.Raw)
	}

	if out.Meta.ServedModel != "test-model" {
		t.Fatalf("meta: want served_model test_model, got %q", out.Meta.ServedModel)
	}

	if a.Identity().Model != "test-model" {
		t.Fatalf("identity: %+v", a.Identity())
	}

}

func TestFakeAnnotatorReturnsError(t *testing.T) {
	f := FakeAnnotator{Err: errors.New("model unavailable")}
	if _, err := f.Annotate(context.Background(), Input{}); err == nil {
		t.Fatal("wanted an error, got nil")
	}
}

func TestParseReadsPatterns(t *testing.T) {

	r, err := Parse(json.RawMessage(`{"patterns": [{"code": "PM-1", "evidence":"pay to win", "explanation": "buys power"}]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(r.Patterns) != 1 || r.Patterns[0].Code != "PM-1" {
		t.Fatalf("patterns: %+v", r.Patterns)
	}
}

func TestRunIdentityParams(t *testing.T) {
	id := RunIdentity{Provider: "test", Model: "test-model", ClientVersion: "1.2.3"}
	raw, err := id.Params()
	if err != nil {
		t.Fatalf("params: %v", err)
	}

	var back RunIdentity
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("round-trip: %v", err)
	}

	if back != id {
		t.Fatalf("round-trip mismatch: %+v, got %+v", back, id)
	}
}
