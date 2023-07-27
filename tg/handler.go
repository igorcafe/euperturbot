package tg

import (
	"log"
	"runtime/debug"
	"strings"
)

type HandlerFunc func(bot *Bot, u Update) error
type CriteriaFunc func(u Update) bool

type UpdateHandler struct {
	source   <-chan Update
	bot      *Bot
	handlers []struct {
		criteria CriteriaFunc
		handler  HandlerFunc
	}
}

func NewUpdateHandler(bot *Bot, source <-chan Update) *UpdateHandler {
	return &UpdateHandler{
		source: source,
		bot:    bot,
	}
}

func (uh *UpdateHandler) Handle(criteria CriteriaFunc, handler func(bot *Bot, u Update) error) {
	uh.handlers = append(uh.handlers, struct {
		criteria CriteriaFunc
		handler  HandlerFunc
	}{
		criteria,
		handler,
	})
}

func (uh *UpdateHandler) HandleText(handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		return u.Message != nil && u.Message.Text != ""
	}
	uh.Handle(criteria, handler)
}

func (uh *UpdateHandler) HandleCommand(cmd string, handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		if u.Message == nil {
			return false
		}
		fields := strings.Fields(u.Message.Text)
		if len(fields) == 0 {
			return false
		}
		first := strings.TrimSuffix(fields[0], "@"+uh.bot.Username)
		return first == "/"+cmd
	}
	uh.Handle(criteria, handler)
}

func (uh *UpdateHandler) HandlePollAnswer(handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		return u.PollAnswer != nil
	}
	uh.Handle(criteria, handler)
}

func (uh *UpdateHandler) HandleCallbackQuery(handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		return u.CallbackQuery != nil
	}
	uh.Handle(criteria, handler)
}

func (uh *UpdateHandler) Start() {
	limit := make(chan struct{}, 10)
	for update := range uh.source {
		for _, handler := range uh.handlers {
			handler := handler
			update := update
			if handler.criteria(update) {
				limit <- struct{}{}
				go func() {
					defer func() {
						if r := recover(); r != nil {
							log.Print("handler panic recovered: ", r)
							debug.PrintStack()
						}
					}()
					defer func() {
						<-limit
					}()
					err := handler.handler(uh.bot, update)

					if params, ok := err.(SendMessageParams); ok {
						params.ChatID = update.Message.Chat.ID
						params.ReplyToMessageID = update.Message.MessageID

						_, err = uh.bot.SendMessage(params)
						if err != nil {
							log.Print(err)
						}
						return
					}

					if err != nil {
						log.Print(err)
					}
				}()
			}
		}
	}
}
