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
		RawResponse:  toJSONB(raw),
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
			// closed-world: an entry with present=false is an explicit absence, store nothing.
			if d.Present != nil && !*d.Present {
				continue
			}

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
				Confidence:   numericOrNil(d.Confidence),
			}); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

// toJSONB guarantees a valid jsonb value: pass valid JSON through untouched; wrap
// anything else (a non-JSON or empty model reply, i.e. the parseError case) as a JSON string.
func toJSONB(raw json.RawMessage) json.RawMessage {
	if len(raw) > 0 && json.Valid(raw) {
		return raw
	}
	b, _ := json.Marshal(string(raw)) // always valid JSON (e.g. "" or "I can't help with that")
	return b
}
