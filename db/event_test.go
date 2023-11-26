package db

import (
	"testing"
	"time"
)

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
