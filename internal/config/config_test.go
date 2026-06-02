package config

import "testing"

func env(m map[string]string) func(string) string {
	return func(k string) string {
		return m[k]
	}
}

func TestLoadValidates(t *testing.T) {
	good := env(map[string]string{"DATABASE_URL": "postgres://x"})

	cases := []struct {
		name    string
		args    []string
		getenv  func(string) string
		wantErr bool
	}{
		{"ok", []string{"-app", "730"}, good, false},
		{"no dsn", nil, good, true},
		{"non-numeric app", []string{"-app", "notanid"}, good, true},
		{"negative max", []string{"-app", "730", "-max", "-1"}, good, true},
		{"bad filter", []string{"-app", "730", "-filter", "badfilter"}, good, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(tc.args, tc.getenv)
			if (err != nil) != tc.wantErr {
				t.Fatalf("args %v: wantErr=%v, got %v", tc.args, tc.wantErr, err)
			}
		})
	}
}

func TestLoadParseGameID(t *testing.T) {
	c, err := Load([]string{"-app", "730", "-max", "50"}, env(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if c.GameID != 730 {
		t.Fatalf("game id: want 730, got %d", c.GameID)
	}

	if c.MaxReviews != 50 {
		t.Fatalf("max: want 50, got %d", c.MaxReviews)
	}
}
