package annotate

import (
	"context"
	"fmt"
	"strings"

	"github.com/assimoes/dsr/internal/db"
)

type Taxonomy struct {
	Version int32
	byCode  map[string]int32
	Block   string
}

func (t Taxonomy) ID(code string) (int32, bool) {
	id, ok := t.byCode[code]
	return id, ok
}

func LoadTaxonomy(ctx context.Context, q *db.Queries, version int32) (Taxonomy, error) {
	rows, err := q.ListMesoPatternsByVersion(ctx, version)
	if err != nil {
		return Taxonomy{}, err
	}

	if len(rows) == 0 {
		return Taxonomy{}, fmt.Errorf("no taxonomy patterns at version %d", version)
	}

	t := Taxonomy{
		Version: version,
		byCode:  make(map[string]int32, len(rows)),
	}

	var b strings.Builder

	for _, r := range rows {
		t.byCode[r.Code] = r.ID
		fmt.Fprintf(&b, "%s - %s (%s): %s\n", r.Code, r.Name, r.HighLevelPattern, r.Description)
	}

	t.Block = b.String()
	return t, nil
}
