package sqliterepo

import (
	"context"
	"testing"

	"github.com/igoracmelo/euperturbot/repo"
)

func newDB(t *testing.T) repo.Repo {
	t.Helper()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Migrate(context.Background(), "./migrations")
	if err != nil {
		t.Fatal(err)
	}

	return db
}
