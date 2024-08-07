package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/igoracmelo/euperturbot/bot"
	"github.com/jmoiron/sqlx"
)

func mentionScheduledTopicsWorker(ctx context.Context, db *sqlx.DB, s bot.Service) {
	for {
		func() {
			log.Print("processing scheduled topic")

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				log.Print(err)
				return
			}
			defer tx.Rollback()

			var chatID int64
			var messageID int
			var topic string

			err = db.QueryRowContext(ctx, `
			SELECT
				chat_id,
				message_id,
				topic
			FROM
				scheduled_topic
			WHERE
				status = 'created' AND
				datetime(time) BETWEEN
					datetime('now', '-5 minutes') AND
					datetime('now')
			`).Scan(&chatID, &messageID, &topic)

			if errors.Is(err, sql.ErrNoRows) {
				log.Print("no scheduled topic")
				return
			}
			if err != nil {
				log.Print(err)
				return
			}

			var users []struct {
				ID        int64  `db:"id"`
				FirstName string `db:"first_name"`
				Username  string `db:"username"`
			}

			err = db.SelectContext(ctx, &users, `
			SELECT
				u.id,
				u.first_name,
				u.username
			FROM 
				user u
			JOIN 
				user_topic ut on u.id = ut.user_id
			WHERE
				chat_id = $1 AND
				topic = $2
			`, chatID, topic)
			if err != nil {
				log.Print(err)
				return
			}

			msg := ""
			for i, u := range users {
				name := u.Username
				if name == "" {
					name = u.FirstName
				}
				msg += fmt.Sprintf("[%s](tg://user?id=%d) ", name, u.ID)

				if (i+1)%4 == 0 {
					_, err = s.SendMessage(bot.SendMessageParams{
						ChatID:                   chatID,
						Text:                     msg,
						ReplyToMessageID:         messageID,
						AllowSendingWithoutReply: true,
						ParseMode:                "MarkdownV2",
					})
					if err != nil {
						log.Print(err)
						return
					}
					msg = ""
				}
			}

			if msg != "" {
				_, err = s.SendMessage(bot.SendMessageParams{
					ChatID:                   chatID,
					Text:                     msg,
					ReplyToMessageID:         messageID,
					AllowSendingWithoutReply: true,
					ParseMode:                "MarkdownV2",
				})
				if err != nil {
					log.Print(err)
					return
				}
			}

			_, err = db.ExecContext(ctx, `
			UPDATE
				scheduled_topic
			SET
				status = 'completed'
			WHERE
				chat_id = $1 AND
				message_id = $2 AND
				status = 'created'
			`, chatID, messageID)
			if err != nil {
				log.Print(err)
				return
			}
		}()

		time.Sleep(10 * time.Second)
	}
}
