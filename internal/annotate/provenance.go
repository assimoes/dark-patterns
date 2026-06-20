package annotate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// SnapshotPanel freezes who voted on a run into run_annotators: the pinned annotator_ids, or every
// active llm when none pinned. skips anything not in the registry. then stamps the config digest so
// the run records exactly what prompt, taxonomy, sampling, and models produced it.
func SnapshotPanel(ctx context.Context, q *db.Queries, run db.Run,
	prompt db.Prompt, tax Taxonomy, registry map[string]Annotator) error {

	annotators, err := q.ListAnnotators(ctx)
	if err != nil {
		return err
	}

	models, err := q.ListActiveModels(ctx)
	if err != nil {
		return err
	}

	slugByModelID := map[int32]string{}
	for _, m := range models {
		slugByModelID[m.ID] = m.Slug
	}

	sampling, _ := json.Marshal(map[string]any{"temperature": numericToFloat(run.Temperature)})

	var want map[int32]bool
	if len(run.AnnotatorIds) > 0 {
		want = make(map[int32]bool, len(run.AnnotatorIds))

		for _, id := range run.AnnotatorIds {
			want[id] = true
		}
	}

	var slugs []string

	for _, a := range annotators {
		if a.Kind != "llm" || a.ModelID == nil {
			continue
		}

		if want != nil && !want[a.ID] {
			continue
		}

		slug := slugByModelID[*a.ModelID]
		ann, ok := registry[slug]

		if !ok {
			continue
		}

		id := ann.Identity()

		if err := q.SnapshotRunAnnotator(ctx, db.SnapshotRunAnnotatorParams{
			RunID:         run.ID,
			AnnotatorID:   a.ID,
			Provider:      id.Provider,
			ModelSlug:     id.Model,
			ClientVersion: id.ClientVersion,
			Sampling:      sampling,
		}); err != nil {
			return err
		}

		slugs = append(slugs, id.Model)
	}

	// only a prompt that injects game context depends on the descriptions. for a no-context run we pin
	// nothing and the digest carries no descriptions, so it stays independent of them. that keeps an
	// A/B pair (same population, context vs no-context prompt) clean.
	var descLines []string
	if UsesGameContext(prompt.Template) {
		// freeze the approved description per game in the runs population, so the run records exactly which
		// description text each game contributed and the digest covers the whole set.
		if err := q.PinRunGameDescriptions(ctx, db.PinRunGameDescriptionsParams{
			RunID:        run.ID,
			PopulationID: run.PopulationID,
		}); err != nil {
			return err
		}

		descs, err := q.ListRunGameDescriptions(ctx, run.ID)
		if err != nil {
			return err
		}

		descLines = make([]string, 0, len(descs))
		for _, d := range descs {
			descLines = append(descLines, fmt.Sprintf("%d:%d:%s", d.ExternalGameID, d.Version, hex.EncodeToString(d.ContentHash)))
		}
	}

	return q.SetRunConfigDigest(ctx, db.SetRunConfigDigestParams{
		ID:           run.ID,
		ConfigDigest: ptrOrNil(configDigest(prompt, tax, run, slugs, descLines)),
	})
}

// UsesGameContext reports whether a prompt template injects the per-game description (the GameContext
// field). it gates the description requirement and the run's description pinning, so a plain prompt never
// pulls in descriptions.
func UsesGameContext(template string) bool {
	return strings.Contains(template, "GameContext")
}

func configDigest(prompt db.Prompt, tax Taxonomy, run db.Run, slugs, descriptions []string) string {
	sort.Strings(slugs)
	sort.Strings(descriptions)
	system := ""

	if prompt.SystemPrompt != nil {
		system = *prompt.SystemPrompt
	}

	h := sha256.New()
	fmt.Fprintf(
		h,
		"system=%s\ntemplate=%s\ntax_version=%d\ntemp=%v\nmodels=%v\ndescriptions=%v\n",
		system, prompt.Template, tax.Version, numericToFloat(run.Temperature), slugs, descriptions,
	)

	return hex.EncodeToString(h.Sum(nil))
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid || n.NaN {
		return 0
	}

	v, err := n.Value()
	if err != nil {
		return 0
	}

	s, ok := v.(string)
	if !ok {
		return 0
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	return f
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func numericOrNil(f *float64) pgtype.Numeric {
	var n pgtype.Numeric
	if f == nil {
		return n
	}

	_ = n.Scan(strconv.FormatFloat(*f, 'f', -1, 64))
	return n
}
