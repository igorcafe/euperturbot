package repo

import "context"

type Voice struct {
	FileID string `db:"file_id"`
	UserID int64  `db:"user_id"`
	ChatID int64  `db:"chat_id"`
}

func (db *Repo) SaveVoice(v Voice) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO voice
			(file_id, user_id, chat_id)
		VALUES
			($1, $2, $3)
		ON CONFLICT DO UPDATE
		SET
			user_id = $2,
			chat_id = $3
	`, v.FileID, v.UserID, v.ChatID)
	return err
}

func (db *Repo) FindRandomVoice(chatID int64) (*Voice, error) {
	var v Voice
	err := db.db.GetContext(context.TODO(), &v, `
		SELECT * FROM voice
		WHERE chat_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`, chatID)
	return &v, err
}
