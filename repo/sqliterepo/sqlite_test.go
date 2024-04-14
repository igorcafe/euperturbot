package sqliterepo

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"
)

func Test_Migrate(t *testing.T) {
	_db, err := Open(context.TODO(), ":memory:", "./migrations")
	if err != nil {
		t.Fatal(err)
	}
	db := _db.(*sqliteRepo)

	// this test has to be updated anytime a new migration is created, on purpose
	if db.Version != 11 {
		t.Fatalf("version - want: %d, got: %d", 11, db.Version)
	}
}
