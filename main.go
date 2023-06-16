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

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		panic(err)
	}

	h, err := th.NewBotHandler(bot, updates)
	if err != nil {
		panic(err)
	}

	h.Handle(handleSubTopic, th.CommandEqual("suba"))
	h.Handle(handleUnsubTopic, th.CommandEqual("desca"))
	h.Handle(handleCallSubs, th.CommandEqual("bora"))
	h.Handle(handleListUserTopics, th.CommandEqual("lista"))
	h.Handle(handleListChatTopics, th.CommandEqual("listudo"))

	defer h.Stop()
	h.Start()
}

func handleSubTopic(bot *tg.Bot, u tg.Update) {
	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
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

	err := mydao.SaveTopic(userTopic)
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
	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	err := mydao.DeleteTopic(dao.UserTopic{
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

func handleCallSubs(bot *tg.Bot, u tg.Update) {
	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if !isTopicValid(topic) {
		_, _ = replyToMessage(bot, u.Message, &tg.SendMessageParams{
			Text: "tópico inválido",
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
		txt += topic.Topic + "\n"
	}

	_, err = replyToMessage(bot, u.Message, &tg.SendMessageParams{
		Text: txt,
	})
	if err != nil {
		log.Print(err)
	}
}

func handleListChatTopics(bot *tg.Bot, u tg.Update) {
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
		txt += topic.Topic + "\n"
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

func isTopicValid(topic string) bool {
	if len(strings.TrimSpace(topic)) == 0 {
		return false
	}
	return true
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
