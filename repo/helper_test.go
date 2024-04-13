package repo

import (
	"context"
	"testing"
)

func newDB(t *testing.T) Repo {
	t.Helper()

	db, err := OpenSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Migrate(context.Background(), "./migrations")
	if err != nil {
		t.Fatal(err)
	}

	return *db
}
