package sqliterepo

import (
	"context"

	"github.com/igoracmelo/euperturbot/repo"
	"github.com/igoracmelo/euperturbot/util"
)

type rawChat struct {
	ID         int64  `db:"id"`
	Title      string `db:"title"`
	EnableCAsk int    `db:"enable_cask"`
}

func (db sqliteRepo) SaveChat(ctx context.Context, chat repo.Chat) error {
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

func (db sqliteRepo) FindChat(ctx context.Context, chatID int64) (*repo.Chat, error) {
	var c rawChat

	err := db.db.GetContext(ctx, &c, `
		SELECT * FROM chat
		WHERE id = $1
	`, chatID)

	return &repo.Chat{
		ID:         c.ID,
		Title:      c.Title,
		EnableCAsk: c.EnableCAsk == 1,
	}, err
}

func (db sqliteRepo) ChatEnables(ctx context.Context, chatID int64, action string) (bool, error) {
	var iAllow int
	err := db.db.GetContext(ctx, &iAllow, `SELECT enable_`+action+` FROM chat WHERE id = $1`, chatID)
	return iAllow == 1, err
}

func (db sqliteRepo) ChatEnable(ctx context.Context, chatID int64, action string) error {
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
		return repo.ErrNotFound
	}
	return err
}

func (db sqliteRepo) ChatDisable(ctx context.Context, chatID int64, action string) error {
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
		return repo.ErrNotFound
	}
	return err
}
