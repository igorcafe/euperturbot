package bot

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

type Service interface {
	Username() string
	GetMe() (*User, error)
	GetChatMember(params GetChatMemberParams) (*ChatMember, error)
	GetUpdates(params GetUpdatesParams) ([]Update, error)
	GetUpdatesChannel() chan Update
	SendVoice(params SendVoiceParams) (*Message, error)
	SendPoll(params SendPollParams) (*Message, error)
	SendMessage(params SendMessageParams) (*Message, error)
	EditMessageText(params EditMessageTextParams) (*Message, error)
	AnswerInlineQuery(params AnswerInlineQueryParams) error
	SendDocument(params SendDocumentParams) error
}

type service struct {
	retry    util.Retry
	token    string
	username string
	baseURL  string
	client   http.Client
}

var _ Service = &service{}

func NewService(token string) Service {
	return &service{
		token:   token,
		baseURL: "https://api.telegram.org/bot",
		retry: util.Retry{
			MaxAttempts: 3,
			Delay:       time.Second,
		},
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func apiJSONRequest[T any](bot *service, path string, data any) (res Result[T], err error) {
	u := bot.baseURL + bot.token + "/" + path

	var reqBody []byte
	var reqReader io.Reader

	if data != nil {
		reqBody, err = json.Marshal(data)
		if err != nil {
			return
		}
		reqReader = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequest("POST", u, reqReader)
	if err != nil {
		err = errors.New(bot.hideToken(err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	err = bot.retry.Do(func() error {
		resp, err = bot.client.Do(req)
		return err
	})

	if err != nil {
		err = errors.New(bot.hideToken(err.Error()))
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err = bot.respError(resp, reqBody, respBody)
		return
	}
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &res)
	return
}

func (s *service) Username() string {
	return s.username
}

func (s *service) GetMe() (*User, error) {
	res, err := apiJSONRequest[User](s, "getMe", nil)
	s.username = res.Result.Username
	return &res.Result, err
}

func (s *service) GetChatMember(params GetChatMemberParams) (*ChatMember, error) {
	res, err := apiJSONRequest[ChatMember](s, "getMe", params)
	return &res.Result, err
}

func (s *service) GetUpdates(params GetUpdatesParams) ([]Update, error) {
	res, err := apiJSONRequest[[]Update](s, "getUpdates", params)
	return res.Result, err
}

func (s *service) GetUpdatesChannel() chan Update {
	ch := make(chan Update)
	go func() {
		updateID := 0
		for {
			params := GetUpdatesParams{
				Offset:         updateID,
				Timeout:        5,
				AllowedUpdates: []string{"message", "poll", "poll_answer", "callback_query", "inline_query"},
			}
			updates, err := s.GetUpdates(params)
			if err != nil {
				log.Print(s.hideToken(err.Error()))
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

func (s *service) SendVoice(params SendVoiceParams) (*Message, error) {
	res, err := apiJSONRequest[Message](s, "sendVoice", params)
	return &res.Result, err
}

func (s *service) SendPoll(params SendPollParams) (*Message, error) {
	res, err := apiJSONRequest[Message](s, "sendPoll", params)
	return &res.Result, err
}

func (s *service) SendMessage(params SendMessageParams) (*Message, error) {
	res, err := apiJSONRequest[Message](s, "sendMessage", params)
	return &res.Result, err
}

func (s *service) EditMessageText(params EditMessageTextParams) (*Message, error) {
	res, err := apiJSONRequest[Message](s, "editMessageText", params)
	return &res.Result, err
}

func (s *service) AnswerInlineQuery(params AnswerInlineQueryParams) error {
	_, err := apiJSONRequest[bool](s, "answerInlineQuery", params)
	return err
}

func (s *service) SendDocument(params SendDocumentParams) error {
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

	u := s.baseURL + s.token + "/sendDocument"
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := s.client.Do(req)
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

func (s *service) hideToken(str string) string {
	return strings.ReplaceAll(str, s.token, "<token>")
}

func (s *service) respError(resp *http.Response, reqBody []byte, respBody []byte) error {
	err := BotError{}
	err.Path = s.hideToken(resp.Request.URL.String())
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
