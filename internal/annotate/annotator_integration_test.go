//go:build integration

package annotate

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"text/template"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testGameID = int32(999000008)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run db integration tests")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

func seedRun(t *testing.T, pool *pgxpool.Pool) (db.Run, db.Prompt, []int64) {
	t.Helper()
	ctx := context.Background()

	// the integration db is shared, so this test owns a fixed footprint (the sentinel game id and its
	// named rows). clear it before seeding so a leftover from a crashed run does not collide, and clear
	// it again after via t.Cleanup, which runs even when the test fails.
	cleanSnapshotTestData(ctx, pool)
	t.Cleanup(func() { cleanSnapshotTestData(context.Background(), pool) })

	q := db.New(pool)

	var popID int32

	if err := pool.QueryRow(ctx,
		`INSERT INTO populations (modality, description, artifacts_cutoff)
		 VALUES ('text', 'annotate-it-test', now()) RETURNING id;
		`).Scan(&popID); err != nil {
		t.Fatalf("population: %v", err)
	}

	var indIDs []int64

	for i, body := range []string{"this game is pay to win, you must buy to compete", "grindy but fair"} {
		srcID := "it-rec-" + string(rune('a'+i))

		var artID int64

		if err := pool.QueryRow(ctx,
			`INSERT INTO artifacts (modality, source, source_id, content_hash, scraped_at, external_game_id)
       VALUES ('text', 'steam', $1, $2, now(), $3) RETURNING id;`,
			srcID, []byte(srcID), testGameID).Scan(&artID); err != nil {
			t.Fatalf("artifact: %v", err)
		}

		if _, err := pool.Exec(ctx,
			`INSERT INTO text_review_details(artifact_id, body, voted_up, hours_played, lang)
			VALUES ($1, $2, true, 600, 'english')`, artID, body); err != nil {
			t.Fatalf("text detail: %v", err)
		}

		var indID int64
		if err := pool.QueryRow(ctx,
			`INSERT INTO individuals (population_id, artifact_id) VALUES ($1, $2) RETURNING id;`,
			popID, artID).Scan(&indID); err != nil {
			t.Fatalf("text detail: %v", err)
		}

		indIDs = append(indIDs, indID)
	}

	modelID, err := q.UpsertModel(ctx, db.UpsertModelParams{
		Family:     "test",
		Slug:       "test-model",
		Name:       "test-model",
		Modalities: []string{"text"},
	})

	if err != nil {
		t.Fatalf("model: %v", err)
	}

	annotatorID, err := q.CreateAnnotator(ctx, db.CreateAnnotatorParams{
		Kind:    "llm",
		ModelID: &modelID,
		Label:   "test-model rater",
	})
	if err != nil {
		t.Fatalf("annotator: %v", err)
	}

	_ = annotatorID

	sys := "you label steam reviews for dark patterns. only use codes you are given."
	promptID, err := q.CreatePrompt(ctx, db.CreatePromptParams{
		Name:         "p2w-it",
		Version:      1,
		Modality:     "text",
		SystemPrompt: &sys,
		Template:     "patterns:\n{{.Taxonomy}}\n\nreview:\n{{.Content}}",
	})
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}

	var runID int32

	if err := pool.QueryRow(ctx,
		`INSERT INTO runs (run_type, population_id, prompt_id, temperature, taxonomy_version)
		VALUES ('llm_panel', $1, $2, 0.0, 1) RETURNING id;
		`, popID, promptID).Scan(&runID); err != nil {
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

	return run, prompt, indIDs
}

// cleanSnapshotTestData removes everything seedRun writes, in foreign-key-safe order: children before
// parents, keyed on the sentinel game id and the test's named rows. errors are ignored so a clean run
// against an already-empty db is a no-op. the model row is left alone, UpsertModel is idempotent.
func cleanSnapshotTestData(ctx context.Context, pool *pgxpool.Pool) {
	exec := func(sql string, args ...any) { _, _ = pool.Exec(ctx, sql, args...) }

	exec(`DELETE FROM annotation_patterns ap USING annotations an, individuals i, artifacts a
	      WHERE ap.annotation_id = an.id AND an.individual_id = i.id AND i.artifact_id = a.id
	        AND a.external_game_id = $1`, testGameID)
	exec(`DELETE FROM annotations an USING individuals i, artifacts a
	      WHERE an.individual_id = i.id AND i.artifact_id = a.id AND a.external_game_id = $1`, testGameID)
	exec(`DELETE FROM run_annotators ra USING runs r, populations p
	      WHERE ra.run_id = r.id AND r.population_id = p.id AND p.description = 'annotate-it-test'`)
	exec(`DELETE FROM runs r USING populations p
	      WHERE r.population_id = p.id AND p.description = 'annotate-it-test'`)
	exec(`DELETE FROM individuals i USING artifacts a
	      WHERE i.artifact_id = a.id AND a.external_game_id = $1`, testGameID)
	exec(`DELETE FROM text_review_details td USING artifacts a
	      WHERE td.artifact_id = a.id AND a.external_game_id = $1`, testGameID)
	exec(`DELETE FROM artifacts WHERE external_game_id = $1`, testGameID)
	exec(`DELETE FROM populations WHERE description = 'annotate-it-test'`)
	exec(`DELETE FROM prompts WHERE name = 'p2w-it' AND version = 1`)
	exec(`DELETE FROM annotators WHERE label = 'test-model rater'`)
}

