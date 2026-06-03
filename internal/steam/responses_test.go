package steam

import (
	"encoding/json"
	"testing"
)

func TestFlexStringUnmarshal(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"json string", `"0.5"`, "0.5"},
		{"base number", `0.5`, "0.5"},
		{"integer string", `"0"`, "0"},
		{"null", "null", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s flexString

			if err := json.Unmarshal([]byte(c.in), &s); err != nil {
				t.Fatalf("in %q: want %q, got %q", c.in, c.want, string(s))
			}

			if string(s) != c.want {
				t.Fatalf("in %q: want %q, got %q", c.in, c.want, string(s))
			}
		})
	}
}

func TestReviewDecodesBareScore(t *testing.T) {
	body := `{"recommendationid": "42", "review": "good", "weighted_vote_score": 0.83}`

	var r Review

	if err := json.Unmarshal([]byte(body), &r); err != nil {
		t.Fatalf("unmarchar review: %v", err)
	}

	if r.WeightedVotedScore != "0.83" {
		t.Fatalf("score: want 0.83, got %q", r.WeightedVotedScore)
	}
}
