package adjudicate

import "encoding/json"

// PanelVerdict is one panel raters verdict on one cell, frozen at decision time.
type PanelVerdict struct {
	AnnotatorID int32  `json:"annotator_id"`
	ModelSlug   string `json:"model_slug"`
	Present     bool   `json:"present"`
	Evidence    string `json:"evidence"`
	Explanation string `json:"explanation"`
}

// PanelSeed is the panel state captured when the human decided.
type PanelSeed struct {
	Vote     bool           `json:"vote"`
	NPresent int            `json:"n_present"`
	NTotal   int            `json:"n_total"`
	Votes    []PanelVerdict `json:"votes"`
}

// Majority is true when more than half the raters detected it (not just non-absence).
func Majority(nPresent, nTotal int) bool {
	return nPresent*2 > nTotal
}

// JSON marshals the seed for storage. No unmarshalable fields so the error cant happen.
func (s PanelSeed) JSON() json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}

// ParseSeed reads a stored seed back. Empty bytes give a zero seed, not an error.
func ParseSeed(raw []byte) (PanelSeed, error) {
	var s PanelSeed
	if len(raw) == 0 {
		return s, nil
	}

	return s, json.Unmarshal(raw, &s)
}