func TestSnapshotAndAnnotate(t *testing.T) {
	pool := testPool(t)

	ctx := context.Background()

	q := db.New(pool)

	run, prompt, indIDs := seedRun(t, pool)

	fake := FakeAnnotator{
		Response: json.RawMessage(`
		{"patterns": [{"code": "PM-1", "evidence": "pay to win", "explanation": "buys power"}]}
		`),
		ID: RunIdentity{Provider: "test", Model: "test-model", ClientVersion: "0.0.1"},
	}

	registry := map[string]Annotator{"test-model": fake}

	// load taxonomy
	tax, err := LoadTaxonomy(ctx, q, *run.TaxonomyVersion)
	if err != nil {
		t.Fatalf("load taxonomy: %v", err)
	}

	// freeze the panel
	if err := SnapshotPanel(ctx, q, run, prompt, tax, registry); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	panel, err := q.ListRunAnnotators(ctx, run.ID)
	if err != nil {
		t.Fatalf("list run annotators: %v", err)
	}

	if len(panel) != 1 || panel[0].ModelSlug != "test-model" {
		t.Fatalf("frozen panel: %+v", err)
	}

	// run config_digest should be stamped
	got, _ := q.GetRun(ctx, run.ID)
	if got.ConfigDigest == nil || *got.ConfigDigest == "" {
		t.Fatal("config_digest not stamped")
	}

	// annotate each individual with the frozen panel member

	tmpl, _ := template.New("prompt").Parse(prompt.Template)

	sys := ""
	if prompt.SystemPrompt != nil {
		sys = *prompt.SystemPrompt
	}

	rc := RenderCtx{
		System:   sys,
		Template: tmpl,
		Taxonomy: tax.Block,
	}

	for _, indID := range indIDs {
		if err := Annotate(ctx, pool, rc, tax, fake, TextLoader{}, run.ID, panel[0].AnnotatorID, indID); err != nil {
			t.Fatalf("annotate %d: %v", indID, err)
		}
	}

	// assert

	var completed, withMeta int

	_ = pool.QueryRow(ctx,
		`SELECT count(*) FILTER (WHERE status = 'completed'),
			count(*) FILTER (WHERE response_meta IS NOT NULL)
	 FROM annotations WHERE run_id = $1`, run.ID).Scan(&completed, &withMeta)

	if completed != 2 || withMeta != 2 {
		t.Fatalf("annotations: want 2 completed / 2 with meta, got %d / %d", completed, withMeta)
	}

	var patterns int

	_ = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM annotation_patterns ap
	JOIN annotations an ON an.id = ap.annotation_id
	JOIN taxonomy_meso_levels t ON t.id = ap.pattern_id
	WHERE an.run_id = $1 AND t.code = 'PM-1'`, run.ID).Scan(&patterns)

	if patterns != 2 {
		t.Fatalf("PM-1 patterns: want 2, got %d", patterns)
	}

	// ensure idempotency

	for _, indID := range indIDs {
		if err := Annotate(ctx, pool, rc, tax, fake, TextLoader{}, run.ID, panel[0].AnnotatorID, indID); err != nil {
			t.Fatalf("re-annotate %d: %v", indID, err)
		}
	}

	_ = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM annotation_patterns ap
	JOIN annotations an ON an.id = ap.annotation_id
	JOIN taxonomy_meso_levels t ON t.id = ap.pattern_id
	WHERE an.run_id = $1 AND t.code = 'PM-1'`, run.ID).Scan(&patterns)

	if patterns != 2 {
		t.Fatalf("PM-1 patterns: want 2, got %d", patterns)
	}
}
