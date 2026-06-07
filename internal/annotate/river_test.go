package annotate

import (
	"encoding/json"
	"testing"
)

func TestWorkerResolvesFrozenSlugAndModality(t *testing.T) {
	registry := map[string]Annotator{
		"test-model": FakeAnnotator{
			Response: json.RawMessage(`{"patterns": []}`),
			ID:       RunIdentity{Provider: "test", Model: "test-model", ClientVersion: "0.0.1"},
		},
	}

	w := NewAnnotateWorker(nil, []Loader{TextLoader{}}, registry, nil)

	if _, ok := w.registry["test-model"]; !ok {
		t.Fatal("registry should resolve the frozen model slug")
	}

	if _, ok := w.loaders["text"]; !ok {
		t.Fatal("loaders should resolve the text modality")
	}

}
