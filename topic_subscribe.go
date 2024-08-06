package main

import (
	"context"
	"database/sql"
	"errors"
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

	for _, name := range topics {
		name = strings.TrimSpace(name)
		name = strings.ToLower(name)

		if !regexp.MustCompile(`^#[a-z0-9_]{1,}$`).MatchString(name) {
			return bh.Reply{Text: "topico invalido bb"}
		}

		var n int
		err := db.QueryRowContext(ctx, `
		SELECT
			1
		FROM
			user_topic
		WHERE
			chat_id = $1 AND
			topic = $2
		`, u.Message.Chat.ID, name).Scan(&n)

		if errors.Is(err, sql.ErrNoRows) {
			return bh.Reply{Text: "foi mal ce n pode criar topico"}
		}
		if err != nil {
			log.Print(err)
			return bh.Reply{Text: "vish deu ruim"}
		}

		_, err = db.ExecContext(ctx, `
		INSERT INTO user
			(id, username, first_name)
		VALUES
			($1, $2, $3)
		ON CONFLICT DO UPDATE
		SET
			first_name = excluded.first_name
		`,
			u.Message.From.ID,
			u.Message.From.Username,
			u.Message.From.FirstName,
		)
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
		`)
		if err != nil {
			log.Print(err)
			return bh.Reply{Text: "vish deu ruim"}
		}
	}

	return bh.Reply{Text: "inscrições adicionadas: " + strings.Join(topics, ", ")}
}
