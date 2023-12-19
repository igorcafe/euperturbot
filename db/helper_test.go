package db

import (
	"context"
	"testing"
)

func newDB(t *testing.T) DB {
	t.Helper()

	db, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Migrate(context.Background(), "./migrations")
	if err != nil {
		t.Fatal(err)
	}

	return *db
}
