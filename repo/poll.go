package repo

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

func (db *Repo) SavePoll(p Poll) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO poll
		(id, chat_id, topic, result_message_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO UPDATE SET result_message_id = $4
	`, p.ID, p.ChatID, p.Topic, p.ResultMessageID)
	return err
}

func (db *Repo) FindPollByMessage(msgID int) (*Poll, error) {
	var p Poll
	err := db.db.GetContext(context.TODO(), &p, `SELECT * FROM poll WHERE result_message_id = $1`, msgID)
	return &p, err
}

func (db *Repo) SavePollVote(v PollVote) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO poll_vote
		(poll_id, user_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT DO UPDATE SET vote = $3
	`, v.PollID, v.UserID, v.Vote)

	return err
}

func (db *Repo) DeletePollVote(pollID string, userID int64) error {
	_, err := db.db.ExecContext(context.TODO(), `
		DELETE FROM poll_vote
		WHERE poll_id = $1 AND user_id = $2
	`, pollID, userID)
	return err
}

func (db *Repo) FindPollVote(pollID string, userID int64) (*PollVote, error) {
	var v PollVote
	err := db.db.GetContext(context.TODO(), &v, `
		SELECT pv.* FROM poll_vote pv
		JOIN user_topic ut ON ut.user_id = $2
		WHERE pv.poll_id = $1 AND pv.user_id = $2
	`, pollID, userID)
	return &v, err
}
