package handler

import (
	"context"
	"errors"

	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/tg"
)

func (h Handler) StartedMiddleware() tg.Middleware {
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
		if u.Message.Chat.Type == "private" {
			return next(bot, u)
		}

		if u.Message.From.ID == h.Config.GodID {
			return next(bot, u)
		}

		member, err := bot.GetChatMember(tg.GetChatMemberParams{
			ChatID: u.Message.Chat.ID,
			UserID: u.Message.From.ID,
		})
		if err != nil {
			return err
		}

		if member.Status == "creator" || member.Status == "administrator" {
			return next(bot, u)
		}

		return tg.SendMessageParams{
			ReplyToMessageID: u.Message.MessageID,
			Text:             "você não tem permissão para isso",
		}
	}
}
