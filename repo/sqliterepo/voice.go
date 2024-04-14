package sqliterepo

import (
	"context"

	"github.com/igoracmelo/euperturbot/repo"
)

func (db *sqliteRepo) SaveVoice(v repo.Voice) error {
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

func (db *sqliteRepo) FindRandomVoice(chatID int64) (*repo.Voice, error) {
	var v repo.Voice
	err := db.db.GetContext(context.TODO(), &v, `
		SELECT * FROM voice
		WHERE chat_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`, chatID)
	return &v, err
}
