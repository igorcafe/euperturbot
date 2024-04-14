package sqliterepo

import (
	"context"
	"time"

	"github.com/igoracmelo/euperturbot/repo"
)

func (db *sqliteRepo) SaveMessage(ctx context.Context, msg repo.Message) error {
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

func (db *sqliteRepo) FindMessage(ctx context.Context, chatID int64, msgID int) (repo.Message, error) {
	var msg repo.Message
	err := db.db.GetContext(ctx, &msg, `
		SELECT * FROM message
		WHERE chat_id = $1 AND id = $2
	`, chatID, msgID)
	return msg, err
}

func (db *sqliteRepo) FindMessagesBeforeDate(ctx context.Context, chatID int64, date time.Time, count int) ([]repo.Message, error) {
	msgs := []repo.Message{}
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

func (db *sqliteRepo) FindMessageThread(ctx context.Context, chatID int64, msgID int) ([]repo.Message, error) {
	var msgs []repo.Message

	err := db.db.SelectContext(ctx, &msgs, `
	WITH RECURSIVE replies(id, reply_to_message_id) AS (
		SELECT id, reply_to_message_id
		FROM message
		WHERE
			chat_id = $1 AND
			id = $2

		UNION ALL

		SELECT m.id, m.reply_to_message_id
		FROM message m
		INNER JOIN replies r
		ON
		 	m.chat_id = $1 AND
			m.id = r.reply_to_message_id
	)

	SELECT m.* FROM message m
	JOIN replies r ON m.id = r.id
	ORDER BY id;
	`, chatID, msgID)

	return msgs, err
}
