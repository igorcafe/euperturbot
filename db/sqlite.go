package db

import (
	"context"
	"errors"
	"io"
	"os/exec"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	db *sqlx.DB
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
	}, nil
}

func (db *DB) Migrate(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "bash", "./migrate.sh")
	// cmd.Dir =
	b, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Join(err, errors.New(string(b)))
	}
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}
