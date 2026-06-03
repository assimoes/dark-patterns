package annotate

import (
	"context"
	"encoding/json"
)

// RunIdentity is the provenance that isn't already a column on runs.
type RunIdentity struct {
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	ClientVersion string `json:"client_version"`
}

func (id RunIdentity) Params() (json.RawMessage, error) {
	return json.Marshal(id)
}

// Image is one attachment for a multimodal annotator
type Image struct {
	URL      string
	MIMEType string
}

// Input is all a model needs for one item. If text-only annotator, Images is nil
type Input struct {
	System string
	User   string
	Images []Image
}

// ResponseMeta is the evidence half of provenance: what the provider actually did for one call
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

func (m ResponseMeta) JSON() json.RawMessage {
	// we can safely ignore the err. ResponseMeta has no un-marshalables fields
	b, _ := json.Marshal(m)
	return b
}

type Output struct {
	Raw  json.RawMessage
	Meta ResponseMeta
}

type Annotator interface {
	Annotate(ctx context.Context, in Input) (Output, error)
	Indentity() RunIdentity
}
