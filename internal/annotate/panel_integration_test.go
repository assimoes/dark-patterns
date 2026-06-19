//go:build integration

package annotate

import (
	"context"
	"sort"
	"testing"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// seedPanelRun makes a population, prompt, and run pinning annotatorIDs (nil => NULL), and returns the run and prompt.
func seedPanelRun(t *testing.T, pool *pgxpool.Pool, annotatorIDs []int32) (db.Run, db.Prompt) {
	t.Helper()
	ctx := context.Background()
	q := db.New(pool)

	cleanPanelTestData(ctx, pool, t.Name())
	t.Cleanup(func() { cleanPanelTestData(context.Background(), pool, t.Name()) })

	var popID int32
	if err := pool.QueryRow(ctx,
		`INSERT INTO populations (modality, description, artifacts_cutoff)
		 VALUES ('text','panel-it',now()) RETURNING id`).Scan(&popID); err != nil {
		t.Fatalf("population: %v", err)
	}

	sys := "label reviews"
	promptID, err := q.CreatePrompt(ctx, db.CreatePromptParams{
		Name: "panel-it-" + t.Name(), Version: 1, Modality: "text",
		SystemPrompt: &sys, Template: "patterns:\n{{.Taxonomy}}\n\n{{.Content}}",
	})
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}

	var runID int32
	if err := pool.QueryRow(ctx,
		`INSERT INTO runs (run_type, population_id, prompt_id, temperature, taxonomy_version, annotator_ids)
		 VALUES ('llm_panel',$1,$2,0.0,1,$3) RETURNING id`, popID, promptID, annotatorIDs).Scan(&runID); err != nil {
		t.Fatalf("run: %v", err)
	}

	run, err := q.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	prompt, err := q.GetPrompt(ctx, promptID)
	if err != nil {
		t.Fatalf("get prompt: %v", err)
	}
	return run, prompt
}

// cleanPanelTestData removes the run, population, and prompt seedPanelRun writes for one test, keyed on
// the fixed population description and the per-test prompt name. fk-safe order, errors ignored so a
// clean db is a no-op.
func cleanPanelTestData(ctx context.Context, pool *pgxpool.Pool, testName string) {
	exec := func(sql string, args ...any) { _, _ = pool.Exec(ctx, sql, args...) }

	exec(`DELETE FROM run_annotators ra USING runs r, populations p
	      WHERE ra.run_id = r.id AND r.population_id = p.id AND p.description = 'panel-it'`)
	exec(`DELETE FROM runs r USING populations p
	      WHERE r.population_id = p.id AND p.description = 'panel-it'`)
	exec(`DELETE FROM populations WHERE description = 'panel-it'`)
	exec(`DELETE FROM prompts WHERE name = $1 AND version = 1`, "panel-it-"+testName)
}

// fakeLLMRegistry maps every seeded llm slug to a fake so SnapshotPanel can resolve each.
func fakeLLMRegistry(t *testing.T, pool *pgxpool.Pool) map[string]Annotator {
	t.Helper()
	anns, err := db.New(pool).ListLLMAnnotators(context.Background())
	if err != nil {
		t.Fatalf("list llm annotators: %v", err)
	}

	reg := make(map[string]Annotator, len(anns))
	for _, a := range anns {
		reg[a.Slug] = FakeAnnotator{ID: RunIdentity{Provider: "test", Model: a.Slug, ClientVersion: "0.0.1"}}
	}
	return reg
}

func frozenAnnotatorIDs(t *testing.T, pool *pgxpool.Pool, runID int32) []int32 {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT annotator_id FROM run_annotators WHERE run_id=$1 ORDER BY annotator_id`, runID)
	if err != nil {
		t.Fatalf("query run_annotators: %v", err)
	}
	defer rows.Close()

	var out []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, id)
	}
	return out
}

func TestPanelNullFreezesAll(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	q := db.New(pool)

	allLLM, err := q.ListLLMAnnotators(ctx)
	if err != nil {
		t.Fatalf("list llm annotators: %v", err)
	}
	if len(allLLM) == 0 {
		t.Skip("no llm annotators seeded")
	}

	run, prompt := seedPanelRun(t, pool, nil)
	tax, err := LoadTaxonomy(ctx, q, 1)
	if err != nil {
		t.Fatalf("load taxonomy: %v", err)
	}

	if err := SnapshotPanel(ctx, q, run, prompt, tax, fakeLLMRegistry(t, pool)); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	got := frozenAnnotatorIDs(t, pool, run.ID)
	if len(got) != len(allLLM) {
		t.Fatalf("NULL panel: want all %d llm annotators frozen, got %d (%v)", len(allLLM), len(got), got)
	}
}

func TestPanelSubsetFreezesChosen(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	q := db.New(pool)

	allLLM, err := q.ListLLMAnnotators(ctx)
	if err != nil {
		t.Fatalf("list llm annotators: %v", err)
	}
	if len(allLLM) < 2 {
		t.Skip("need >= 2 llm annotators")
	}

	want := []int32{allLLM[0].ID, allLLM[1].ID}
	sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })

	run, prompt := seedPanelRun(t, pool, want)
	tax, err := LoadTaxonomy(ctx, q, 1)
	if err != nil {
		t.Fatalf("load taxonomy: %v", err)
	}

	if err := SnapshotPanel(ctx, q, run, prompt, tax, fakeLLMRegistry(t, pool)); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	got := frozenAnnotatorIDs(t, pool, run.ID)
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("subset panel: want exactly %v frozen, got %v", want, got)
	}

	g, _ := q.GetRun(ctx, run.ID)
	if g.ConfigDigest == nil || *g.ConfigDigest == "" {
		t.Fatal("config_digest not stamped for subset run")
	}
}
