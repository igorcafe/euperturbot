package openai

var _ Service = ServiceDouble{}

type ServiceDouble struct {
}

func (s ServiceDouble) Completion(params *CompletionParams) (*CompletionResponse, error) {
	return &CompletionResponse{
		Choices: []Choice{
			{
				Message: Message{
					Role:    "assistant",
					Content: "response from assistant",
				},
			},
		},
	}, nil
}
