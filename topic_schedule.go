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
		return bh.Reply{Text: "formato: /agenda #topico 17:00"}
	}

	topic := fields[1]
	timeStr := fields[2]

	date := time.Now().Format("2006-01-02")
	t, err := time.ParseInLocation("2006-01-02 15:04", date+" "+timeStr, time.Local)
	if err != nil {
		return bh.Reply{Text: "formato: /agenda #topico 17:00"}
	}

	if t.Before(time.Now()) {
		t = t.AddDate(0, 0, 1)
	}

	if !regexp.MustCompile(`^#[a-z0-9_]{1,}$`).MatchString(topic) {
		return bh.Reply{Text: "formato: /agenda #topico 17:00"}
	}

	chatID := update.Message.Chat.ID
	msgID := update.Message.MessageID
	timeStr = t.Format("2006-01-02 15:04")

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
		Text: "agendado para " + t.Format("15:04, dia 02"),
	}
}
