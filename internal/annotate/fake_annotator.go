package annotate

import (
	"context"
	"encoding/json"
)

// FakeAnnotator lets tests run without a real model.
type FakeAnnotator struct {
	Response json.RawMessage
	Err      error
	ID       RunIdentity
}

// Identity returns whatever the test set.
func (f FakeAnnotator) Identity() RunIdentity {
	return f.ID
}

// Annotate hands back the canned Response, or Err if one was set.
func (f FakeAnnotator) Annotate(_ context.Context, _ Input) (Output, error) {
	if f.Err != nil {
		return Output{}, f.Err
	}

	return Output{
		Raw: f.Response,
		Meta: ResponseMeta{
			ServedModel:  f.ID.Model,
			FinishReason: "stop",
		},
	}, nil
}
