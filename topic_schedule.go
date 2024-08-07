package main

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/jmoiron/sqlx"
)

func scheduleMentionSubscribers(ctx context.Context, db *sqlx.DB, s bot.Service, update bot.Update) error {
	fields := strings.Fields(update.Message.Text)
	if len(fields) != 3 {
		return bh.Reply{Text: "formato: /agenda #topico 1h"}
	}

	topic := fields[1]
	deltaStr := fields[2]

	delta, err := time.ParseDuration(deltaStr)
	if err != nil {
		return bh.Reply{Text: "formato: /agenda #topico 1h"}
	}

	if delta < time.Minute {
		return bh.Reply{Text: "o intervalo precisa ser +1min"}
	}

	if delta > 24*time.Hour {
		return bh.Reply{Text: "o intervalo maximo eh de 24h"}
	}

	if !regexp.MustCompile(`^#[a-z0-9_]{1,}$`).MatchString(topic) {
		return bh.Reply{Text: "formato: /agenda #topico 1h"}
	}

	chatID := update.Message.Chat.ID
	msgID := update.Message.MessageID
	t := time.Now().UTC().Add(delta)
	timeStr := t.Format("2006-01-02 15:04")

	_, err = db.ExecContext(ctx, `
	INSERT INTO scheduled_topic
		(chat_id, message_id, topic, time)
	VALUES
		($1, $2, $3, $4)
	`, chatID, msgID, topic, timeStr)
	if err != nil {
		log.Print(err)
		return bh.Reply{Text: "vish deu ruim"}
	}

	log.Printf("scheduled topic: chat_id = %d, message_id = %d, topic = %s, time = %s", chatID, msgID, topic, timeStr)
	return bh.Reply{
		Text: "agendado para daqui a " + delta.String(),
	}
}
