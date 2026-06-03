package annotate

import (
	"context"
	"encoding/json"
)

// FakeAnnotator is here so that tests don't need a real model
type FakeAnnotator struct {
	Response json.RawMessage
	Err      error
	ID       RunIdentity
}

func (f FakeAnnotator) Indentity() RunIdentity {
	return f.ID
}

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
