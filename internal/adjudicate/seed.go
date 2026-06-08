package adjudicate

import "encoding/json"

// PanelVerdict is one panel rater's verdict on one cell, frozen at decision time.
type PanelVerdict struct {
	AnnotatorID int32  `json:"annotator_id"`
	ModelSlug   string `json:"model_slug"`
	Present     bool   `json:"present"`
	Evidence    string `json:"evidence"`
	Explanation string `json:"explanation"`
}

// PanelSeed is the audit field that shows the exact panel state when the human decided
type PanelSeed struct {
	Vote     bool           `json:"vote"`
	NPresent int            `json:"n_present"`
	NTotal   int            `json:"n_total"`
	Votes    []PanelVerdict `json:"votes"`
}

// Majority present iif strictly more than half the completing raters detected it.
// Detection needs a majority, not just non-absence
func Majority(nPresent, nTotal int) bool {
	return nPresent*2 > nTotal
}

func (s PanelSeed) JSON() json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}

func ParseSeed(raw []byte) (PanelSeed, error) {
	var s PanelSeed
	if len(raw) == 0 {
		return s, nil
	}

	return s, json.Unmarshal(raw, &s)
}
