package handler

import (
	"context"
	"errors"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/igoracmelo/euperturbot/db"
)

func (h Handler) EnsureStarted() bh.Middleware {
	return func(next bh.HandlerFunc) bh.HandlerFunc {
		return func(s bot.Service, u bot.Update) error {
			if u.Message.Text == "/start" {
				return next(s, u)
			}

			_, err := h.DB.FindChat(context.TODO(), u.Message.Chat.ID)
			if errors.Is(err, db.ErrNotFound) {
				// chat not /start'ed. ignore
				return nil
			}

			return next(s, u)
		}
	}
}

func (h Handler) IgnoreForwardedCommand() bh.Middleware {
	return func(next bh.HandlerFunc) bh.HandlerFunc {
		return func(s bot.Service, u bot.Update) error {
			if u.Message.ForwardSenderName != "" || u.Message.FowardFrom != nil {
				return nil
			}
			return next(s, u)
		}
	}
}

func (h Handler) RequireGod(next bh.HandlerFunc) bh.HandlerFunc {
	return func(s bot.Service, u bot.Update) error {
		if u.Message.Chat.Type == "private" && u.Message.From.ID == h.Config.GodID {
			return next(s, u)
		}

		return bh.Reply{
			Text: "você não tem permissão para isso",
		}
	}
}

func (h Handler) RequireAdmin(next bh.HandlerFunc) bh.HandlerFunc {
	return func(s bot.Service, u bot.Update) error {
		isAdmin, err := h.isAdmin(s, u)
		if err != nil {
			return err
		}
		if !isAdmin {
			return bh.Reply{
				Text: "você não tem permissão para isso",
			}
		}

		return next(s, u)
	}
}
