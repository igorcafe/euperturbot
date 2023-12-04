package db

import "context"

type Poll struct {
	ID              string
	ChatID          int64 `db:"chat_id"`
	Topic           string
	ResultMessageID int `db:"result_message_id"`
}

type PollVote struct {
	PollID string `db:"poll_id"`
	UserID int64  `db:"user_id"`
	Vote   int
}

const (
	VoteUp   = 0
	VoteDown = 1
)

func (db *DB) SavePoll(p Poll) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO poll
		(id, chat_id, topic, result_message_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO UPDATE SET result_message_id = $4
	`, p.ID, p.ChatID, p.Topic, p.ResultMessageID)
	return err
}

func (db *DB) FindPoll(pollID string) (*Poll, error) {
	var p Poll
	err := db.db.GetContext(context.TODO(), &p, `SELECT * FROM poll WHERE id = $1`, pollID)
	return &p, err
}

func (db *DB) FindPollByMessage(msgID int) (*Poll, error) {
	var p Poll
	err := db.db.GetContext(context.TODO(), &p, `SELECT * FROM poll WHERE result_message_id = $1`, msgID)
	return &p, err
}

func (db *DB) FindLastPollByTopic(topic string) (*Poll, error) {
	var p Poll
	err := db.db.GetContext(context.TODO(), &p, `
		SELECT * FROM poll
		WHERE topic = $1
		ORDER BY result_message_id DESC
	`, topic)
	return &p, err
}

func (db *DB) SavePollVote(v PollVote) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO poll_vote
		(poll_id, user_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT DO UPDATE SET vote = $3
	`, v.PollID, v.UserID, v.Vote)

	return err
}

func (db *DB) DeletePollVote(pollID string, userID int64) error {
	_, err := db.db.ExecContext(context.TODO(), `
		DELETE FROM poll_vote
		WHERE poll_id = $1 AND user_id = $2
	`, pollID, userID)
	return err
}

func (db *DB) FindPollVotes(pollID string) ([]PollVote, error) {
	sql := `
		SELECT * FROM poll_vote
		WHERE poll_id = $1
	`
	var votes []PollVote
	err := db.db.SelectContext(context.TODO(), &votes, sql, pollID)
	return votes, err
}

func (db *DB) FindPollVote(pollID string, userID int64) (*PollVote, error) {
	var v PollVote
	err := db.db.GetContext(context.TODO(), &v, `
		SELECT * FROM poll_vote
		WHERE poll_id = $1 AND user_id = $2
	`, pollID, userID)
	return &v, err
}
