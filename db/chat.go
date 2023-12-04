package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/igoracmelo/euperturbot/util"
)

type Chat struct {
	ID                int64
	Title             string
	AllowSaveMessages bool
	AllowGPT          bool
	AllowAudio        bool
}

var ErrChatActionNotAllowed = errors.New("chat action not allowed")

type rawChat struct {
	ID                int64  `db:"id"`
	Title             string `db:"title"`
	AllowSaveMessages int    `db:"allow_save_msgs"`
	AllowGPT          int    `db:"allow_gpt"`
	AllowAudio        int    `db:"allow_audio"`
}

func (db DB) SaveChat(ctx context.Context, chat Chat) error {
	rawChat := rawChat{
		ID:                chat.ID,
		Title:             chat.Title,
		AllowSaveMessages: util.BoolToInt(chat.AllowSaveMessages),
		AllowGPT:          util.BoolToInt(chat.AllowGPT),
		AllowAudio:        util.BoolToInt(chat.AllowAudio),
	}

	_, err := db.db.NamedExecContext(ctx, `
		INSERT INTO chat (
			id,
			title,
			allow_save_msgs,
			allow_gpt,
			allow_audio
		) VALUES (
			:id,
			:title,
			:allow_save_msgs,
			:allow_gpt,
			:allow_audio
		)
		ON CONFLICT DO UPDATE
		SET
			title = :title,
			allow_save_msgs = :allow_save_msgs,
			allow_gpt = :allow_gpt,
			allow_audio = :allow_audio
	`, rawChat)

	return err
}

func (db DB) FindChat(ctx context.Context, chatID int64) (*Chat, error) {
	var c rawChat

	err := db.db.GetContext(ctx, &c, `
		SELECT * FROM chat
		WHERE id = $1
	`, chatID)

	if errors.Is(err, sql.ErrNoRows) {
		return &Chat{
			ID: chatID,
		}, nil
	}

	return &Chat{
		ID:                c.ID,
		Title:             c.Title,
		AllowSaveMessages: c.AllowSaveMessages == 1,
		AllowGPT:          c.AllowGPT == 1,
		AllowAudio:        c.AllowAudio == 1,
	}, err
}

func (db DB) chatAllows(ctx context.Context, chatID int64, action string) (bool, error) {
	var iAllow int
	err := db.db.GetContext(ctx, &iAllow, `SELECT allow_`+action+` FROM chat WHERE id = $1`, chatID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return iAllow == 1, err
}
