package dao

import (
	"database/sql"
	"log"
	"time"
)

type DAO struct {
	db *sql.DB
}

func NewSqlite(dsn string) (*DAO, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_topic (
			id INTEGER PRIMARY KEY,
			chat_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			username TEXT NOT NULL,
			topic TEXT NOT NULL,
			UNIQUE(chat_id, user_id, topic)
		)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS event (
			id INTEGER PRIMARY KEY,
			chat_id INTEGER NOT NULL,
			time TIMESTAMP NOT NULL,
			name TEXT NOT NULL,
			msg_id INTEGER NOT NULL,
			UNIQUE(chat_id, msg_id, name)
		)
	`)
	if err != nil {
		return nil, err
	}

	return &DAO{db}, nil
}

type UserTopic struct {
	ID       int64
	ChatID   int64
	UserID   int64
	Username string
	Topic    string
}

func (dao *DAO) ExistsChatTopic(chatID int64, topic string) (bool, error) {
	row := dao.db.QueryRow(`
		SELECT EXISTS (
			SELECT * FROM user_topic
			WHERE chat_id = $1 AND topic = $2
		)
	`, chatID, topic)

	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

func (dao *DAO) SaveUserTopic(topic UserTopic) error {
	_, err := dao.db.Exec(`
		INSERT INTO user_topic
		(chat_id, user_id, username, topic)
		VALUES ($1, $2, $3, $4)
	`, topic.ChatID, topic.UserID, topic.Username, topic.Topic)

	return err
}

func (dao *DAO) DeleteUserTopic(topic UserTopic) error {
	_, err := dao.db.Exec(`
		DELETE FROM user_topic
		WHERE chat_id = $1 AND user_id = $2 AND topic = $3
	`, topic.ChatID, topic.UserID, topic.Topic)

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (dao *DAO) FindUserChatTopics(chatID, userID int64) ([]UserTopic, error) {
	rows, err := dao.db.Query(`
		SELECT * FROM user_topic
		WHERE chat_id = $1 AND user_id = $2
	`, chatID, userID)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	topics := []UserTopic{}

	for rows.Next() {
		t := UserTopic{}
		err := rows.Scan(&t.ID, &t.ChatID, &t.UserID, &t.Username, &t.Topic)
		if err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}

	return topics, nil
}

func (dao *DAO) FindChatTopics(chatID int64) ([]UserTopic, error) {
	rows, err := dao.db.Query(`
		SELECT DISTINCT * FROM user_topic
		WHERE chat_id = $1
		GROUP BY topic
	`, chatID)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	topics := []UserTopic{}

	for rows.Next() {
		t := UserTopic{}
		err := rows.Scan(&t.ID, &t.ChatID, &t.UserID, &t.Username, &t.Topic)
		if err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}

	return topics, nil
}

func (dao *DAO) FindSubscriptionsByTopic(chatID int64, topic string) ([]UserTopic, error) {
	rows, err := dao.db.Query(`
		SELECT * FROM user_topic
		WHERE chat_id = $1 AND topic = $2
	`, chatID, topic)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	topics := []UserTopic{}

	for rows.Next() {
		t := UserTopic{}
		err := rows.Scan(&t.ID, &t.ChatID, &t.UserID, &t.Username, &t.Topic)
		if err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}

	return topics, nil
}

type ChatEvent struct {
	ID     int64
	ChatID int64
	MsgID  int
	Time   time.Time
	Name   string
}

func (dao *DAO) SaveChatEvent(e ChatEvent) error {
	_, err := dao.db.Exec(`
		INSERT INTO event
		(chat_id, msg_id, time, name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO UPDATE SET time = $3
	`, e.ChatID, e.MsgID, e.Time, e.Name)

	return err
}

func (dao *DAO) FindChatEventsByName(e ChatEvent) ([]ChatEvent, error) {
	rows, err := dao.db.Query(`
		SELECT id, chat_id, msg_id, time, name FROM event
		WHERE chat_id = $1 AND name = $2
		ORDER BY time DESC
	`, e.ChatID, e.Name)
	if err != nil {
		return nil, err
	}

	events := []ChatEvent{}
	for rows.Next() {
		e := ChatEvent{}
		err := rows.Scan(&e.ID, &e.ChatID, &e.MsgID, &e.Time, &e.Name)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

func (dao *DAO) DeleteChatEvent(e ChatEvent) (int64, error) {
	res, err := dao.db.Exec(`
		DELETE FROM event
		WHERE chat_id = $1 AND msg_id = $2 AND name = $3
	`, e.ChatID, e.MsgID, e.Name)
	if err != nil {
		return 0, err
	}

	n, err := res.RowsAffected()
	return n, err
}
