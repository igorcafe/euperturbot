package openai

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (fn RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func Test(t *testing.T) {
	wantRole := "assistant"
	wantContent := "hello there!"

	payload := `{
		"choices": [
			{ "message": {"role": "` + wantRole + `", "content": "` + wantContent + `"}}
		]
	}`

	http := http.Client{
		Transport: RoundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(payload)),
			}, nil
		}),
	}

	s := NewService("", &http)
	cmp, err := s.Completion(&CompletionParams{
		Messages: []Message{
			{
				Role:    "user",
				Content: "hello",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	gotRole := cmp.Choices[0].Message.Role
	if wantRole != gotRole {
		t.Fatalf("role - want: %s, got: %s", wantRole, gotRole)
	}

	gotContent := cmp.Choices[0].Message.Content
	if wantContent != gotContent {
		t.Fatalf("content - want: '%s', got: '%s'", wantContent, gotContent)
	}
}
