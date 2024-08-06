package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repo interface {
	DB() *sqlx.DB
	Close() error
	SaveChat(ctx context.Context, chat Chat) error
	FindChat(ctx context.Context, chatID int64) (*Chat, error)
	ChatEnables(ctx context.Context, chatID int64, action string) (bool, error)
	ChatEnable(ctx context.Context, chatID int64, action string) error
	ChatDisable(ctx context.Context, chatID int64, action string) error
	SaveMessage(ctx context.Context, msg Message) error
	FindMessage(ctx context.Context, chatID int64, msgID int) (Message, error)
	FindMessagesBeforeDate(ctx context.Context, chatID int64, date time.Time, count int) ([]Message, error)
	FindMessageThread(ctx context.Context, chatID int64, msgID int) ([]Message, error)
	SaveUser(u User) error
	FindUser(id int64) (*User, error)
	ExistsChatTopic(chatID int64, topic string) (bool, error)
	SaveUserTopic(topic UserTopic) error
	DeleteUserTopic(topic UserTopic) (int64, error)
	FindUserChatTopics(chatID, userID int64) ([]UserTopic, error)
	FindChatTopics(chatID int64) ([]UserTopic, error)
	FindUsersByTopic(chatID int64, topic string) ([]User, error)
	SavePoll(p Poll) error
	FindPollByMessage(msgID int) (*Poll, error)
	SavePollVote(v PollVote) error
	DeletePollVote(pollID string, userID int64) error
	FindPollVote(pollID string, userID int64) (*PollVote, error)
	SaveVoice(v Voice) error
	FindRandomVoice(chatID int64) (*Voice, error)
}

var (
	ErrChatActionNotAllowed = errors.New("chat action not allowed")
	ErrNotFound             = sql.ErrNoRows // FIXME
)

type Chat struct {
	ID         int64
	Title      string
	EnableCAsk bool
}

type Message struct {
	ID               int
	ChatID           int64 `db:"chat_id"`
	Text             string
	Date             time.Time
	UserID           int64  `db:"user_id"`
	UserName         string `db:"user_name"`
	ReplyToMessageID int    `db:"reply_to_message_id"`
}

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

type Voice struct {
	FileID string `db:"file_id"`
	UserID int64  `db:"user_id"`
	ChatID int64  `db:"chat_id"`
}
