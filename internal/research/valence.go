package research

import (
	"regexp"
	"sort"
	"strings"
)

// Flag is one valenced term found in a description, with where it sits, so the reviewer can clear it.
type Flag struct {
	Term  string `json:"term"`
	Index int    `json:"index"`
}

// valenceTerms are evaluative/reputational words a neutral description must not contain. matched
// whole-word, case-insensitive. the list is deliberately broad: a flag is a prompt for the reviewer to
// look, not a verdict, and approval is blocked while any remain.
var valenceTerms = []string{
	"predatory", "exploitative", "exploitorary", "abusive", "egregious", "greedy", "scummy", "shady",
	"manipulative", "deceptive", "dark pattern", "anti-consumer", "anti consumer", "consumer-friendly",
	"consumer friendly", "fair", "unfair", "generous", "stingy", "ethical", "unethical", "predator",
	"problematic", "aggressive", "cash grab", "cash-grab", "pay-to-win", "pay to win", "p2w", "ripoff",
	"rip-off", "rip off", "fleece", "fleece", "gouging", "notorious", "infamous", "beloved", "praised",
	"criticized", "criticised", "controversial", "backlash", "outrage", "toxic", "evil", "rewarding",
	"punishing", "grindy", "addictive", "predation",
}

var valencePattern = buildValencePattern()

func buildValencePattern() *regexp.Regexp {
	quoted := make([]string, len(valenceTerms))
	for i, t := range valenceTerms {
		quoted[i] = regexp.QuoteMeta(t)
	}
	// longest first so multi-word terms match before their parts.
	sort.Slice(quoted, func(i, j int) bool { return len(quoted[i]) > len(quoted[j]) })
	return regexp.MustCompile(`(?i)\b(` + strings.Join(quoted, "|") + `)\b`)
}

// ScanValence finds evaluative/reputational terms in a description. an empty result means the text reads
// as neutral by the lexicon; any flags block approval until the reviewer rewords them.
func ScanValence(text string) []Flag {
	matches := valencePattern.FindAllStringIndex(text, -1)
	flags := make([]Flag, 0, len(matches))
	for _, m := range matches {
		flags = append(flags, Flag{Term: strings.ToLower(text[m[0]:m[1]]), Index: m[0]})
	}
	return flags
}
