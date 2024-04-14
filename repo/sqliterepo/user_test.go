package sqliterepo

import (
	"errors"
	"testing"

	"github.com/igoracmelo/euperturbot/repo"
)

func TestCreateAndUpdateUser(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	user := repo.User{
		ID:        1,
		FirstName: "FirstName",
		Username:  "Username",
	}

	// not stored yet
	_, err := db.FindUser(user.ID)
	if !errors.Is(err, repo.ErrNotFound) {
		t.Fatalf("err - want: %v, got: %v", repo.ErrNotFound, err)
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

func TestCreateAndDeleteUserTopic(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	user := repo.User{
		ID:       1,
		Username: "player",
	}

	userTopic := repo.UserTopic{
		ID:     1,
		ChatID: 1,
		UserID: user.ID,
		Topic:  "brawlhalla",
	}

	// ensure topic does not exist yet
	exists, err := db.ExistsChatTopic(userTopic.ChatID, userTopic.Topic)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatalf("topic not stored yet, but got exists = true")
	}

	// store user
	err = db.SaveUser(user)
	if err != nil {
		t.Fatal(err)
	}

	// store topic
	err = db.SaveUserTopic(userTopic)
	if err != nil {
		t.Fatal(err)
	}

	// double storing the topic should do nothing
	err = db.SaveUserTopic(userTopic)
	if err != nil {
		t.Fatal(err)
	}

	// ensure topic exists
	exists, err = db.ExistsChatTopic(userTopic.ChatID, userTopic.Topic)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("want topic to exist on db")
	}

	// ensure user is subscribed to topic
	users, err := db.FindUsersByTopic(userTopic.ChatID, userTopic.Topic)
	if err != nil {
		t.Fatal(err)
	}
	if users[0] != user {
		t.Fatalf("user - want: %+v, got: %+v", user, users[0])
	}

	// ensure topic is found on user topics
	topics, err := db.FindUserChatTopics(userTopic.ChatID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if topics[0].Topic != userTopic.Topic {
		t.Fatalf("topic - want: %s, got: %s", userTopic.Topic, topics[0].Topic)
	}
	userTopic = topics[0]

	// delete topic
	n, err := db.DeleteUserTopic(userTopic)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want: delete 1 topic, got: deleted %d", n)
	}

	// ensure topic is deleted
	topics, err = db.FindChatTopics(userTopic.ChatID)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 0 {
		t.Fatalf("want: no topic, got: %d topics", len(topics))
	}
}
