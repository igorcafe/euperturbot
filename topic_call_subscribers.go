package main

import (
	"context"
	"fmt"
	"log"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/jmoiron/sqlx"
)

func callSubscribers(ctx context.Context, db *sqlx.DB, update bot.Update, topic string) error {
	rows, err := db.QueryContext(ctx, `
	SELECT
		ut.chat_id,
		u.id,
		u.first_name,
		u.username
	FROM
		user u 
	JOIN
		user_topic ut ON u.id = ut.user_id
	WHERE
		ut.chat_id = $1 AND
		ut.topic = $2
	`, update.Message.Chat.ID, topic)
	if err != nil {
		log.Print(err)
		return bh.Reply{Text: "vish deu ruim"}
	}
	defer rows.Close()

	msg := ""

	for rows.Next() {
		var chatID int64
		var userID int64
		var firstName string
		var username string

		err := rows.Scan(&chatID, &userID, &firstName, &username)
		if err != nil {
			return bh.Reply{Text: "vish deu ruim"}
		}

		name := username
		if name == "" {
			name = firstName
		}

		msg += fmt.Sprintf("[%s](tg://user?id=%d) ", name, userID)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return bh.Reply{Text: "vish deu ruim"}
	}

	if msg == "" {
		return nil
	}

	return bh.Reply{
		Text:      msg,
		ParseMode: "MarkdownV2",
	}
}
