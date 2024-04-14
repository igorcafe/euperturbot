package sqliterepo

import (
	"context"

	"github.com/igoracmelo/euperturbot/repo"
)

func (db *sqliteRepo) SavePoll(p repo.Poll) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO poll
		(id, chat_id, topic, result_message_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO UPDATE SET result_message_id = $4
	`, p.ID, p.ChatID, p.Topic, p.ResultMessageID)
	return err
}

func (db *sqliteRepo) FindPollByMessage(msgID int) (*repo.Poll, error) {
	var p repo.Poll
	err := db.db.GetContext(context.TODO(), &p, `SELECT * FROM poll WHERE result_message_id = $1`, msgID)
	return &p, err
}

func (db *sqliteRepo) SavePollVote(v repo.PollVote) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO poll_vote
		(poll_id, user_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT DO UPDATE SET vote = $3
	`, v.PollID, v.UserID, v.Vote)

	return err
}

func (db *sqliteRepo) DeletePollVote(pollID string, userID int64) error {
	_, err := db.db.ExecContext(context.TODO(), `
		DELETE FROM poll_vote
		WHERE poll_id = $1 AND user_id = $2
	`, pollID, userID)
	return err
}

func (db *sqliteRepo) FindPollVote(pollID string, userID int64) (*repo.PollVote, error) {
	var v repo.PollVote
	err := db.db.GetContext(context.TODO(), &v, `
		SELECT pv.* FROM poll_vote pv
		JOIN user_topic ut ON ut.user_id = $2
		WHERE pv.poll_id = $1 AND pv.user_id = $2
	`, pollID, userID)
	return &v, err
}
