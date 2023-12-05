package db

import (
	"context"
	"time"
)

type Message struct {
	ID               int
	ChatID           int64 `db:"chat_id"`
	Text             string
	Date             time.Time
	UserID           int64  `db:"user_id"`
	UserName         string `db:"user_name"`
	ReplyToMessageID int    `db:"reply_to_message_id"`
}

func (db *DB) SaveMessage(msg Message) error {
	if len(msg.Text) > 500 {
		msg.Text = msg.Text[:497] + "..."
	}

	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO message (
			id,
			chat_id,
			date,
			text,
			user_id,
			user_name,
			reply_to_message_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO UPDATE SET text = $4
	`,
		msg.ID,
		msg.ChatID,
		msg.Date,
		msg.Text,
		msg.UserID,
		msg.UserName,
		msg.ReplyToMessageID,
	)
	return err
}

func (db *DB) FindMessagesByDate(chatID int64, date time.Time) ([]Message, error) {
	msgs := []Message{}
	err := db.db.SelectContext(context.TODO(), &msgs, `
		SELECT *
		FROM message
		WHERE
			chat_id = $1 AND
			date >= DATE($2) AND
			date < DATE($2, '+1 day')
		ORDER BY date
	`, chatID, date)

	return msgs, err
}

func (db *DB) FindMessagesAfterDate(chatID int64, date time.Time, count int) ([]Message, error) {
	msgs := []Message{}
	err := db.db.SelectContext(context.TODO(), &msgs, `
		SELECT *
		FROM message
		WHERE
			chat_id = $1 AND
			date >= $2
		ORDER BY date
		LIMIT $3
	`, chatID, date, count)

	return msgs, err
}

func (db *DB) FindMessagesBeforeDate(chatID int64, date time.Time, count int) ([]Message, error) {
	msgs := []Message{}
	err := db.db.SelectContext(context.TODO(), &msgs, `
	 	SELECT * FROM (
			SELECT *
			FROM message
			WHERE
				chat_id = $1 AND
				date <= $2
			ORDER BY date DESC
			LIMIT $3
		)
		ORDER BY date ASC
	`, chatID, date, count)

	return msgs, err
}
