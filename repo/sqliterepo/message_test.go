package sqliterepo

import (
	"context"
	"testing"
	"time"

	"github.com/igoracmelo/euperturbot/repo"
)

func TestSaveAndFindMessage(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	want := repo.Message{
		ID:               1,
		ChatID:           1,
		Text:             "text",
		Date:             time.Time{},
		UserID:           1,
		UserName:         "name",
		ReplyToMessageID: 2,
	}

	err := db.SaveMessage(context.TODO(), want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.FindMessage(context.TODO(), want.ChatID, want.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("want: %+v, got: %+v", want, got)
	}
}

func TestFindMessageThread(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	const chatID = 1

	msgs := []repo.Message{
		{
			ID:               1,
			ChatID:           chatID,
			Text:             "a",
			ReplyToMessageID: 0,
		},
		{
			ID:               2,
			ChatID:           chatID,
			Text:             "b",
			ReplyToMessageID: 1,
		},
		{
			ID:               3,
			ChatID:           chatID,
			Text:             "c",
			ReplyToMessageID: 2,
		},
		{
			ID:               4,
			ChatID:           chatID,
			Text:             "d",
			ReplyToMessageID: 3,
		},
	}

	for _, msg := range msgs {
		err := db.SaveMessage(context.TODO(), msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	gotMsgs, err := db.FindMessageThread(context.TODO(), chatID, 4) // "d"
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != len(gotMsgs) {
		t.Fatalf("want: len %d, got: len %d", len(msgs), len(gotMsgs))
	}

	for i := 0; i < len(msgs); i++ {
		if msgs[i] != gotMsgs[i] {
			t.Errorf("msg - want: \n%+v, \ngot: \n%+v", msgs[i], gotMsgs[i])
		}
	}
}
