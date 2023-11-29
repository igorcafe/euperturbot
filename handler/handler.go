package handler

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/oai"
	"github.com/igoracmelo/euperturbot/tg"
)

type Handler struct {
	DB      *db.DB
	OAI     *oai.Client
	BotInfo *tg.User
	Config  *config.Config
}

func (h Handler) SubToTopic(bot *tg.Bot, u tg.Update) error {
	fields := strings.SplitN(u.Message.Text, " ", 2)
	topics := []string{}
	if len(fields) > 1 {
		topics = strings.Split(fields[1], "\n")
	}

	if len(topics) == 0 {
		return tg.SendMessageParams{
			Text: "cadê o(s) tópico(s)?",
		}
	}

	if len(topics) > 3 {
		return tg.SendMessageParams{
			Text: "no máximo 3 tópicos por vez",
		}
	}

	user := db.User{
		ID:        u.Message.From.ID,
		FirstName: sanitizeUsername(u.Message.From.FirstName),
		Username:  sanitizeUsername(u.Message.From.Username),
	}

	if u.Message.ReplyToMessage != nil {
		if u.Message.ReplyToMessage.From.IsBot {
			return tg.SendMessageParams{
				Text: "bot nao pode man",
			}
		}
		user.ID = u.Message.ReplyToMessage.From.ID
		user.FirstName = sanitizeUsername(u.Message.ReplyToMessage.From.FirstName)
		user.Username = sanitizeUsername(u.Message.ReplyToMessage.From.Username)
	}

	err := h.DB.SaveUser(user)
	if err != nil {
		return err
	}

	for i, topic := range topics {
		topics[i] = strings.TrimSpace(topic)
		topic := topics[i]

		if err := validateTopic(topic); err != nil {
			return err
		}

		exists, err := h.DB.ExistsChatTopic(u.Message.Chat.ID, topic)
		if err != nil {
			return err
		}

		if !exists && u.Message.From.ID != h.Config.GodID {
			return tg.SendMessageParams{
				Text: "macaquearam demais... chega!",
			}
		}

		userTopic := db.UserTopic{
			ChatID: u.Message.Chat.ID,
			UserID: user.ID,
			Topic:  topic,
		}
		err = h.DB.SaveUserTopic(userTopic)
		if err != nil {
			log.Print(err)
			return tg.SendMessageParams{
				Text: "falha ao salvar tópico " + topic,
			}
		}
	}

	txt := fmt.Sprintf("inscrições adicionadas para %s:\n", user.Name())
	for _, topic := range topics {
		txt += fmt.Sprintf("- %s\n", topic)
	}
	return tg.SendMessageParams{
		Text: txt,
	}
}

func (h Handler) UnsubTopic(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		return err
	}

	n, err := h.DB.DeleteUserTopic(db.UserTopic{
		ChatID: u.Message.Chat.ID,
		UserID: u.Message.From.ID,
		Topic:  topic,
	})
	if err != nil {
		return fmt.Errorf("falha ao descer :/ (%w)", err)
	}

	user, err := h.DB.FindUser(u.Message.From.ID)
	if err != nil {
		return err
	}

	if n == 0 {
		return tg.SendMessageParams{
			Text: fmt.Sprintf("usuário %s não está inscrito nesse tópico", user.Name()),
		}
	}

	return tg.SendMessageParams{
		Text: "inscrição removida para " + user.Name(),
	}
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
