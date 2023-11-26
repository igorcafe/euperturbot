package db

import (
	"context"
	"time"
)

type ChatEvent struct {
	ID     int64
	ChatID int64 `db:"chat_id"`
	MsgID  int   `db:"msg_id"`
	Time   time.Time
	Name   string
}

func (db *DB) SaveChatEvent(e ChatEvent) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO event
		(chat_id, msg_id, time, name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO UPDATE SET time = $3
	`, e.ChatID, e.MsgID, e.Time, e.Name)

	return err
}

func (db *DB) FindChatEventsByName(chatID int64, name string) ([]ChatEvent, error) {
	sql := `
		SELECT * FROM event
		WHERE chat_id = $1 AND name = $2
		ORDER BY time DESC
	`
	var events []ChatEvent
	err := db.db.SelectContext(context.TODO(), &events, sql, chatID, name)
	return events, err
}

func (db *DB) DeleteChatEvent(e ChatEvent) error {
	_, err := db.db.ExecContext(context.TODO(), `
		DELETE FROM event
		WHERE chat_id = $1 AND msg_id = $2 AND name = $3
	`, e.ChatID, e.MsgID, e.Name)
	return err
}
