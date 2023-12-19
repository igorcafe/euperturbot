package db

import (
	"context"
	"testing"
)

func TestUserTopic(t *testing.T) {
	myDB, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer myDB.Close()

	err = myDB.Migrate(context.Background(), "./migrations")
	if err != nil {
		t.Fatal(err)
	}

	err = myDB.SaveUser(User{
		ID:       1,
		Username: "me",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = myDB.SaveUser(User{
		ID:       2,
		Username: "you",
	})
	if err != nil {
		t.Fatal(err)
	}

	// me: /sub game
	err = myDB.SaveUserTopic(UserTopic{
		ChatID: 1,
		UserID: 1,
		Topic:  "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	// you: sub game
	err = myDB.SaveUserTopic(UserTopic{
		ChatID: 1,
		UserID: 2,
		Topic:  "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	// you: sub other
	err = myDB.SaveUserTopic(UserTopic{
		ChatID: 1,
		UserID: 2,
		Topic:  "other",
	})
	if err != nil {
		t.Fatal(err)
	}

	// exists topic game
	exists, err := myDB.ExistsChatTopic(1, "game")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("topic 'game' should be found")
	}

	// me: /list
	topics, err := myDB.FindUserChatTopics(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 1 || topics[0].Topic != "game" {
		t.Fatal("topic 'game' not found for user 'me'")
	}

	// /who game
	users, err := myDB.FindUsersByTopic(1, "game")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatal("topic 'game' not found for user 'me'")
	}

	u1 := users[0]
	u2 := users[1]
	if u1.Username != "me" || u2.Username != "you" {
		t.Fatalf("unexpected subscriptions for topic 'game': %s, %s", u1.Username, u2.Username)
	}

	up, err := myDB.FindUser(u1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if up.Username != "me" {
		t.Fatalf("want: %+v, got: %+v", u1, *up)
	}

	topics, err = myDB.FindChatTopics(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 2 {
		t.Fatalf("want: 2 topics, got: %d", len(topics))
	}
	if topics[0].Topic != "game" || topics[1].Topic != "other" {
		t.Fatalf("want: game and other, got: %s and %s", topics[0].Topic, topics[1].Topic)
	}

	// me: /unsub game
	_, err = myDB.DeleteUserTopic(UserTopic{
		ChatID: 1,
		UserID: 1,
		Topic:  "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	// you: unsub game
	_, err = myDB.DeleteUserTopic(UserTopic{
		ChatID: 1,
		UserID: 2,
		Topic:  "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	exists, err = myDB.ExistsChatTopic(1, "game")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("topic 'game' should not be found")
	}
}
