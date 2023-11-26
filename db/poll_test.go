package db

import "testing"

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
