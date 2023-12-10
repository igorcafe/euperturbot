package tg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/igoracmelo/euperturbot/util"
)

type Bot struct {
	retry    *util.Retry
	token    string
	Username string
	baseURL  string
	client   http.Client
}

func NewBot(token string) *Bot {
	return &Bot{
		token:   token,
		baseURL: "https://api.telegram.org/bot",
		retry: &util.Retry{
			MaxAttempts: 3,
			Delay:       time.Second,
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
		return nil, errors.New(bot.hideToken(err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	err = bot.retry.Do(func() error {
		resp, err = bot.client.Do(req)
		return err
	})

	if err != nil {
		return nil, errors.New(bot.hideToken(err.Error()))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, bot.respError(resp, reqBody, respBody)
	}
	if err != nil {
		return nil, err
	}

	var v Result[T]
	err = json.Unmarshal(respBody, &v)
	if err != nil {
		return nil, err
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

func (bot *Bot) GetChatMember(params GetChatMemberParams) (*ChatMember, error) {
	res, err := apiJSONRequest[ChatMember](bot, "getMe", params)
	if err != nil {
		return nil, err
	}
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
				AllowedUpdates: []string{"message", "poll", "poll_answer", "callback_query", "inline_query"},
			}
			updates, err := bot.GetUpdates(params)
			if err != nil {
				log.Print(bot.hideToken(err.Error()))
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

func (bot *Bot) AnswerInlineQuery(params AnswerInlineQueryParams) error {
	_, err := apiJSONRequest[bool](bot, "answerInlineQuery", params)
	return err
}

func (bot *Bot) SendDocument(params SendDocumentParams) error {
	f, err := os.Open(params.FileName)
	if err != nil {
		return err
	}
	defer f.Close()

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	part, err := mw.CreateFormFile("document", params.FileName)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, f)
	if err != nil {
		return err
	}

	err = mw.WriteField("chat_id", fmt.Sprint(params.ChatID))
	if err != nil {
		return err
	}

	err = mw.Close()
	if err != nil {
		return err
	}

	u := bot.baseURL + bot.token + "/sendDocument"
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := bot.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res Result[any]
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return err
	}

	return nil
}

func (bot *Bot) hideToken(s string) string {
	return strings.ReplaceAll(s, bot.token, "<token>")
}

func (bot *Bot) respError(resp *http.Response, reqBody []byte, respBody []byte) error {
	err := BotError{}
	err.Path = bot.hideToken(resp.Request.URL.String())
	err.Status = resp.StatusCode
	err.RequestBody = reqBody
	err.ResponseBody = respBody
	return err
}

type BotError struct {
	Path         string
	Status       int
	RequestBody  []byte
	ResponseBody []byte
}

func (e BotError) Error() string {
	return fmt.Sprintf(
		"POST %s\n%s\n\nStatus %s\n\n%s",
		e.Path,
		string(e.RequestBody),
		http.StatusText(e.Status),
		string(e.ResponseBody),
	)
}
