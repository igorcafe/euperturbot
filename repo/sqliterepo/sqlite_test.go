package sqliterepo

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"
)

func Test_Migrate(t *testing.T) {
	_db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db := _db.(*sqliteRepo)

	err = db.Migrate(context.TODO(), "./migrations")
	if err != nil {
		t.Fatal(err)
	}

	// this test has to be updated anytime a new migration is created, on purpose
	if db.Version != 11 {
		t.Fatalf("version - want: %d, got: %d", 11, db.Version)
	}
}
