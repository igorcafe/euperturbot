package db

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

func TestUserTopic(t *testing.T) {
	myDB, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer myDB.Close()

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

func TestChatEvent(t *testing.T) {
	myDB, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer myDB.Close()

	err = myDB.SaveChatEvent(ChatEvent{
		ChatID: 1,
		MsgID:  1,
		Time:   time.Time{},
		Name:   "event1",
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := myDB.FindChatEventsByName(1, "event1")
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.Fatalf("want: 1 event, got: %+v", events)
	}

	err = myDB.DeleteChatEvent(ChatEvent{
		ChatID: 1,
		MsgID:  1,
		Name:   "event1",
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err = myDB.FindChatEventsByName(1, "event1")
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 0 {
		t.Fatalf("want: 0 events, got: %+v", events)
	}
}

func TestPollVote(t *testing.T) {
	myDB, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer myDB.Close()

	err = myDB.SaveUser(User{
		ID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = myDB.SavePoll(Poll{
		ID:              "poll",
		ChatID:          1,
		Topic:           "topic",
		ResultMessageID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = myDB.FindPoll("poll")
	if err != nil {
		t.Fatal(err)
	}

	err = myDB.SavePollVote(PollVote{
		PollID: "poll",
		UserID: 1,
		Vote:   1,
	})
	if err != nil {
		t.Fatal(err)
	}

	votes, err := myDB.FindPollVotes("poll")
	if err != nil {
		t.Fatal(err)
	}

	err = myDB.DeletePollVote("poll", 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(votes) == 0 {
		t.Fatal("want: 1 poll vote, got: none")
	}
}

func Test_queryRow(t *testing.T) {
	db, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	var u User
	err = queryRow(db.db, &u, `SELECT * FROM user`)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("want: %v, got: %v", sql.ErrNoRows, err)
	}
}
