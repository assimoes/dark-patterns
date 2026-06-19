package annotate

type oraMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// oraReqMessage is the request side of a message. Content is `any` because a text turn sends a plain
// string while a multimodal turn sends an array of parts (text plus image_url).
type oraReqMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// oraContentPart is one piece of a multimodal user turn: either a text chunk or an image reference.
type oraContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *oraImageURL `json:"image_url,omitempty"`
}

type oraImageURL struct {
	URL string `json:"url"`
}

type oraRequest struct {
	Model       string          `json:"model"`
	Messages    []oraReqMessage `json:"messages"`
	Temperature float64         `json:"temperature"`

	// some models respect type: json_object and others dont.
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
