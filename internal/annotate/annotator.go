package annotate

import (
	"context"
	"encoding/json"
)

// RunIdentity is the provenance that isnt already a column on runs.
type RunIdentity struct {
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	ClientVersion string `json:"client_version"`
}

// Params is the identity as jsonb for the runs row.
func (id RunIdentity) Params() (json.RawMessage, error) {
	return json.Marshal(id)
}

// Image is one attachment for a multimodal annotator
type Image struct {
	URL      string
	MIMEType string
}

// Input is what a model needs for one item. Images is nil for text-only annotators.
type Input struct {
	System string
	User   string
	Images []Image
}

// ResponseMeta is what the provider actually did for one call.
type ResponseMeta struct {
	ServedModel       string          `json:"served_model"`
	FinishReason      string          `json:"finish_reason,omitempty"`
	PromptTokens      int             `json:"prompt_tokens,omitempty"`
	CompletionTokens  int             `json:"completion_tokens,omitempty"`
	LatencyMS         int64           `json:"latency_ms,omitempty"`
	RequestID         string          `json:"request_id,omitempty"`
	SystemFingerprint string          `json:"system_fingerprint,omitempty"`
	Extra             json.RawMessage `json:"extra,omitempty"`
}

// JSON marshals the meta for storage. no unmarshalable fields, err cant happen.
func (m ResponseMeta) JSON() json.RawMessage {
	b, _ := json.Marshal(m)
	return b
}

// Output is one models answer plus what the provider did. Raw is the unparsed json body.
type Output struct {
	Raw  json.RawMessage
	Meta ResponseMeta
}

// Annotator is anything that can label one item. real models and the fake both satisfy it.
type Annotator interface {
	Annotate(ctx context.Context, in Input) (Output, error)
	Identity() RunIdentity
}
