package main

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/glebarez/go-sqlite"
	"github.com/igoracmelo/euperturbot/dao"
	"github.com/igoracmelo/euperturbot/env"
	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	sqlite3 "modernc.org/sqlite/lib"
)

var token = env.Must("TOKEN")
var godID = env.MustInt64("GOD_ID")
var mydao *dao.DAO

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	mydao, err = dao.NewSqlite("euperturbot.db")
	if err != nil {
		panic(err)
	}

	bot, err := tg.NewBot(token)
	if err != nil {
		panic(err)
	}

	_, err = bot.GetMe()
	if err != nil {
		panic(err)
	}

	updates, err := bot.UpdatesViaLongPolling(&tg.GetUpdatesParams{
		Timeout:        10,
		AllowedUpdates: []string{"message"},
	})
	if err != nil {
		panic(err)
	}

	h, err := th.NewBotHandler(bot, updates)
	if err != nil {
		panic(err)
	}

	h.Handle(handleSubTopic, th.CommandEqual("suba"))
	h.Handle(handleUnsubTopic, th.CommandEqual("desca"))
	h.Handle(handleCreatePoll, th.CommandEqual("pollo"))
	h.Handle(handleCallSubs, th.CommandEqual("bora"))
	h.Handle(handleListSubs, th.CommandEqual("quem"))
	h.Handle(handleListUserTopics, th.CommandEqual("lista"))
	h.Handle(handleListChatTopics, th.CommandEqual("listudo"))

	defer h.Stop()
	h.Start()
}

func handleSubTopic(bot *tg.Bot, u tg.Update) {
	log.Print(u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: err.Error(),
		})
		return
	}

	exists, err := mydao.ExistsChatTopic(u.Message.Chat.ID, topic)
	if err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: err.Error(),
		})
		return
	}

	if !exists && u.Message.From.ID != godID {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "macaquearam demais... chega!",
		})
		return
	}

	userTopic := dao.UserTopic{
		ChatID:   u.Message.Chat.ID,
		UserID:   u.Message.From.ID,
		Username: username(u.Message.From),
		Topic:    topic,
	}
	if u.Message.ReplyToMessage != nil {
		userTopic.UserID = u.Message.ReplyToMessage.From.ID
		userTopic.Username = username(u.Message.ReplyToMessage.From)
	}

	err = mydao.SaveUserTopic(userTopic)
	if err, ok := err.(*sqlite.Error); ok &&
		err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "já inscrito nesse tópico",
		})
		return
	}
	if err != nil {
		fmt.Println(err)
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "falha ao salvar tópico",
		})
		return
	}

	_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
		Text: "inscrição adicionada para " + userTopic.Username,
	})
	if err != nil {
		log.Print(err)
	}
}

func handleUnsubTopic(bot *tg.Bot, u tg.Update) {
	log.Print(u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: err.Error(),
		})
		return
	}

	err := mydao.DeleteUserTopic(dao.UserTopic{
		ChatID:   u.Message.Chat.ID,
		UserID:   u.Message.From.ID,
		Username: username(u.Message.From),
		Topic:    topic,
	})
	if err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "falha ao descer :/",
		})
		return
	}

	_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
		Text: "inscrição removida para o tópico " + topic,
	})
	if err != nil {
		log.Print(err)
	}
}

func handleCreatePoll(bot *tg.Bot, u tg.Update) {
	log.Print(username(u.Message.From) + ": " + u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	question := ""
	if len(fields) > 1 {
		question = strings.TrimSpace(fields[1])
	}

	if question == "" {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "cade o titulo joe",
		})
		return
	}

	_, err := bot.SendPoll(&tg.SendPollParams{
		ChatID: tg.ChatID{
			ID: u.Message.Chat.ID,
		},
		Question:    question,
		Options:     []string{"👍🏿", "👎🏻"},
		IsAnonymous: tg.ToPtr(false),
	})
	if err != nil {
		log.Print(err)
		return
	}
}

