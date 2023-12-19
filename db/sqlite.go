package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	db      *sqlx.DB
	Version int
}

var _ io.Closer = (*DB)(nil)

func NewSqlite(dsn string) (*DB, error) {
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DB{
		db,
		0,
	}, nil
}

func (db *DB) Migrate(ctx context.Context, dir string) error {
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

	return db.Migrate(ctx, dir)
}

func (db *DB) Close() error {
	return db.db.Close()
}
