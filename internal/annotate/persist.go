package annotate

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Persist writes one annotation (and response_meta) and it's pattern rows in a single transaction. Idempotent
func Persist(
	ctx context.Context,
	pool *pgxpool.Pool,
	runID int32,
	individualID int64,
	annotatorID int32,
	status string,
	raw json.RawMessage,
	meta ResponseMeta,
	result Result,
	tax Taxonomy,
) error {

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)

	id, err := q.UpsertAnnotation(ctx, db.UpsertAnnotationParams{
		RunID:        runID,
		IndividualID: individualID,
		AnnotatorID:  annotatorID,
		Status:       status,
		RawResponse:  raw,
		ResponseMeta: meta.JSON(),
	})

	// occurs when a previous attempt was already completed
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	if err := q.DeleteAnnotationPatterns(ctx, id); err != nil {
		return err
	}

	if status == "completed" {
		for _, d := range result.Patterns {
			patternID, ok := tax.ID(d.Code)
			if !ok {
				// the prompt forbids codes outside the taxonomy
				continue
			}

			if err := q.InsertAnnotationPattern(ctx, db.InsertAnnotationPatternParams{
				AnnotationID: id,
				PatternID:    patternID,
				Evidence:     ptrOrNil(d.Evidence),
				Explanation:  ptrOrNil(d.Explanation),
			}); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
