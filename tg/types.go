package tg

import "encoding/json"

type Message struct {
	MessageID         int      `json:"message_id"`
	Date              int64    `json:"date"`
	Text              string   `json:"text,omitempty"`
	ForwardSenderName string   `json:"forward_sender_name,omitempty"`
	From              *User    `json:"from,omitempty"`
	FowardFrom        *User    `json:"forward_from,omitempty"`
	Chat              *Chat    `json:"chat,omitempty"`
	ReplyToMessage    *Message `json:"reply_to_message,omitempty"`
	Poll              *Poll    `json:"poll,omitempty"`
	Voice             *Voice   `json:"voice,omitempty"`
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

func (p SendMessageParams) Error() string {
	return p.Text
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
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
	IsBot     bool   `json:"is_bot,omitempty"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	FirstName string `json:"first_name"`
}

func (c Chat) Name() string {
	if c.Type == "private" {
		return c.FirstName
	}
	return c.Title
}

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	PollAnswer    *PollAnswer    `json:"poll_answer,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
	InlineQuery   *InlineQuery   `json:"inline_query,omitempty"`
}

func (u Update) String() string {
	b, err := json.Marshal(u)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

type PollAnswer struct {
	PollID    string `json:"poll_id"`
	User      User   `json:"user"`
	OptionIDs []int  `json:"option_ids"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type InlineQuery struct {
	ID    string `json:"id"`
	From  *User  `json:"from"`
	Query string `json:"query"`
}

type AnswerInlineQueryParams struct {
	InlineQueryID string              `json:"inline_query_id"`
	Results       []InlineQueryResult `json:"results"`
}

type InlineQueryResult struct {
	Type                string              `json:"type"`
	ID                  string              `json:"id"`
	Title               string              `json:"title"`
	InputMessageContent InputMessageContent `json:"input_message_content"`
}

type InputMessageContent struct {
	MessageText string `json:"message_text"`
	ParseMode   string `json:"parse_mode,omitempty"`
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
