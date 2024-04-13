package repo

import (
	"time"
)

type ChatEvent struct {
	ID     int64
	ChatID int64 `db:"chat_id"`
	MsgID  int   `db:"msg_id"`
	Time   time.Time
	Name   string
}
