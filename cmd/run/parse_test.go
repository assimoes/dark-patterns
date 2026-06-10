package main

import "testing"

func TestParseAnnotatorIDs(t *testing.T) {
	cases := []struct {
		in   string
		want []int32
		ok   bool
	}{
		{"", nil, true},
		{"  ", nil, true},
		{"1,3,4", []int32{1, 3, 4}, true},
		{" 2 , 4 ", []int32{2, 4}, true},
		{"2,,4", []int32{2, 4}, true},
		{"2,abc", nil, false},
		{"x", nil, false},
	}

	for _, c := range cases {
		got, err := parseAnnotatorsIDs(c.in)
		if c.ok != (err == nil) {
			t.Fatalf("%q: want ok=%v, got err=%v", c.in, c.ok, err)
		}
		if !c.ok {
			continue
		}
		if len(got) != len(c.want) {
			t.Fatalf("%q: got %v, want %v", c.in, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("%q: got %v, want %v", c.in, got, c.want)
			}
		}
	}
}
