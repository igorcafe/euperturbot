package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	_ "github.com/glebarez/go-sqlite"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/env"
	"github.com/igoracmelo/euperturbot/tg"
	"github.com/igoracmelo/euperturbot/util"
)

var token string
var godID int64
var myDB *db.DB

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	token = env.Must("TOKEN")
	godID = env.MustInt64("GOD_ID")

	var err error
	myDB, err = db.NewSqlite("euperturbot.db")
	if err != nil {
		panic(err)
	}

	bot := tg.NewBot(token)
	if err != nil {
		panic(err)
	}

	_, err = bot.GetMe()
	if err != nil {
		panic(err)
	}

	updates := bot.GetUpdatesChannel()
	h := tg.NewUpdateHandler(bot, updates)

	h.HandleCommand("suba", handleSubTopic)
	h.HandleCommand("desca", handleUnsubTopic)
	h.HandleCommand("pollo", handleCreatePoll)
	h.HandleCommand("bora", handleCallSubs)
	h.HandleCommand("quem", handleListSubs)
	h.HandleCommand("lista", handleListUserTopics)
	h.HandleCommand("listudo", handleListChatTopics)
	h.HandleCommand("conta", handleCountEvent)
	h.HandleCommand("desconta", handleUncountEvent)
	h.HandlePollAnswer(handlePollAnswer)
	h.Start()
}

