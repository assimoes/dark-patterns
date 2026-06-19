package annotate

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Persist writes one annotation and its pattern rows in a single transaction. idempotent.
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
		RawResponse:  toJSONB(raw),
		ResponseMeta: meta.JSON(),
	})

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
			if d.Present != nil && !*d.Present {
				continue
			}

			patternID, ok := tax.ID(d.Code)
			if !ok {
				continue
			}

			if err := q.InsertAnnotationPattern(ctx, db.InsertAnnotationPatternParams{
				AnnotationID: id,
				PatternID:    patternID,
				Evidence:     ptrOrNil(d.Evidence),
				Explanation:  ptrOrNil(d.Explanation),
				Confidence:   numericOrNil(d.Confidence),
			}); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

// toJSONB returns valid jsonb: pass JSON through, wrap anything else (the parse_error case) as a string.
func toJSONB(raw json.RawMessage) json.RawMessage {
	if len(raw) > 0 && json.Valid(raw) {
		return raw
	}
	b, _ := json.Marshal(string(raw))
	return b
}
