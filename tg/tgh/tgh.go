package tgh

import (
	"log"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/igoracmelo/euperturbot/tg"
)

type HandlerFunc func(bot *tg.Bot, u tg.Update) error
type Middleware = func(next HandlerFunc) HandlerFunc
type CriteriaFunc func(bot *tg.Bot, u tg.Update) bool

var And = func(criterias ...CriteriaFunc) CriteriaFunc {
	return func(bot *tg.Bot, u tg.Update) bool {
		for _, c := range criterias {
			if !c(bot, u) {
				return false
			}
		}
		return true
	}
}

var AnyMessage CriteriaFunc = func(bot *tg.Bot, u tg.Update) bool {
	return u.Message != nil
}

var AnyText CriteriaFunc = func(bot *tg.Bot, u tg.Update) bool {
	return u.Message != nil && u.Message.Text != ""
}

var AnyCommand CriteriaFunc = func(bot *tg.Bot, u tg.Update) bool {
	if u.Message == nil {
		return false
	}
	return regexp.MustCompile(`^\/\S+`).MatchString(u.Message.Text)
}

var Command = func(cmd string) CriteriaFunc {
	return func(bot *tg.Bot, u tg.Update) bool {
		if u.Message == nil {
			return false
		}
		fields := strings.Fields(u.Message.Text)
		if len(fields) == 0 {
			return false
		}
		first := strings.TrimSuffix(fields[0], "@"+bot.Username)
		return first == "/"+cmd
	}
}

var AnyCallbackQuery CriteriaFunc = func(bot *tg.Bot, u tg.Update) bool {
	return u.CallbackQuery != nil
}

var AnyInlineQuery CriteriaFunc = func(bot *tg.Bot, u tg.Update) bool {
	return u.InlineQuery != nil
}

type UpdateController struct {
	source   <-chan tg.Update
	bot      *tg.Bot
	handlers []struct {
		criteria CriteriaFunc
		fn       HandlerFunc
	}
	middlewares []Middleware
}

func NewUpdateController(bot *tg.Bot, source <-chan tg.Update) *UpdateController {
	return &UpdateController{
		source: source,
		bot:    bot,
	}
}

func (uc *UpdateController) Middleware(mw Middleware, criterias ...CriteriaFunc) {
	criteria := func(bot *tg.Bot, u tg.Update) bool {
		for _, c := range criterias {
			if !c(bot, u) {
				return false
			}
		}
		return true
	}

	_mw := func(hf HandlerFunc) HandlerFunc {
		return func(bot *tg.Bot, u tg.Update) error {
			if criteria(bot, u) {
				return mw(hf)(bot, u)
			} else {
				return hf(bot, u)
			}
		}
	}

	uc.middlewares = append(uc.middlewares, _mw)
}

func (uh *UpdateController) Handle(criteria CriteriaFunc, fn func(bot *tg.Bot, u tg.Update) error) {
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

				if params, ok := err.(tg.SendMessageParams); ok {
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
			break
		}
	}
}