func handleListSubs(bot *tg.Bot, u tg.Update) {
	log.Print(u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: err.Error(),
		})
		return
	}

	topics, err := mydao.FindSubscriptionsByTopic(u.Message.Chat.ID, topic)
	if err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "falha ao listar usuários",
		})
		return
	}

	if len(topics) == 0 {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "não tem ninguém inscrito nesse tópico",
		})
		return
	}

	txt := "inscritos nesse tópico:\n"
	for _, t := range topics {
		txt += fmt.Sprintf("- %s\n", t.Username)
	}
	_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
		Text: txt,
	})
	if err != nil {
		log.Print(err)
		return
	}
}

func handleCallSubs(bot *tg.Bot, u tg.Update) {
	log.Print(username(u.Message.From) + ": " + u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: err.Error(),
		})
		return
	}

	topics, err := mydao.FindSubscriptionsByTopic(u.Message.Chat.ID, topic)
	if err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "falha ao listar usuários",
		})
		return
	}

	if len(topics) == 0 {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "não tem ninguém inscrito nesse tópico",
		})
		return
	}

	_, err = bot.SendPoll(&tg.SendPollParams{
		ChatID: tg.ChatID{
			ID: u.Message.Chat.ID,
		},
		Question:    topic,
		Options:     []string{"bo 👍🏿", "bo nao 👎🏻"},
		IsAnonymous: tg.ToPtr(false),
	})
	if err != nil {
		log.Print(err)
		return
	}

	txt := ""
	for i, t := range topics {
		txt += fmt.Sprintf("[%s](tg://user?id=%d)\n", t.Username, t.UserID)
		if (i+1)%4 == 0 {
			_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
				Text:      txt,
				ParseMode: "MarkdownV2",
			})
			if err != nil {
				log.Print(err)
				return
			}
			txt = ""
		}
	}

	if txt != "" {
		_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text:      txt,
			ParseMode: "MarkdownV2",
		})
		if err != nil {
			log.Print(err)
			return
		}
	}
}

func handleListUserTopics(bot *tg.Bot, u tg.Update) {
	log.Print(u.Message.Text)

	topics, err := mydao.FindUserChatTopics(u.Message.Chat.ID, u.Message.From.ID)
	if err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "falha ao listar tópicos",
		})
		return
	}

	if len(topics) == 0 {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "você não está inscrito em nenhum tópico",
		})
		return
	}

	txt := "seus tópicos:\n"
	for _, topic := range topics {
		txt += "- " + topic.Topic + "\n"
	}

	_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
		Text: txt,
	})
	if err != nil {
		log.Print(err)
	}
}

func handleListChatTopics(bot *tg.Bot, u tg.Update) {
	log.Print(u.Message.Text)

	topics, err := mydao.FindChatTopics(u.Message.Chat.ID)
	if err != nil {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "falha ao listar tópicos",
		})
		return
	}

	if len(topics) == 0 {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "não existe nenhum tópico registrado nesse chat",
		})
		return
	}

	txt := "tópicos:\n"
	for _, topic := range topics {
		txt += "- " + topic.Topic + "\n"
	}

	_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
		Text: txt,
	})
	if err != nil {
		log.Print(err)
	}
}

func replyToMessage(bot *tg.Bot, msg *tg.Message, params *tg.SendMessageParams) (*tg.Message, error) {
	if params == nil {
		params = &tg.SendMessageParams{}
	}

	params.ChatID = tg.ChatID{
		ID: msg.Chat.ID,
	}
	params.ReplyToMessageID = msg.MessageID

	return bot.SendMessage(params)
}

func validateTopic(topic string) error {
	topic = strings.TrimSpace(topic)
	if len(topic) == 0 {
		return fmt.Errorf("tópico vazio")
	}
	if len(topic) > 30 {
		return fmt.Errorf("tópico muito grande")
	}
	return nil
}

func sanitizeUsername(topic string) string {
	s := ""
	for _, r := range topic {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			s += string(r)
		}
	}
	return strings.TrimSpace(s)
}

func username(user *tg.User) string {
	s := ""
	if user.Username != "" {
		s = sanitizeUsername(user.Username)
	} else {
		s = sanitizeUsername(user.FirstName)
	}
	return s
}
