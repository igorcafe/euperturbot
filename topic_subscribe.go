package main

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/jmoiron/sqlx"
)

type subscribeToTopicOpts struct {
	botSvc bot.Service
	update bot.Update
	db     *sqlx.DB
}

func subscribeToTopic(ctx context.Context, db *sqlx.DB, b bot.Service, u bot.Update) error {
	if u.Message.From.IsBot {
		return bh.Reply{Text: "nao pode inscrever bot"}
	}

	topics := strings.Fields(u.Message.Text)
	if len(topics) <= 1 {
		return bh.Reply{Text: "cadê os tópicos bb?"}
	}
	topics = topics[1:]

	chatID := u.Message.Chat.ID
	userID := u.Message.From.ID
	username := u.Message.From.Username
	firstName := u.Message.From.FirstName

	if u.Message.ReplyToMessage != nil {
		msg := u.Message.ReplyToMessage
		userID = msg.From.ID
		username = msg.From.Username
		firstName = msg.From.FirstName
	}

	for _, name := range topics {
		name = strings.TrimSpace(name)
		name = strings.ToLower(name)

		if !regexp.MustCompile(`^#[a-z0-9_]{1,}$`).MatchString(name) {
			return bh.Reply{Text: "topico invalido bb"}
		}

		var err error

		// var n int
		// err := db.QueryRowContext(ctx, `
		// SELECT
		// 	1
		// FROM
		// 	user_topic
		// WHERE
		// 	chat_id = $1 AND
		// 	topic = $2
		// `, chatID, name).Scan(&n)

		// if errors.Is(err, sql.ErrNoRows) {
		// 	return bh.Reply{Text: "foi mal ce n pode criar topico"}
		// }
		// if err != nil {
		// 	log.Print(err)
		// 	return bh.Reply{Text: "vish deu ruim"}
		// }

		_, err = db.ExecContext(ctx, `
		INSERT INTO user
			(id, username, first_name)
		VALUES
			($1, $2, $3)
		ON CONFLICT DO UPDATE
		SET
			first_name = excluded.first_name
		`, userID, username, firstName)
		if err != nil {
			log.Print(err)
			return bh.Reply{Text: "vish deu ruim"}
		}

		_, err = db.ExecContext(ctx, `
		INSERT INTO user_topic
			(chat_id, user_id, topic)
		VALUES
			($1, $2, $3)
		ON CONFLICT DO NOTHING
		`, chatID, userID, name)
		if err != nil {
			log.Print(err)
			return bh.Reply{Text: "vish deu ruim"}
		}
	}

	return bh.Reply{Text: "inscrições adicionadas: " + strings.Join(topics, ", ")}
}
