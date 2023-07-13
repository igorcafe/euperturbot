package dao_test

import (
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/igoracmelo/euperturbot/dao"
)

func TestIntegration(t *testing.T) {
	mydao, err := dao.NewSqlite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// me: sub game
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
