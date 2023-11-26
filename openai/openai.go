package openai

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/igoracmelo/euperturbot/util"
)

type Client struct {
	key  string
	http http.Client
}

func NewClient(key string) *Client {
	return &Client{
		key: key,
	}
}

type CompletionParams struct {
	Model       string
	Messages    []string
	Temperature float64
}

type CompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string
		}
	}
}

func (c *Client) Completion(params *CompletionParams) (*CompletionResponse, error) {
	if params.Model == "" {
		params.Model = "gpt-3.5-turbo"
	}
	if params.Temperature == 0 {
		params.Temperature = 0.7
	}

	messages := make([]map[string]string, len(params.Messages))
	for i, m := range params.Messages {
		messages[i] = map[string]string{
			"role":    "user",
			"content": m,
		}
	}

	payload := map[string]any{
		"model":       params.Model,
		"messages":    messages,
		"temperature": params.Temperature,
	}

	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.key)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, util.HTTPResponseError(resp)
	}

	var completion CompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&completion)
	return &completion, err
}
