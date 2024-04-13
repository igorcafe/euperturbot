package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/igoracmelo/euperturbot/util"
)

type Chat struct {
	ID         int64
	Title      string
	EnableCAsk bool
}

var ErrChatActionNotAllowed = errors.New("chat action not allowed")

var ErrNotFound = sql.ErrNoRows

type rawChat struct {
	ID         int64  `db:"id"`
	Title      string `db:"title"`
	EnableCAsk int    `db:"enable_cask"`
}

func (db Repo) SaveChat(ctx context.Context, chat Chat) error {
	rawChat := rawChat{
		ID:         chat.ID,
		Title:      chat.Title,
		EnableCAsk: util.BoolToInt(chat.EnableCAsk),
	}

	_, err := db.db.NamedExecContext(ctx, `
		INSERT INTO chat (
			id,
			title,
			enable_cask
		) VALUES (
			:id,
			:title,
			:enable_cask
		)
		ON CONFLICT DO UPDATE
		SET
			title           = :title,
			enable_cask     = :enable_cask
	`, rawChat)

	return err
}

func (db Repo) FindChat(ctx context.Context, chatID int64) (*Chat, error) {
	var c rawChat

	err := db.db.GetContext(ctx, &c, `
		SELECT * FROM chat
		WHERE id = $1
	`, chatID)

	return &Chat{
		ID:         c.ID,
		Title:      c.Title,
		EnableCAsk: c.EnableCAsk == 1,
	}, err
}

func (db Repo) ChatEnables(ctx context.Context, chatID int64, action string) (bool, error) {
	var iAllow int
	err := db.db.GetContext(ctx, &iAllow, `SELECT enable_`+action+` FROM chat WHERE id = $1`, chatID)
	return iAllow == 1, err
}

func (db Repo) ChatEnable(ctx context.Context, chatID int64, action string) error {
	res, err := db.db.ExecContext(ctx, `
		UPDATE chat
		SET enable_`+action+` = 1
		WHERE id = $1
	`, chatID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNotFound
	}
	return err
}

func (db Repo) ChatDisable(ctx context.Context, chatID int64, action string) error {
	res, err := db.db.ExecContext(ctx, `
		UPDATE chat
		SET enable_`+action+` = 0
		WHERE id = $1
	`, chatID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNotFound
	}
	return err
}
