package db

import (
	"context"
	"testing"

	_ "github.com/glebarez/go-sqlite"
)

func Test_Migrate(t *testing.T) {
	db, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Migrate(context.TODO(), "./migrations")
	if err != nil {
		t.Fatal(err)
	}

	// this test has to be updated anytime a new migration is created, on purpose
	if db.Version != 11 {
		t.Fatalf("version - want: %d, got: %d", 11, db.Version)
	}
}
