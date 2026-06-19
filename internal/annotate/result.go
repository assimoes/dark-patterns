package annotate

import "encoding/json"

// Detected is a pattern the model reported. Code is a meso code (e.g. PM-1).
// Present and Confidence are optional: simple prompts omit them; richer ones
// emit one entry per pattern, present=false for those ruled out.
type Detected struct {
	Code        string   `json:"code"`
	Evidence    string   `json:"evidence"`
	Explanation string   `json:"explanation"`
	Present     *bool    `json:"present"`
	Confidence  *float64 `json:"confidence"`
}

// Result is the whole parsed model answer, just the list of patterns it reported.
type Result struct {
	Patterns []Detected `json:"patterns"`
}

// Parse unmarshals a model reply into a Result. a bad body is the parse_error case.
func Parse(raw json.RawMessage) (Result, error) {
	var r Result

	if err := json.Unmarshal(raw, &r); err != nil {
		return Result{}, err
	}
	return r, nil
}
