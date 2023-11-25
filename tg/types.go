package tg

type Message struct {
	MessageID         int      `json:"message_id"`
	Date              int64    `json:"date"`
	Text              string   `json:"text"`
	ForwardSenderName string   `json:"forward_sender_name"`
	From              *User    `json:"from"`
	FowardFrom        *User    `json:"forward_from"`
	Chat              *Chat    `json:"chat"`
	ReplyToMessage    *Message `json:"reply_to_message"`
	Poll              *Poll    `json:"poll"`
	Voice             *Voice   `json:"voice"`
}

type Voice struct {
	FileID string `json:"file_id"`
}

type Poll struct {
	ID string `json:"id"`
	// Question string `json:"question"`
}

type SendMessageParams struct {
	ChatID                   int64                 `json:"chat_id"`
	ReplyToMessageID         int                   `json:"reply_to_message_id,omitempty"`
	Text                     string                `json:"text,omitempty"`
	ParseMode                string                `json:"parse_mode,omitempty"`
	AllowSendingWithoutReply bool                  `json:"allow_sending_without_reply,omitempty"`
	ReplyMarkup              *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func (p SendMessageParams) Error() string {
	return p.Text
}

type EditMessageTextParams struct {
	ChatID      int64                 `json:"chat_id"`
	MessageID   int                   `json:"message_id"`
	Text        string                `json:"text,omitempty"`
	ParseMode   string                `json:"parse_mode,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type SendVoiceParams struct {
	ChatID           int64  `json:"chat_id"`
	Voice            string `json:"voice"`
	ReplyToMessageID int    `json:"reply_to_message_id"`
}

type SendPollParams struct {
	ChatID      int64    `json:"chat_id"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	IsAnonymous *bool    `json:"is_anonymous,omitempty"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	PollAnswer    *PollAnswer    `json:"poll_answer"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type PollAnswer struct {
	PollID    string `json:"poll_id"`
	User      User   `json:"user"`
	OptionIDs []int  `json:"option_ids"`
}

type GetUpdatesParams struct {
	Offset         int      `json:"offset,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

type Result[T any] struct {
	Ok     bool `json:"ok"`
	Result T    `json:"result"`
}
