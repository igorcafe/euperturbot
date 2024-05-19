package bothandler

import (
	"errors"
	"log"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/igoracmelo/euperturbot/bot"
)

type HandlerFunc func(s bot.Service, u bot.Update) error
type Middleware = func(next HandlerFunc) HandlerFunc
type CriteriaFunc func(s bot.Service, u bot.Update) bool

type Reply struct {
	Text      string
	ParseMode string
}

func (r Reply) Error() string {
	return ""
}

var And = func(criterias ...CriteriaFunc) CriteriaFunc {
	return func(s bot.Service, u bot.Update) bool {
		for _, c := range criterias {
			if !c(s, u) {
				return false
			}
		}
		return true
	}
}

var AnyMessage CriteriaFunc = func(s bot.Service, u bot.Update) bool {
	return u.Message != nil
}

var AnyText CriteriaFunc = func(s bot.Service, u bot.Update) bool {
	return u.Message != nil && u.Message.Text != ""
}

var AnyCommand CriteriaFunc = func(s bot.Service, u bot.Update) bool {
	if u.Message == nil {
		return false
	}
	return regexp.MustCompile(`^\/\S+`).MatchString(u.Message.Text)
}

var Command = func(cmd string) CriteriaFunc {
	return func(s bot.Service, u bot.Update) bool {
		if u.Message == nil {
			return false
		}
		fields := strings.Fields(u.Message.Text)
		if len(fields) == 0 {
			return false
		}
		first := strings.TrimSuffix(fields[0], "@"+s.Username())
		return first == "/"+cmd
	}
}

var AnyCallbackQuery CriteriaFunc = func(s bot.Service, u bot.Update) bool {
	return u.CallbackQuery != nil
}

var AnyInlineQuery CriteriaFunc = func(s bot.Service, u bot.Update) bool {
	return u.InlineQuery != nil
}

type UpdateController struct {
	source   <-chan bot.Update
	bot      bot.Service
	handlers []struct {
		criteria CriteriaFunc
		fn       HandlerFunc
	}
	middlewares []Middleware
}

func NewUpdateHandler(s bot.Service, source <-chan bot.Update) *UpdateController {
	return &UpdateController{
		source: source,
		bot:    s,
	}
}

func (uc *UpdateController) Middleware(mw Middleware, criterias ...CriteriaFunc) {
	criteria := func(s bot.Service, u bot.Update) bool {
		for _, c := range criterias {
			if !c(s, u) {
				return false
			}
		}
		return true
	}

	_mw := func(hf HandlerFunc) HandlerFunc {
		return func(s bot.Service, u bot.Update) error {
			if criteria(s, u) {
				return mw(hf)(s, u)
			} else {
				return hf(s, u)
			}
		}
	}

	uc.middlewares = append(uc.middlewares, _mw)
}

func (uh *UpdateController) Handle(criteria CriteriaFunc, fn func(s bot.Service, u bot.Update) error) {
	uh.handlers = append(uh.handlers, struct {
		criteria CriteriaFunc
		fn       HandlerFunc
	}{
		criteria,
		fn,
	})
}

func (uh *UpdateController) Start() {
	limit := make(chan struct{}, 10)
	for update := range uh.source {
		for _, handler := range uh.handlers {
			handler := handler
			update := update
			if !handler.criteria(uh.bot, update) {
				continue
			}

			fn := handler.fn
			for _, mw := range uh.middlewares {
				fn = mw(handler.fn)
			}

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
				log.Print(update)
				err := fn(uh.bot, update)

				var reply Reply
				if errors.As(err, &reply) {
					_, err = uh.bot.SendMessage(bot.SendMessageParams{
						ChatID:                   update.Message.Chat.ID,
						ReplyToMessageID:         update.Message.MessageID,
						AllowSendingWithoutReply: true,
						Text:                     reply.Text,
						ParseMode:                reply.ParseMode,
					})
				}
				if err != nil {
					log.Print(err)
				}
			}()
			break
		}
	}
}
