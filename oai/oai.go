package oai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/igoracmelo/euperturbot/util"
)

type Client struct {
	key               string
	http              http.Client
	mut               *sync.Mutex
	rateLimitDeadline *atomic.Value
}

func NewClient(key string) *Client {
	deadline := &atomic.Value{}
	deadline.Store(time.Time{})
	return &Client{
		key:               key,
		mut:               new(sync.Mutex),
		rateLimitDeadline: deadline,
	}
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
	Choices []struct {
		Message Message `json:"message"`
	}
}

type ErrRateLimit int

func (err ErrRateLimit) Error() string {
	return "rate limit"
}

func (c *Client) Completion(params *CompletionParams) (*CompletionResponse, error) {
	// if params.WaitRateLimit {
	// 	c.mut.Lock()
	// } else if !c.mut.TryLock() {
	// 	secs := int(time.Until(c.rateLimitDeadline.Load().(time.Time)).Seconds())
	// 	return nil, ErrRateLimit(secs)
	// }

	deadline := time.Now().Add(20 * time.Second)
	c.rateLimitDeadline.Store(deadline)
	go func() {
		time.Sleep(time.Until(deadline))
		c.mut.Unlock()
	}()

	if params.Model == "" {
		params.Model = "gpt-3.5-turbo"
	}
	if params.Temperature == 0 {
		params.Temperature = 0.7
	}

	for i, m := range params.Messages {
		if m.Role == "" {
			m.Role = "user"
		}
		params.Messages[i] = m
	}

	payload := map[string]any{
		"model":       params.Model,
		"messages":    params.Messages,
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

	if resp.StatusCode == 429 {
		return nil, ErrRateLimit(30)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, util.HTTPResponseError(resp)
	}

	var completion CompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&completion)
	return &completion, err
}
