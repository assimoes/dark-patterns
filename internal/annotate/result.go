package annotate

import "encoding/json"

// Detected is a pattern the model says it found. Code is a taxonomy meso code (e.g. PM-1)
type Detected struct {
	Code        string `json:"code"`
	Evidence    string `json:"evidence"`
	Explanation string `json:"explanation"`
}

type Result struct {
	Patterns []Detected `json:"patterns"`
}

func Parse(raw json.RawMessage) (Result, error) {
	var r Result

	if err := json.Unmarshal(raw, &r); err != nil {
		return Result{}, err
	}
	return r, nil
}
