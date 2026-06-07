package annotate

type oraMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type oraRequest struct {
	Model       string       `json:"model"`
	Messages    []oraMessage `json:"messages"`
	Temperature float64      `json:"temperature"`

	// some models respect type: json_object and others don't.
	ResponseFormat *oraResponseFormat `json:"response_format,omitempty"`
}

type oraResponseFormat struct {
	Type string `json:"type"`
}

type oraResponse struct {
	ID      string      `json:"id"`
	Model   string      `json:"model"`
	Choices []oraChoice `json:"choices"`
	Usage   oraUsage    `json:"usage"`
	Error   *oraError   `json:"error,omitempty"`
}

type oraChoice struct {
	Message      oraMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

type oraUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type oraError struct {
	Message string `json:"message"`
	Code    any    `json:"code"`
}
