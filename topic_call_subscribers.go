package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/jmoiron/sqlx"
)

func callSubscribers(ctx context.Context, db *sqlx.DB, s bot.Service, update bot.Update) error {
	if strings.HasPrefix(update.Message.Text, "/") {
		return nil
	}
	topic := regexp.MustCompile(`#[a-z0-9_]{1,}`).FindString(update.Message.Text)
	if topic == "" {
		return nil
	}

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

	count := 0
	for rows.Next() {
		count++
		var chatID int64
		var userID int64
		var firstName string
		var username string

		err := rows.Scan(&chatID, &userID, &firstName, &username)
		if err != nil {
			log.Print(err)
			return bh.Reply{Text: "vish deu ruim"}
		}

		name := username
		if name == "" {
			name = firstName
		}

		msg += fmt.Sprintf("[%s](tg://user?id=%d) ", name, userID)

		if count%4 == 0 {
			_, err = s.SendMessage(bot.SendMessageParams{
				ChatID:                   chatID,
				Text:                     msg,
				ReplyToMessageID:         update.Message.MessageID,
				AllowSendingWithoutReply: true,
				ParseMode:                "MarkdownV2",
			})
			if err != nil {
				log.Print(err)
				return bh.Reply{Text: "vish deu ruim"}
			}
			msg = ""
		}
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
