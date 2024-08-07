package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/jmoiron/sqlx"
)

func mentionSubscribers(ctx context.Context, db *sqlx.DB, s bot.Service, update bot.Update) error {
	if strings.HasPrefix(update.Message.Text, "/") {
		return nil
	}
	topics := regexp.MustCompile(`#[a-z0-9_]{1,}`).FindAllString(update.Message.Text, -1)
	if len(topics) == 0 {
		return nil
	}

	topicsStr := ""
	for i, topic := range topics {
		if i != 0 {
			topicsStr += ","
		}
		topicsStr += "'" + topic + "'"
	}

	rows, err := db.QueryContext(ctx, `
	SELECT DISTINCT
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
		ut.topic IN (`+topicsStr+`)
	`, update.Message.Chat.ID)
	if err != nil {
		log.Print(err)
		return bh.Reply{Text: "vish deu ruim"}
	}
	defer rows.Close()

	msg := ""
	for count := 1; rows.Next(); count++ {
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
			if count/4 > 1 {
				time.Sleep(time.Second)
			}
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
