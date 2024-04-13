package repo

import "context"

type User struct {
	ID        int64
	FirstName string `db:"first_name"`
	Username  string
}

func (u *User) Name() string {
	if u.Username != "" {
		return u.Username
	}
	return u.FirstName
}

type UserTopic struct {
	ID          int64
	ChatID      int64 `db:"chat_id"`
	UserID      int64 `db:"user_id"`
	Topic       string
	Subscribers int
}

func (db *Repo) SaveUser(u User) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO user
		(id, first_name, username)
		VALUES ($1, $2, $3)
		ON CONFLICT DO UPDATE
		SET first_name = $2, username = $3
	`, u.ID, u.FirstName, u.Username)

	return err
}

func (db *Repo) FindUser(id int64) (*User, error) {
	var u User
	err := db.db.GetContext(context.TODO(), &u, `SELECT * FROM user WHERE id = $1`, id)
	return &u, err
}

func (db *Repo) ExistsChatTopic(chatID int64, topic string) (bool, error) {
	row := db.db.QueryRowContext(context.TODO(), `
		SELECT EXISTS (
			SELECT * FROM user_topic
			WHERE chat_id = $1 AND topic = $2
		)
	`, chatID, topic)

	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

func (db *Repo) SaveUserTopic(topic UserTopic) error {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO user_topic
		(chat_id, user_id, topic)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, topic.ChatID, topic.UserID, topic.Topic)

	return err
}

func (db *Repo) DeleteUserTopic(topic UserTopic) (int64, error) {
	res, err := db.db.ExecContext(context.TODO(), `
		DELETE FROM user_topic
		WHERE chat_id = $1 AND user_id = $2 AND topic = $3
	`, topic.ChatID, topic.UserID, topic.Topic)

	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (db *Repo) FindUserChatTopics(chatID, userID int64) ([]UserTopic, error) {
	sql := `
		SELECT *, (
			SELECT COUNT(*) FROM user_topic
			WHERE chat_id = $1 AND topic = ut.topic
			GROUP BY topic
		) AS subscribers
		FROM user_topic ut
		WHERE chat_id = $1 AND user_id = $2
	`
	var topics []UserTopic
	err := db.db.SelectContext(context.TODO(), &topics, sql, chatID, userID)
	return topics, err
}

func (db *Repo) FindChatTopics(chatID int64) ([]UserTopic, error) {
	sql := `
		SELECT DISTINCT *, COUNT(*) AS subscribers FROM user_topic
		WHERE chat_id = $1
		GROUP BY topic, chat_id
		ORDER BY subscribers DESC
	`
	var topics []UserTopic
	err := db.db.SelectContext(context.TODO(), &topics, sql, chatID)
	return topics, err
}

func (db *Repo) FindUsersByTopic(chatID int64, topic string) ([]User, error) {
	sql := `
		SELECT u.* FROM user u
		JOIN user_topic ut ON u.id = ut.user_id
		WHERE ut.chat_id = $1 AND ut.topic = $2
	`
	var users []User
	err := db.db.SelectContext(context.TODO(), &users, sql, chatID, topic)
	return users, err
}
