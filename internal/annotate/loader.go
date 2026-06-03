package annotate

import (
	"context"
	"html/template"
	"strings"

	"github.com/assimoes/dsr/internal/db"
)

// RenderCtx is the per-run prompt material, shared across modalities
// handed on every run
type RenderCtx struct {
	System   string
	Template *template.Template
	Taxonomy string
}

// Loader fetches one item's content and renders it into the model Input
// One implementation per modality.
type Loader interface {
	Modality() string
	Load(ctx context.Context, q *db.Queries, rc RenderCtx, individualID int64) (Input, error)
}

type TextLoader struct{}

func (TextLoader) Modality() string {
	return "text"
}

func (TextLoader) Load(ctx context.Context, q *db.Queries,
	rc RenderCtx, individualID int64) (Input, error) {

	row, err := q.GetTextReviewForIndividual(ctx, individualID)
	if err != nil {
		return Input{}, err
	}

	var b strings.Builder
	if err := rc.Template.Execute(&b, map[string]string{"Taxonomy": rc.Taxonomy, "Content": row.Body}); err != nil {
		return Input{}, err
	}

	return Input{
		System: rc.System,
		User:   b.String(),
	}, nil
}
