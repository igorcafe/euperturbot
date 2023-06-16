package dao

import (
	"database/sql"
	"log"
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
			chat_id INTEGER,
			user_id INTEGER,
			username TEXT,
			topic TEXT,
			UNIQUE(chat_id, user_id, topic)
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

func (dao *DAO) SaveTopic(topic UserTopic) error {
	_, err := dao.db.Exec(`
		INSERT INTO user_topic
		(chat_id, user_id, username, topic)
		VALUES ($1, $2, $3, $4)
	`, topic.ChatID, topic.UserID, topic.Username, topic.Topic)

	return err
}

func (dao *DAO) DeleteTopic(topic UserTopic) error {
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
		SELECT * FROM user_topic
		WHERE chat_id = $1
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
