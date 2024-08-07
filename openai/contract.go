package openai

type Service interface {
	Completion(params *CompletionParams) (*CompletionResponse, error)
}

type CompletionParams struct {
	WaitRateLimit bool
	Model         string
	Messages      []Message
	Temperature   float64
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResponse struct {
	Choices []Choice
}

type Choice struct {
	Message Message `json:"message"`
}

type ErrRateLimit int

func (err ErrRateLimit) Error() string {
	return "rate limit"
}
