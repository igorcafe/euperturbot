package handler

import (
	"context"
	"errors"

	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/tg"
)

func (h Handler) EnsureStarted() tg.Middleware {
	return func(next tg.HandlerFunc) tg.HandlerFunc {
		return func(bot *tg.Bot, u tg.Update) error {
			if u.Message.Text == "/start" {
				return next(bot, u)
			}

			_, err := h.DB.FindChat(context.TODO(), u.Message.Chat.ID)
			if errors.Is(err, db.ErrNotFound) {
				// chat not /start'ed. ignore
				return nil
			}

			return next(bot, u)
		}
	}
}

func (h Handler) IgnoreForwardedCommand() tg.Middleware {
	return func(next tg.HandlerFunc) tg.HandlerFunc {
		return func(bot *tg.Bot, u tg.Update) error {
			if u.Message.ForwardSenderName != "" || u.Message.FowardFrom != nil {
				return nil
			}
			return next(bot, u)
		}
	}
}

func (h Handler) RequireGod(next tg.HandlerFunc) tg.HandlerFunc {
	return func(bot *tg.Bot, u tg.Update) error {
		if u.Message.Chat.Type == "private" && u.Message.From.ID == h.Config.GodID {
			return next(bot, u)
		}

		return tg.SendMessageParams{
			ReplyToMessageID: u.Message.MessageID,
			Text:             "você não tem permissão para isso",
		}
	}
}

func (h Handler) RequireAdmin(next tg.HandlerFunc) tg.HandlerFunc {
	return func(bot *tg.Bot, u tg.Update) error {
		isAdmin, err := h.isAdmin(bot, u)
		if err != nil {
			return err
		}
		if !isAdmin {
			return tg.SendMessageParams{
				ReplyToMessageID: u.Message.MessageID,
				Text:             "você não tem permissão para isso",
			}
		}

		return next(bot, u)
	}
}
