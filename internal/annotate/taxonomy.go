package annotate

import (
	"context"
	"fmt"
	"strings"

	"github.com/assimoes/dsr/internal/db"
)

// Taxonomy is one pinned taxonomy version loaded for a run. Block is the flat text rendered into
// simple prompts; HighLevels and Patterns feed the richer ones; byCode maps a meso code to its row id.
type Taxonomy struct {
	Version    int32
	byCode     map[string]int32
	Block      string
	HighLevels []HighLevelView
	Patterns   []PatternView
}

// HighLevelView is one strategic-intent parent, rendered into the prompt.
type HighLevelView struct {
	Code       string
	Name       string
	Definition string
}

// PatternView is one meso pattern with its literature mappings and exemplars.
type PatternView struct {
	Code            string
	Name            string
	Parent          string
	Definition      string
	GrayMapping     []string
	SourceMapping   []string
	Examples        []string
	Counterexamples []string
}

// ID looks up the row id for a meso code. ok is false when the code isnt in this version.
func (t Taxonomy) ID(code string) (int32, bool) {
	id, ok := t.byCode[code]
	return id, ok
}

// LoadTaxonomy reads one taxonomy version from the db and builds the lookup map, prompt block,
// and the high-level/pattern views. errors if the version has no patterns.
func LoadTaxonomy(ctx context.Context, q *db.Queries, version int32) (Taxonomy, error) {
	rows, err := q.ListMesoPatternsByVersion(ctx, version)
	if err != nil {
		return Taxonomy{}, err
	}

	if len(rows) == 0 {
		return Taxonomy{}, fmt.Errorf("no taxonomy patterns at version %d", version)
	}

	highs, err := q.ListHighLevelsForMesoVersion(ctx, version)
	if err != nil {
		return Taxonomy{}, err
	}

	t := Taxonomy{
		Version: version,
		byCode:  make(map[string]int32, len(rows)),
	}

	for _, h := range highs {
		t.HighLevels = append(t.HighLevels, HighLevelView{
			Code:       h.Code,
			Name:       h.Name,
			Definition: h.Definition,
		})
	}

	var b strings.Builder

	for _, r := range rows {
		t.byCode[r.Code] = r.ID
		fmt.Fprintf(&b, "%s - %s (%s): %s\n", r.Code, r.Name, r.HighLevelPattern, r.Description)
		t.Patterns = append(t.Patterns, PatternView{
			Code:            r.Code,
			Name:            r.Name,
			Parent:          r.HighLevelCode,
			Definition:      r.Description,
			GrayMapping:     r.GrayMapping,
			SourceMapping:   r.SourceMapping,
			Examples:        r.Examples,
			Counterexamples: r.CounterExamples,
		})
	}

	t.Block = b.String()
	return t, nil
}
