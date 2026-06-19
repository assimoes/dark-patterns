package run

import (
	"context"
	"fmt"

	"github.com/assimoes/dsr/internal/db"
)

// Validate checks a run can be opened: annotators exist and are llm, the population has individuals,
// the prompt modality is supported, and the taxonomy version has codes. takes db.Querier so the api
// (pool or tx bound) and the cli (*db.Queries) both call it. the errors here are bad user input, the
// caller maps them to a 400.
func Validate(ctx context.Context, q db.Querier, annotatorIDs []int32, populationID, promptID, taxVersion int32) error {
	if len(annotatorIDs) > 0 {
		rows, err := q.ListAnnotatorsByIDs(ctx, annotatorIDs)
		if err != nil {
			return fmt.Errorf("loading annotators %v: %w", annotatorIDs, err)
		}

		found := make(map[int32]db.Annotator, len(rows))
		for _, a := range rows {
			found[a.ID] = a
		}

		for _, id := range annotatorIDs {
			a, ok := found[id]
			if !ok {
				return fmt.Errorf("annotator %d does not exist", id)
			}
			if a.Kind != "llm" || a.ModelID == nil {
				return fmt.Errorf("annotator %d is not an llm annotator", id)
			}
		}
	}

	n, err := q.CountIndividuals(ctx, populationID)
	if err != nil {
		return fmt.Errorf("counting individuals for population %d: %w", populationID, err)
	}
	if n == 0 {
		return fmt.Errorf("population %d has no individuals (curate it first)", populationID)
	}

	prompt, err := q.GetPrompt(ctx, promptID)
	if err != nil {
		return fmt.Errorf("loading prompt %d: %w", promptID, err)
	}
	if prompt.Modality != "text" && prompt.Modality != "image" && prompt.Modality != "multimodal" {
		return fmt.Errorf("prompt %d has unsupported modality %q", promptID, prompt.Modality)
	}

	codes, err := q.ListMesoPatternsByVersion(ctx, taxVersion)
	if err != nil {
		return fmt.Errorf("loading taxonomy version %d: %w", taxVersion, err)
	}
	if len(codes) == 0 {
		return fmt.Errorf("taxonomy version %d has no codes", taxVersion)
	}

	return nil
}
