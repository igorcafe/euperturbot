package handler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/tg"
)

func (h Handler) callSubs(bot *tg.Bot, u tg.Update, topic string, quiet bool) error {
	users, err := h.DB.FindUsersByTopic(u.Message.Chat.ID, topic)
	if err != nil {
		if quiet {
			return err
		}
		return tg.SendMessageParams{
			Text: "falha ao listar usuários",
		}
	}

	if len(users) == 0 {
		if quiet {
			return nil
		}
		return tg.SendMessageParams{
			Text: "não tem ninguém inscrito nesse tópico",
		}
	}

	txt := fmt.Sprintf(
		"*sim \\(0 votos\\)*\n\n"+
			"*não \\(0 votos\\)*\n\n"+
			"*restam \\(%d votos\\)*\n",
		len(users),
	)

	for _, u := range users {
		txt += fmt.Sprintf("[%s](tg://user?id=%d)\n", u.Name(), u.ID)
	}

	up := "👍 0"
	down := "👎 0"

	msg, err := bot.SendMessage(tg.SendMessageParams{
		ChatID:           u.Message.Chat.ID,
		Text:             txt,
		ParseMode:        "MarkdownV2",
		ReplyToMessageID: u.Message.MessageID,
		ReplyMarkup: &tg.InlineKeyboardMarkup{
			InlineKeyboard: [][]tg.InlineKeyboardButton{{
				tg.InlineKeyboardButton{
					Text:         up,
					CallbackData: "0",
				},
				tg.InlineKeyboardButton{
					Text:         down,
					CallbackData: "1",
				},
			}},
		},
	})
	if err != nil {
		return err
	}

	err = h.DB.SavePoll(db.Poll{
		ID:              strconv.Itoa(msg.MessageID),
		ChatID:          u.Message.Chat.ID,
		Topic:           topic,
		ResultMessageID: msg.MessageID,
	})

	return err
}

func prepareMessagesForGPT(msgs []db.Message) []string {
	msgTxts := []string{}
	totalLen := 0

	reURL := regexp.MustCompile(`https?:\/\/\S+`)
	reMultiSpace := regexp.MustCompile(`\s+`)
	reLaugh := regexp.MustCompile(`([kK]{7})[kK]+`)

	for i := len(msgs) - 1; i >= 0; i-- {
		txt := msgs[i].UserName + ": " + msgs[i].Text
		txt = reLaugh.ReplaceAllString(txt, "$1")
		txt = reURL.ReplaceAllString(txt, "")
		txt = reMultiSpace.ReplaceAllString(txt, " ")
		totalLen += len(txt)
		if totalLen > 2000 {
			break
		}
		msgTxts = append(msgTxts, txt)
	}

	for i, j := 0, len(msgTxts)-1; i < j; i, j = i+1, j-1 {
		msgTxts[i], msgTxts[j] = msgTxts[j], msgTxts[i]
	}

	return msgTxts
}

func sanitizeUsername(name string) string {
	s := ""
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			s += string(r)
		}
	}
	return strings.TrimSpace(s)
}

func username(user *tg.User) string {
	if user == nil {
		return ""
	}
	if user.Username != "" {
		return sanitizeUsername(user.Username)
	}
	return sanitizeUsername(user.FirstName)
}

func validateTopic(topic string) error {
	topic = strings.TrimSpace(topic)
	if len(topic) == 0 {
		return fmt.Errorf("tópico vazio")
	}
	if len(topic) > 30 {
		return fmt.Errorf("tópico muito grande")
	}
	if strings.Contains(topic, "\n") {
		return fmt.Errorf("tópico não pode ter mais de uma linha")
	}
	if strings.Contains(topic, "#") && strings.Contains(topic, " ") {
		return fmt.Errorf("tópico com # não pode ter espaço")
	}
	return nil
}

func (h Handler) isAdmin(bot *tg.Bot, u tg.Update) (bool, error) {
	if u.Message.Chat.Type == "private" {
		return true, nil
	}

	if u.Message.From.ID == h.Config.GodID {
		return true, nil
	}

	member, err := bot.GetChatMember(tg.GetChatMemberParams{
		ChatID: u.Message.Chat.ID,
		UserID: u.Message.From.ID,
	})
	if err != nil {
		return false, err
	}

	return member.Status == "creator" || member.Status == "administrator", nil
}