func handleSubTopic(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

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

	err := myDB.SaveUser(user)
	if err != nil {
		return err
	}

	for i, topic := range topics {
		topics[i] = strings.TrimSpace(topic)
		topic := topics[i]

		if err := validateTopic(topic); err != nil {
			return err
		}

		exists, err := myDB.ExistsChatTopic(u.Message.Chat.ID, topic)
		if err != nil {
			return err
		}

		if !exists && u.Message.From.ID != godID {
			return tg.SendMessageParams{
				Text: "macaquearam demais... chega!",
			}
		}

		userTopic := db.UserTopic{
			ChatID: u.Message.Chat.ID,
			UserID: user.ID,
			Topic:  topic,
		}
		err = myDB.SaveUserTopic(userTopic)
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

func handleUnsubTopic(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		return err
	}

	n, err := myDB.DeleteUserTopic(db.UserTopic{
		ChatID: u.Message.Chat.ID,
		UserID: u.Message.From.ID,
		Topic:  topic,
	})
	if err != nil {
		return fmt.Errorf("falha ao descer :/ (%w)", err)
	}

	user, err := myDB.FindUser(u.Message.From.ID)
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

func handleCreatePoll(bot *tg.Bot, u tg.Update) error {
	log.Print(username(u.Message.From) + ": " + u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	question := ""
	if len(fields) > 1 {
		question = strings.TrimSpace(fields[1])
	}

	if question == "" {
		return fmt.Errorf("cade o titulo joe")
	}

	_, err := bot.SendPoll(tg.SendPollParams{
		ChatID:      u.Message.Chat.ID,
		Question:    question,
		Options:     []string{"👍🏿", "👎🏻"},
		IsAnonymous: util.ToPtr(false),
	})
	return err
}

func handleListSubs(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		return err
	}

	users, err := myDB.FindUsersByTopic(u.Message.Chat.ID, topic)
	if err != nil {
		return tg.SendMessageParams{
			Text: "falha ao listar usuários",
		}
	}

	if len(users) == 0 {
		return tg.SendMessageParams{
			Text: "não tem ninguém inscrito nesse tópico",
		}
	}

	txt := fmt.Sprintf("*inscritos \\(%d\\)*\n", len(users))
	for _, user := range users {
		txt += fmt.Sprintf("\\- %s\n", user.Name())
	}
	return tg.SendMessageParams{
		Text:      txt,
		ParseMode: "MarkdownV2",
	}
}

func handleCallSubs(bot *tg.Bot, u tg.Update) error {
	log.Print(username(u.Message.From) + ": " + u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		return tg.SendMessageParams{
			Text: err.Error(),
		}
	}

	users, err := myDB.FindUsersByTopic(u.Message.Chat.ID, topic)
	if err != nil {
		return tg.SendMessageParams{
			Text: "falha ao listar usuários",
		}
	}

	if len(users) == 0 {
		return tg.SendMessageParams{
			Text: "não tem ninguém inscrito nesse tópico",
		}
	}

	msg, err := bot.SendPoll(tg.SendPollParams{
		ChatID:      u.Message.Chat.ID,
		Question:    topic,
		Options:     []string{"bo 👍🏿", "bo nao 👎🏻"},
		IsAnonymous: util.ToPtr(false),
	})
	if err != nil {
		return err
	}
	poll := msg.Poll

	txt := fmt.Sprintf("*sim \\(0 votos\\)*\n\n"+
		"*não \\(0 votos\\)*\n\n"+
		"*restam \\(%d votos\\)*\n", len(users))

	for _, u := range users {
		txt += fmt.Sprintf("[%s](tg://user?id=%d)\n", u.Name(), u.ID)
	}

	msg, err = bot.SendMessage(tg.SendMessageParams{
		ChatID:    u.Message.Chat.ID,
		Text:      txt,
		ParseMode: "MarkdownV2",
	})
	if err != nil {
		return err
	}

	err = myDB.SavePoll(db.Poll{
		ID:              poll.ID,
		ChatID:          u.Message.Chat.ID,
		Topic:           topic,
		ResultMessageID: msg.MessageID,
	})
	if err != nil {
		return err
	}

	return err
}

func handleListUserTopics(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	topics, err := myDB.FindUserChatTopics(u.Message.Chat.ID, u.Message.From.ID)
	if err != nil {
		return tg.SendMessageParams{
			Text: "falha ao listar tópicos",
		}
	}

	if len(topics) == 0 {
		return tg.SendMessageParams{
			Text: "você não está inscrito em nenhum tópico",
		}
	}

	txt := "seus tópicos:\n"
	for _, topic := range topics {
		txt += fmt.Sprintf("(%02d)  %s\n", topic.Subscribers, topic.Topic)
	}

	return tg.SendMessageParams{
		Text: txt,
	}
}

func handleListChatTopics(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	topics, err := myDB.FindChatTopics(u.Message.Chat.ID)
	if err != nil {
		log.Print(err)
		return tg.SendMessageParams{
			Text: "falha ao listar tópicos",
		}
	}

	if len(topics) == 0 {
		return tg.SendMessageParams{
			Text: "não existe nenhum tópico registrado nesse chat",
		}
	}

	txt := "tópicos:\n"
	for _, topic := range topics {
		txt += fmt.Sprintf("- (%02d)  %s\n", topic.Subscribers, topic.Topic)
	}

	return tg.SendMessageParams{
		Text: txt,
	}
}

func handleCountEvent(bot *tg.Bot, u tg.Update) error {
	fields := strings.SplitN(u.Message.Text, " ", 2)
	if len(fields) == 1 {
		return tg.SendMessageParams{
			Text: "faltando nome do evento",
		}
	}

	event := db.ChatEvent{
		ChatID: u.Message.Chat.ID,
		Name:   strings.TrimSpace(fields[1]),
	}

	if u.Message.ReplyToMessage != nil {
		event.MsgID = u.Message.ReplyToMessage.MessageID
		event.Time = time.Unix(u.Message.ReplyToMessage.Date, 0)
		if u.Message.From.ID != godID {
			return tg.SendMessageParams{
				Text: "sai macaco",
			}
		}

		err := myDB.SaveChatEvent(event)
		return err
	}

	events, err := myDB.FindChatEventsByName(event.ChatID, event.Name)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return tg.SendMessageParams{
			Text: fmt.Sprintf("%s 0 vez(es)", event.Name),
		}
	}

	last := time.Now().Sub(events[0].Time)
	relative := util.RelativeDuration(last)

	var txt string
	if len(events) == 1 {
		txt = fmt.Sprintf("%s %d vez há %s", event.Name, len(events), relative)
	} else {
		txt = fmt.Sprintf("%s %d vezes. última vez há %s", event.Name, len(events), relative)
	}

	return tg.SendMessageParams{
		Text: txt,
	}
}

func handleUncountEvent(bot *tg.Bot, u tg.Update) error {
	fields := strings.SplitN(u.Message.Text, " ", 2)
	if len(fields) == 1 {
		return tg.SendMessageParams{
			Text: "faltando nome do evento",
		}
	}

	if u.Message.ReplyToMessage == nil {
		return tg.SendMessageParams{
			Text: "responda a mensagem que quer descontar",
		}
	}

	if u.Message.From.ID != godID {
		return tg.SendMessageParams{
			Text: "já disse pra sair, macaco",
		}
	}

	event := db.ChatEvent{
		ChatID: u.Message.Chat.ID,
		MsgID:  u.Message.ReplyToMessage.MessageID,
		Name:   strings.TrimSpace(fields[1]),
	}

	err := myDB.DeleteChatEvent(event)
	if err != nil {
		return err
	}

	return tg.SendMessageParams{
		Text: "descontey",
	}
}

func handleSpam(bot *tg.Bot, u tg.Update) error {
	panic("TODO")
	// if u.Message.From.ID != godID {
	// 	_, err := replyToMessage(bot, u.Message, &tg.SendMessageParams{
	// 		Text: "sai man so faço isso pro @igorcafe",
	// 	})
	// 	return err
	// }

	fields := strings.SplitN(u.Message.Text, " ", 3)
	if len(fields) != 3 {
		return tg.SendMessageParams{
			Text: "uso: /spam <quantidade> <mensagem>",
		}
	}

	count, err := strconv.Atoi(fields[1])
	if err != nil {
		return tg.SendMessageParams{
			Text: fmt.Sprintf("quantidade inválida: '%s'", fields[1]),
		}
	}

	limit := make(chan struct{}, 10)

	for i := 0; i < count; i++ {
		limit <- struct{}{}
		go func() {
			// _, err = bot.SendMessage(tg.SendMessageParams{
			// 	ChatID: u.Message.Chat.ID,
			// 	Text:   fields[2],
			// })
			// if err != nil {
			// 	log.Print(err)
			// }
			<-limit
		}()
	}
	return nil
}

func handlePollAnswer(bot *tg.Bot, u tg.Update) error {
	var err error

	if len(u.PollAnswer.OptionIDs) == 0 {
		err = myDB.DeletePollVote(u.PollAnswer.PollID, u.PollAnswer.User.ID)
	} else {
		err = myDB.SavePollVote(db.PollVote{
			PollID: u.PollAnswer.PollID,
			UserID: u.PollAnswer.User.ID,
			Vote:   u.PollAnswer.OptionIDs[0],
		})
	}
	if err != nil {
		return err
	}

	poll, err := myDB.FindPoll(u.PollAnswer.PollID)
	if err != nil {
		return err
	}

	users, err := myDB.FindUsersByTopic(poll.ChatID, poll.Topic)
	if err != nil {
		return err
	}

	found := false
	for _, user := range users {
		if user.ID == u.PollAnswer.User.ID {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	positiveCount := 0
	positives := ""
	negativeCount := 0
	negatives := ""
	remainingCount := 0
	remainings := ""

	for _, user := range users {
		mention := fmt.Sprintf("[%s](tg://user?id=%d)\n", user.Name(), user.ID)

		vote, err := myDB.FindPollVote(poll.ID, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			remainings += mention
			remainingCount++
			continue
		} else if err != nil {
			return err
		}

		const yes = 0
		const no = 1

		if vote.Vote == yes {
			positiveCount++
			positives += mention
		} else if vote.Vote == no {
			negativeCount++
			negatives += mention
		}
	}

	txt := fmt.Sprintf(
		"*sim \\(%d votos\\)*\n%s\n*não \\(%d votos\\)*\n%s\n*restam \\(%d votos\\)*\n%s",
		positiveCount,
		positives,
		negativeCount,
		negatives,
		remainingCount,
		remainings,
	)

	_, err = bot.EditMessageText(tg.EditMessageTextParams{
		ChatID:    poll.ChatID,
		MessageID: poll.ResultMessageID,
		Text:      txt,
		ParseMode: "MarkdownV2",
	})
	return err
}

// TODO: per chat
var lastVoice atomic.Int32

func handleTextMessage(bot *tg.Bot, u tg.Update) error {
	questions := []string{"and", "e?", "askers", "askers?", "perguntadores", "perguntadores?"}
	found := false
	for _, q := range questions {
		if u.Message.Text == q {
			found = true
			break
		}
	}

	if found {
		msgID := 0
		if u.Message.ReplyToMessage != nil {
			msgID = u.Message.ReplyToMessage.MessageID
		}
		_, err := bot.SendMessage(tg.SendMessageParams{
			ChatID:                   u.Message.Chat.ID,
			Text:                     "perguntadores not found",
			ReplyToMessageID:         msgID,
			AllowSendingWithoutReply: true,
		})
		return err
	}

	n := lastVoice.Add(1)
	if n > 50 {
		lastVoice.Swap(0)
	}

	if n%10 == 0 {
		log.Printf("%d messages remaining", 50-n)
	}

	if lastVoice.CompareAndSwap(50, 0) {
		b, err := os.ReadFile("./audio_ids.txt")
		if err != nil {
			return err
		}

		lines := bytes.Split(b, []byte("\n"))
		line := lines[rand.Intn(len(lines))]

		log.Println("sending voice: ", string(line))

		_, err = bot.SendVoice(tg.SendVoiceParams{
			ChatID:           u.Message.Chat.ID,
			Voice:            string(line),
			ReplyToMessageID: u.Message.MessageID,
		})

		return err
	}

	return nil
}

	params.ChatID = msg.Chat.ID
	params.ReplyToMessageID = msg.MessageID

	return bot.SendMessage(*params)
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
