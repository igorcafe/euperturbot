package tg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"time"
)

type Bot struct {
	retry    *Retry
	token    string
	Username string
	baseURL  string
	client   http.Client
}

func NewBot(token string) *Bot {
	return &Bot{
		token:   token,
		baseURL: "https://api.telegram.org/bot",
		retry: &Retry{
			maxAttempts: 3,
			delay:       time.Second,
		},
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func apiJSONRequest[T any](bot *Bot, path string, data any) (*Result[T], error) {
	u := bot.baseURL + bot.token + "/" + path

	var reqBody []byte
	var reqReader io.Reader
	var err error

	if data != nil {
		reqBody, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
		reqReader = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequest("POST", u, reqReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	err = bot.retry.Do(func() error {
		resp, err = bot.client.Do(req)
		return err
	})

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, botRespError(resp, reqBody, respBody)
	}
	if err != nil {
		return nil, err
	}

	var v Result[T]
	err = json.Unmarshal(respBody, &v)
	if err != nil {
		return nil, err
	}
	if !v.Ok {
		panic("TODO")
	}
	return &v, nil
}

func (bot *Bot) GetMe() (*User, error) {
	res, err := apiJSONRequest[User](bot, "getMe", nil)
	if err != nil {
		return nil, err
	}
	bot.Username = res.Result.Username
	return &res.Result, err
}

func (bot *Bot) GetUpdates(params GetUpdatesParams) ([]Update, error) {
	res, err := apiJSONRequest[[]Update](bot, "getUpdates", params)
	if err != nil {
		return nil, err
	}
	return res.Result, nil
}

func (bot *Bot) GetUpdatesChannel() chan Update {
	ch := make(chan Update)
	go func() {
		updateID := 0
		for {
			params := GetUpdatesParams{
				Offset:         updateID,
				Timeout:        5,
				AllowedUpdates: []string{"message", "poll", "poll_answer"},
			}
			updates, err := bot.GetUpdates(params)
			if err != nil {
				log.Print(err)
			}
			for _, u := range updates {
				updateID = u.UpdateID + 1
				ch <- u
			}
			time.Sleep(time.Second)
		}
	}()
	return ch
}

func (bot *Bot) SendVoice(params SendVoiceParams) (*Message, error) {
	res, err := apiJSONRequest[Message](bot, "sendVoice", params)
	if err != nil {
		return nil, err
	}
	return &res.Result, nil
}

func (bot *Bot) SendPoll(params SendPollParams) (*Message, error) {
	res, err := apiJSONRequest[Message](bot, "sendPoll", params)
	if err != nil {
		return nil, err
	}
	return &res.Result, nil
}

func (bot *Bot) SendMessage(params SendMessageParams) (*Message, error) {
	res, err := apiJSONRequest[Message](bot, "sendMessage", params)
	if err != nil {
		return nil, err
	}
	return &res.Result, nil
}

func (bot *Bot) EditMessageText(params EditMessageTextParams) (*Message, error) {
	res, err := apiJSONRequest[Message](bot, "editMessageText", params)
	if err != nil {
		return nil, err
	}
	return &res.Result, nil
}

func botRespError(resp *http.Response, reqBody []byte, respBody []byte) error {
	u := path.Base(resp.Request.URL.String())
	return fmt.Errorf("call to %s (%s): status %s %s", u, string(reqBody), resp.Status, string(respBody))
}
