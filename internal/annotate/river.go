package annotate

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// AnnotateArgs is one durable job: annotate one individual with one panel member.
// ModelSlug and Modality are frozen at enqueue time.
type AnnotateArgs struct {
	RunID        int32  `json:"run_id"`
	IndividualID int64  `json:"individual_id"`
	AnnotatorID  int32  `json:"annotator_id"`
	ModelSlug    string `json:"model_slug"`
	Modality     string `json:"modality"`
}

// Kind is the River job kind.
func (AnnotateArgs) Kind() string {
	return "annotate"
}

// InsertOpts routes to the annotate queue and dedups by args so the same job cant land twice.
func (AnnotateArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "annotate",
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

// AnnotateWorker runs the panel.
type AnnotateWorker struct {
	river.WorkerDefaults[AnnotateArgs]
	pool     *pgxpool.Pool
	loaders  map[string]Loader
	registry map[string]Annotator
	logger   *slog.Logger
	runCtx   sync.Map
}

// NewAnnotateWorker wires up the worker, indexing the loaders by modality. nil logger falls back to default.
func NewAnnotateWorker(
	pool *pgxpool.Pool, loaders []Loader, registry map[string]Annotator, logger *slog.Logger) *AnnotateWorker {

	if logger == nil {
		logger = slog.Default()
	}

	byModality := make(map[string]Loader, len(loaders))
	for _, l := range loaders {
		byModality[l.Modality()] = l
	}

	return &AnnotateWorker{
		pool:     pool,
		loaders:  byModality,
		registry: registry,
		logger:   logger,
	}
}

// runContext is built once per run and cached in the workers sync map.
type runContext struct {
	rc  RenderCtx
	tax Taxonomy
}

func (w *AnnotateWorker) loadRunContext(ctx context.Context, runID int32) (*runContext, error) {
	if v, ok := w.runCtx.Load(runID); ok {
		return v.(*runContext), nil
	}

	q := db.New(w.pool)

	run, err := q.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	prompt, err := q.GetPrompt(ctx, run.PromptID)
	if err != nil {
		return nil, err
	}

	if run.TaxonomyVersion == nil {
		return nil, fmt.Errorf("run %d has no taxonomy_version pinned", runID)
	}

	tax, err := LoadTaxonomy(ctx, q, *run.TaxonomyVersion)
	if err != nil {
		return nil, err
	}

	userTmpl, err := template.New("user").Parse(prompt.Template)
	if err != nil {
		return nil, err
	}

	system := ""
	if prompt.SystemPrompt != nil {
		sysTmpl, err := template.New("system").Parse(*prompt.SystemPrompt)
		if err != nil {
			return nil, err
		}

		var sb strings.Builder
		if err := sysTmpl.Execute(&sb, promptData{
			HighLevels:      tax.HighLevels,
			Patterns:        tax.Patterns,
			TaxonomyVersion: strconv.Itoa(int(tax.Version)),
			PromptVersion:   strconv.Itoa(int(prompt.Version)),
			Taxonomy:        tax.Block,
		}); err != nil {
			return nil, err
		}
		system = sb.String()
	}

	rcx := &runContext{
		rc: RenderCtx{
			System:   system,
			Template: userTmpl,
			Taxonomy: tax.Block,
		},
		tax: tax,
	}

	actual, _ := w.runCtx.LoadOrStore(runID, rcx)
	return actual.(*runContext), nil
}

// Work picks the annotator and loader for the job, loads the cached run context, then annotates.
func (w *AnnotateWorker) Work(ctx context.Context, job *river.Job[AnnotateArgs]) error {
	a := job.Args

	annotator, ok := w.registry[a.ModelSlug]
	if !ok {
		return fmt.Errorf("no annotator registred for model slug %q", a.ModelSlug)
	}

	loader, ok := w.loaders[a.Modality]
	if !ok {
		return fmt.Errorf("no loader registred for modality %q", a.Modality)
	}

	rcx, err := w.loadRunContext(ctx, a.RunID)
	if err != nil {
		return err
	}

	return Annotate(ctx, w.pool, rcx.rc, rcx.tax, annotator, loader, a.RunID, a.AnnotatorID, a.IndividualID)
}

// Enqueue freezes the panel for a run, then inserts one job per annotator per still-unannotated
// individual. returns how many jobs went in. safe to re-run, the unique opts skip dupes.
func Enqueue(ctx context.Context, client *river.Client[pgx.Tx],
	pool *pgxpool.Pool, runID int32, registry map[string]Annotator) (int, error) {

	q := db.New(pool)

	run, err := q.GetRun(ctx, runID)
	if err != nil {
		return 0, err
	}

	prompt, err := q.GetPrompt(ctx, run.PromptID)
	if err != nil {
		return 0, err
	}

	if run.TaxonomyVersion == nil {
		return 0, fmt.Errorf("run %d has no taxonomy_version pinned", runID)
	}

	tax, err := LoadTaxonomy(ctx, q, *run.TaxonomyVersion)
	if err != nil {
		return 0, err
	}

	if err := SnapshotPanel(ctx, q, run, prompt, tax, registry); err != nil {
		return 0, err
	}

	panel, err := q.ListRunAnnotators(ctx, runID)
	if err != nil {
		return 0, err
	}

	inserted := 0

	for _, pa := range panel {
		individuals, err := q.ListUnnanotatedIndividuals(ctx, db.ListUnnanotatedIndividualsParams{
			RunID:       runID,
			AnnotatorID: pa.AnnotatorID,
		})
		if err != nil {
			return inserted, err
		}

		for _, indID := range individuals {
			if _, err := client.Insert(ctx, AnnotateArgs{
				RunID:        runID,
				IndividualID: indID,
				AnnotatorID:  pa.AnnotatorID,
				ModelSlug:    pa.ModelSlug,
				Modality:     prompt.Modality,
			}, nil); err != nil {
				return inserted, err
			}
			inserted++
		}
	}

	return inserted, nil
}
