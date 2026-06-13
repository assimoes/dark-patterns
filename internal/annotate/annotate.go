package annotate

import (
	"context"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Annotate is the unit of work the River worker runs.
func Annotate(ctx context.Context, pool *pgxpool.Pool, rc RenderCtx,
	tax Taxonomy, annotator Annotator, loader Loader, runID, annotatorID int32, individualID int64) error {

	q := db.New(pool)

	input, err := loader.Load(ctx, q, rc, individualID)
	if err != nil {
		return err
	}

	out, err := annotator.Annotate(ctx, input)
	if err != nil {
		return err
	}

	result, parseErr := Parse(out.Raw)
	status := "completed"

	if parseErr != nil {
		status = "parse_error"
	}

	return Persist(ctx, pool, runID, individualID, annotatorID, status, out.Raw, out.Meta, result, tax)
}
