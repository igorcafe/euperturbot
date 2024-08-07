package main

import (
	"context"
	"fmt"
	"log"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/jmoiron/sqlx"
)

func listByUser(ctx context.Context, db *sqlx.DB, update bot.Update) error {
	rows, err := db.QueryContext(ctx, `
	WITH my_topics AS (
		SELECT
			*
		FROM 
			user_topic
		WHERE
			chat_id = $1 AND
			user_id = $2
	)
	SELECT
		COUNT(*),
		topic
	FROM user_topic
	WHERE
		topic IN (SELECT topic FROM my_topics) AND
		chat_id = $1
	GROUP BY
		topic
	`, update.Message.Chat.ID, update.Message.From.ID)
	if err != nil {
		log.Print(err)
		return bh.Reply{Text: "vish deu ruim"}
	}
	defer rows.Close()

	msg := ""
	for rows.Next() {
		var count int
		var topic string

		err := rows.Scan(&count, &topic)
		if err != nil {
			return bh.Reply{Text: "vish deu ruim"}
		}

		msg += fmt.Sprintf("(%02d) %s\n", count, topic)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return bh.Reply{Text: "vish deu ruim"}
	}

	if msg == "" {
		return bh.Reply{Text: "voce nao ta inscrito em nenhum topico"}
	}

	msg = "seus topicos:\n" + msg
	return bh.Reply{
		Text: msg,
	}
}
