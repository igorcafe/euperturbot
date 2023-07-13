package dao_test

import (
	"testing"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/igoracmelo/euperturbot/dao"
)

func TestUserTopic(t *testing.T) {
	mydao, err := dao.NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer mydao.Close()

	// me: /sub game
	err = mydao.SaveUserTopic(dao.UserTopic{
		ChatID:   1,
		UserID:   1,
		Username: "me",
		Topic:    "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	// you: sub game
	err = mydao.SaveUserTopic(dao.UserTopic{
		ChatID:   1,
		UserID:   2,
		Username: "you",
		Topic:    "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	// you: sub other
	err = mydao.SaveUserTopic(dao.UserTopic{
		ChatID:   1,
		UserID:   2,
		Username: "you",
		Topic:    "other",
	})
	if err != nil {
		t.Fatal(err)
	}

	// exists topic game
	exists, err := mydao.ExistsChatTopic(1, "game")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("topic 'game' should be found")
	}

	// me: /list
	topics, err := mydao.FindUserChatTopics(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 1 || topics[0].Topic != "game" {
		t.Fatal("topic 'game' not found for user 'me'")
	}

	// /who game
	topics, err = mydao.FindSubscriptionsByTopic(1, "game")
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 2 {
		t.Fatal("topic 'game' not found for user 'me'")
	}
	if topics[0].Username != "me" || topics[1].Username != "you" {
		t.Fatalf("unexpected subscriptions for topic 'game': %s, %s", topics[0].Username, topics[1].Username)
	}

	topics, err = mydao.FindChatTopics(1)
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
	err = mydao.DeleteUserTopic(dao.UserTopic{
		ChatID: 1,
		UserID: 1,
		Topic:  "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	// you: unsub game
	err = mydao.DeleteUserTopic(dao.UserTopic{
		ChatID: 1,
		UserID: 2,
		Topic:  "game",
	})
	if err != nil {
		t.Fatal(err)
	}

	exists, err = mydao.ExistsChatTopic(1, "game")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("topic 'game' should not be found")
	}
}

func TestChatEvent(t *testing.T) {
	mydao, err := dao.NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer mydao.Close()

	err = mydao.SaveChatEvent(dao.ChatEvent{
		ChatID: 1,
		MsgID:  1,
		Time:   time.Time{},
		Name:   "event1",
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := mydao.FindChatEventsByName(1, "event1")
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.Fatalf("want: 1 event, got: %+v", events)
	}

	err = mydao.DeleteChatEvent(dao.ChatEvent{
		ChatID: 1,
		MsgID:  1,
		Name:   "event1",
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err = mydao.FindChatEventsByName(1, "event1")
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 0 {
		t.Fatalf("want: 0 events, got: %+v", events)
	}
}

func TestPollVote(t *testing.T) {
	mydao, err := dao.NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer mydao.Close()

	err = mydao.SavePoll(dao.Poll{
		ID:              "poll",
		ChatID:          1,
		Topic:           "topic",
		ResultMessageID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = mydao.SavePollVote(dao.PollVote{
		PollID: "poll",
		UserID: 1,
		Vote:   1,
	})
	if err != nil {
		t.Fatal(err)
	}

	votes, err := mydao.FindPollVotes("poll")
	if err != nil {
		t.Fatal(err)
	}

	if len(votes) == 0 {
		t.Fatal("want: 1 poll vote, got: none")
	}
}
