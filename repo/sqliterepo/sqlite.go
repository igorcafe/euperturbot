package sqliterepo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/igoracmelo/euperturbot/repo"
	"github.com/jmoiron/sqlx"
)

type sqliteRepo struct {
	db      *sqlx.DB
	Version int
}

func Open(ctx context.Context, dsn string, dir string) (repo.Repo, error) {
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	repo := &sqliteRepo{
		db,
		0,
	}

	err = repo.migrate(ctx, dir)
	return repo, err
}

func (db *sqliteRepo) DB() *sqlx.DB {
	return db.db
}

func (db *sqliteRepo) migrate(ctx context.Context, dir string) error {
	var version int

	err := db.db.QueryRowContext(ctx, "PRAGMA user_version;").Scan(&version)
	if err != nil {
		return err
	}
	db.Version = version

	fname := filepath.Join(dir, fmt.Sprintf("%d.sql", version+1))
	_, err = os.Stat(fname)
	if errors.Is(err, os.ErrNotExist) {
		// migration does not exist. assume we are on latest
		return nil
	}

	log.Printf("RUN %s", fname)
	query, err := os.ReadFile(fname)
	if err != nil {
		return err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, string(query))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// PRAGMA doesn't support $1, but it is safe to use fmt.Sprintf here
	_, err = db.db.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", version+1))
	if err != nil {
		return err
	}

	db.Version = version + 1
	log.Printf("SUCCESS")

	return db.migrate(ctx, dir)
}

func (db *sqliteRepo) Close() error {
	return db.db.Close()
}
