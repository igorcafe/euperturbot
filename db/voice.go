package db

import "context"

type Voice struct {
	FileID string `db:"file_id"`
	UserID int64  `db:"user_id"`
}

func (db *DB) SaveVoice(v Voice) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO voice (file_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, v.FileID, v.UserID)
	return err
}

func (db *DB) FindRandomVoice() (*Voice, error) {
	var v Voice
	err := db.db.GetContext(context.TODO(), &v, `SELECT * FROM voice ORDER BY RANDOM() LIMIT 1`)
	return &v, err
}
