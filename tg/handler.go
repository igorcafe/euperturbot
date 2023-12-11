package tg

import (
	"log"
	"regexp"
	"runtime/debug"
	"strings"
)

type HandlerFunc func(bot *Bot, u Update) error
type Middleware = func(next HandlerFunc) HandlerFunc
type CriteriaFunc func(u Update) bool

var AnyMessage CriteriaFunc = func(u Update) bool {
	return u.Message != nil
}

var AnyCommand CriteriaFunc = func(u Update) bool {
	if u.Message == nil {
		return false
	}
	return regexp.MustCompile(`^\/\S+`).MatchString(u.Message.Text)
}

type UpdateController struct {
	source   <-chan Update
	bot      *Bot
	handlers []struct {
		criteria CriteriaFunc
		fn       HandlerFunc
	}
	middlewares []Middleware
}

func NewUpdateController(bot *Bot, source <-chan Update) *UpdateController {
	return &UpdateController{
		source: source,
		bot:    bot,
	}
}

func (uc *UpdateController) Middleware(mw Middleware, criterias ...CriteriaFunc) {
	criteria := func(u Update) bool {
		for _, c := range criterias {
			if !c(u) {
				return false
			}
		}
		return true
	}

	_mw := func(hf HandlerFunc) HandlerFunc {
		return func(bot *Bot, u Update) error {
			if criteria(u) {
				return mw(hf)(bot, u)
			} else {
				return hf(bot, u)
			}
		}
	}

	uc.middlewares = append(uc.middlewares, _mw)
}

func (uh *UpdateController) Handle(criteria CriteriaFunc, fn func(bot *Bot, u Update) error) {
	uh.handlers = append(uh.handlers, struct {
		criteria CriteriaFunc
		fn       HandlerFunc
	}{
		criteria,
		fn,
	})
}

func (uh *UpdateController) HandleMessage(handler func(bot *Bot, u Update) error) {
	uh.Handle(AnyMessage, handler)
}

func (uh *UpdateController) HandleText(handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		return u.Message != nil && strings.TrimSpace(u.Message.Text) != ""
	}
	uh.Handle(criteria, handler)
}

func (uh *UpdateController) HandleTextEqual(texts []string, handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		if u.Message == nil {
			return false
		}
		for _, t := range texts {
			if u.Message.Text == t {
				return true
			}
		}
		return false
	}

	uh.Handle(criteria, handler)
}

func (uh *UpdateController) HandleCommand(cmd string, handler func(bot *Bot, u Update) error) {
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

func (uh *UpdateController) HandlePollAnswer(handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		return u.PollAnswer != nil
	}
	uh.Handle(criteria, handler)
}

func (uh *UpdateController) HandleCallbackQuery(handler func(bot *Bot, u Update) error) {
	criteria := func(u Update) bool {
		return u.CallbackQuery != nil
	}
	uh.Handle(criteria, handler)
}

func (uh *UpdateController) HandleInlineQuery(handler func(bot *Bot, u Update) error) {
	uh.Handle(func(u Update) bool {
		return u.InlineQuery != nil
	}, handler)
}

func (uh *UpdateController) Start() {
	limit := make(chan struct{}, 10)
	for update := range uh.source {
		for _, handler := range uh.handlers {
			handler := handler
			update := update
			if !handler.criteria(update) {
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
			break
		}
	}
}
