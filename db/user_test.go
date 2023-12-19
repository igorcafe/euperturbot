package db

import (
	"errors"
	"testing"
)

func TestCreatesAndUpdatesUser(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	user := User{
		ID:        1,
		FirstName: "FirstName",
		Username:  "Username",
	}

	// not stored yet
	_, err := db.FindUser(user.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err - want: %v, got: %v", ErrNotFound, err)
	}

	err = db.SaveUser(user)
	if err != nil {
		t.Fatal(err)
	}

	// must be stored
	loadedUser, err := db.FindUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if *loadedUser != user {
		t.Fatalf("want: %+v, got: %+v", user, *loadedUser)
	}

	user.FirstName = "Newname"
	user.Username = "Newusername"

	err = db.SaveUser(user)
	if err != nil {
		t.Fatal(err)
	}

	// must be updated
	loadedUser, err = db.FindUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if *loadedUser != user {
		t.Fatalf("want: %+v, got: %+v", user, *loadedUser)
	}
}
