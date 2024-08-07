package main

import (
	"context"
	"log"
	"strings"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/jmoiron/sqlx"
)

func unsubscribe(ctx context.Context, db *sqlx.DB, update bot.Update) error {
	topics := strings.Split(update.Message.Text, " ")
	if len(topics) <= 1 {
		return bh.Reply{Text: "cade os topicos fofa"}
	}
	topics = topics[1:]

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	for _, topic := range topics {
		_, err := db.ExecContext(ctx, `
		DELETE FROM
			user_topic
		WHERE
			chat_id = $1 AND
			user_id = $2 AND
			topic = $3
		`, chatID, userID, topic)
		if err != nil {
			log.Print(err)
			return bh.Reply{Text: "vish deu ruim"}
		}
	}

	return bh.Reply{Text: "feito meu querido"}
}
