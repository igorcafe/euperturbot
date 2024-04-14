package sqliterepo

import (
	"context"
	"testing"

	"github.com/igoracmelo/euperturbot/repo"
)

func newDB(t *testing.T) repo.Repo {
	t.Helper()

	db, err := Open(context.TODO(), ":memory:", "./migrations")
	if err != nil {
		t.Fatal(err)
	}

	return db
}
