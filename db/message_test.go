package db

import (
	"context"
	"testing"
)

func TestFindMessageThread(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	const chatID = 1

	msgs := []Message{
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
