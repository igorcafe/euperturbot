package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	_ "github.com/glebarez/go-sqlite"
	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/handler"
	"github.com/igoracmelo/euperturbot/oai"
	"github.com/igoracmelo/euperturbot/tg"
	"github.com/igoracmelo/euperturbot/util"
)

var myDB *db.DB
var myOAI *oai.Client
var botInfo *tg.User

var conf config.Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error

	conf, err = config.Load()
	if err != nil {
		panic(err)
	}

	myDB, err = db.NewSqlite("euperturbot.db")
	if err != nil {
		panic(err)
	}
	err = myDB.Migrate(context.Background())
	if err != nil {
		panic(err)
	}

	myOAI = oai.NewClient(conf.OpenAIKey)

	bot := tg.NewBot(conf.BotToken)
	if err != nil {
		panic(err)
	}

	botInfo, err = bot.GetMe()
	if err != nil {
		panic(err)
	}

	updates := bot.GetUpdatesChannel()
	h := tg.NewUpdateHandler(bot, updates)

	h2 := handler.Handler{
		DB:      myDB,
		OAI:     myOAI,
		BotInfo: botInfo,
		Config:  &conf,
	}

	h.HandleCommand("suba", h2.SubToTopic)
	h.HandleCommand("desca", handleUnsubTopic)
	h.HandleCommand("pollo", handleCreatePoll)
	h.HandleCommand("bora", handleCallSubs)
	h.HandleCommand("quem", handleListSubs)
	h.HandleCommand("lista", handleListUserTopics)
	h.HandleCommand("listudo", handleListChatTopics)
	h.HandleCommand("conta", handleCountEvent)
	h.HandleCommand("desconta", handleUncountEvent)
	h.HandleCommand("a", handleSaveAudio)
	h.HandleCommand("arand", handleSendRandomAudio)
	h.HandleCommand("ask", handleGPTCompletion)
	h.HandleCommand("cask", handleGPTChatCompletion)
	h.HandleCallbackQuery(handleCallbackQuery)
	h.HandleInlineQuery(handleInlineQuery)
	h.HandleText(handleText)
	h.HandleMessage(handleMessage)
	// h.HandleTextEqual([]string{"and", "e", "and?", "e?", "askers", "askers?"}, handleAskers)
	h.Start()
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
		return tg.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   err.Error(),
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

	return callSubs(bot, u, topic, false)
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
		if u.Message.From.ID != conf.GodID {
			return tg.SendMessageParams{
				Text: "sai maluco",
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

	last := time.Since(events[0].Time)
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

	if u.Message.From.ID != conf.GodID {
		return tg.SendMessageParams{
			Text: "já disse pra sair, maluco",
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

func handleSaveAudio(bot *tg.Bot, u tg.Update) error {
	if u.Message.ReplyToMessage == nil {
		return tg.SendMessageParams{
			Text: "responda ao audio que quer salvar",
		}
	}

	if u.Message.ReplyToMessage.Voice == nil {
		return tg.SendMessageParams{
			Text: "tem que ser uma mensagem de voz",
		}
	}

	err := myDB.SaveVoice(db.Voice{
		FileID: u.Message.ReplyToMessage.Voice.FileID,
		UserID: u.Message.ReplyToMessage.From.ID,
	})
	if err != nil {
		return err
	}

	return tg.SendMessageParams{
		Text: "áudio salvo",
	}
}

func handleSendRandomAudio(bot *tg.Bot, u tg.Update) error {
	voice, err := myDB.FindRandomVoice()
	if errors.Is(err, sql.ErrNoRows) {
		return tg.SendMessageParams{
			Text: "nenhum áudio salvo para mandar",
		}
	}
	if err != nil {
		return err
	}
	_, err = bot.SendVoice(tg.SendVoiceParams{
		ChatID:           u.Message.Chat.ID,
		Voice:            voice.FileID,
		ReplyToMessageID: u.Message.MessageID,
	})
	return err
}

func handleText(bot *tg.Bot, u tg.Update) error {
	txt := u.Message.Text
	if strings.HasPrefix(txt, "/") {
		return nil
	}

	name := username(u.Message.From)
	id := u.Message.From.ID

	if u.Message.FowardFrom != nil {
		name = username(u.Message.FowardFrom) + "(forward)"
		id = u.Message.FowardFrom.ID
	}

	if u.Message.ForwardSenderName != "" {
		id = 0
		name = sanitizeUsername(u.Message.ForwardSenderName)
	}

	replyID := 0
	if u.Message.ReplyToMessage != nil {
		replyID = u.Message.ReplyToMessage.MessageID
	}

	err := myDB.SaveMessage(db.Message{
		ID:               u.Message.MessageID,
		ReplyToMessageID: replyID,
		ChatID:           u.Message.Chat.ID,
		Text:             txt,
		Date:             time.Unix(u.Message.Date, 0),
		UserID:           id,
		UserName:         name,
	})

	return err
}

var messageCount atomic.Int32

func handleMessage(bot *tg.Bot, u tg.Update) error {
	t := u.Message.Text
	t = strings.TrimSpace(t)
	if strings.HasPrefix(t, "#") {
		if err := validateTopic(t); err != nil {
			return nil
		}
		return callSubs(bot, u, t, true)
	}

	date := time.Unix(u.Message.Date, 0)
	if time.Since(date).Minutes() > 1 {
		return nil
	}

	n := messageCount.Add(1)
	const target = 150
	if n > target {
		messageCount.Store(0)
	}

	if n%10 == 0 {
		fmt.Printf("%d messages remaining\n", target-n)
	}

	if messageCount.CompareAndSwap(target, 0) {
		voice, err := myDB.FindRandomVoice()
		if err != nil {
			return err
		}
		_, err = bot.SendVoice(tg.SendVoiceParams{
			ChatID:           u.Message.Chat.ID,
			Voice:            voice.FileID,
			ReplyToMessageID: u.Message.MessageID,
		})
		return err
	}

	return nil
}

func handleGPTCompletion(bot *tg.Bot, u tg.Update) error {
	chunks := strings.SplitN(u.Message.Text, " ", 2)
	if len(chunks) != 2 {
		return tg.SendMessageParams{
			ReplyToMessageID: u.Message.MessageID,
			Text:             "faltou a pergunta",
		}
	}

	name := username(u.Message.From)

	msg, err := bot.SendMessage(tg.SendMessageParams{
		ChatID:           u.Message.Chat.ID,
		ReplyToMessageID: u.Message.MessageID,
		Text:             "Carregando...",
	})
	if err != nil {
		return err
	}

	resp, err := myOAI.Completion(&oai.CompletionParams{
		Messages: []oai.Message{
			{
				Content: fmt.Sprintf(
					"Meu nome é %s. Me responda com @%s.\n%s",
					name,
					name,
					chunks[1],
				),
			},
		},
	})
	if errors.Is(err, oai.ErrRateLimit) {
		_, err = bot.EditMessageText(tg.EditMessageTextParams{
			ChatID:    u.Message.Chat.ID,
			MessageID: msg.MessageID,
			Text:      "ignorated kk rate limit",
		})
		return err
	}
	if err != nil {
		_, _ = bot.EditMessageText(tg.EditMessageTextParams{
			ChatID:    u.Message.Chat.ID,
			MessageID: msg.MessageID,
			Text:      "vish deu ruim",
		})
		return err
	}

	_, err = bot.EditMessageText(tg.EditMessageTextParams{
		ChatID:    u.Message.Chat.ID,
		MessageID: msg.MessageID,
		Text:      resp.Choices[0].Message.Content,
		ParseMode: "MarkdownV2",
	})

	var boterr tg.BotError
	if errors.As(err, &boterr); boterr.Status == http.StatusBadRequest {
		_, err = bot.EditMessageText(tg.EditMessageTextParams{
			ChatID:    u.Message.Chat.ID,
			MessageID: msg.MessageID,
			Text:      resp.Choices[0].Message.Content,
		})
	}

	return err
}

func handleGPTChatCompletion(bot *tg.Bot, u tg.Update) error {
	// if u.Message.From.ID != godID {
	// 	return tg.SendMessageParams{
	// 		ReplyToMessageID: u.Message.MessageID,
	// 		Text:             "desculpe mas nao posso te deixar gastar os creditos do igorcafe",
	// 	}
	// }

	chunks := strings.SplitN(u.Message.Text, " ", 2)
	if len(chunks) != 2 {
		return tg.SendMessageParams{
			ReplyToMessageID: u.Message.MessageID,
			Text:             "faltou a pergunta",
		}
	}

	date := time.Unix(u.Message.Date, 0)
	if u.Message.ReplyToMessage != nil {
		date = time.Unix(u.Message.ReplyToMessage.Date, 0)
	}

	msgs, err := myDB.FindMessagesBeforeDate(u.Message.Chat.ID, date, 100)
	if err != nil {
		return err
	}

	name := username(u.Message.From)
	title := u.Message.Chat.Title
	if title == "" {
		title = u.Message.Chat.FirstName
	}

	prompts := []oai.Message{
		{
			Content: fmt.Sprintf(
				"Mensagens recentes do chat %s para voce se contextualizar, no formato '<usuario>: <texto>'\n\n%s",
				title,
				strings.Join(prepareTextForGPT(msgs), "\n"),
			),
		},
		{
			Content: fmt.Sprintf(
				"Me chame de @%s e responda a mensagem abaixo. Se baseie no historico de mensagens acima e nos nomes de usuarios para responde. NÃO crie diálogos, apenas me responda com as informações fornecidas. As palavras 'grupo', 'chat', 'conversa', 'historico' todas se referecem ao historico do chat %s acima. Nao mencione o nome do grupo. Responda a seguinte me mencionando em segunda pessoa, usando @%s\n\n%s",
				name,
				title,
				name,
				chunks[1],
			),
		},
	}

	// for _, p := range prompts {
	// 	fmt.Println(p)
	// 	fmt.Println()
	// }

	msg, err := bot.SendMessage(tg.SendMessageParams{
		ChatID:           u.Message.Chat.ID,
		ReplyToMessageID: u.Message.MessageID,
		Text:             "Carregando...",
	})
	if err != nil {
		return err
	}

	resp, err := myOAI.Completion(&oai.CompletionParams{
		Messages:    prompts,
		Temperature: 0.5,
	})
	if errors.Is(err, oai.ErrRateLimit) {
		return tg.SendMessageParams{
			ReplyToMessageID: u.Message.MessageID,
			Text:             "ignorated kk rate limit",
		}
	}
	if err != nil {
		return err
	}

	_, err = bot.EditMessageText(tg.EditMessageTextParams{
		ChatID:    u.Message.Chat.ID,
		MessageID: msg.MessageID,
		Text:      resp.Choices[0].Message.Content,
	})
	return err
}

func handleInlineQuery(bot *tg.Bot, u tg.Update) error {
	return bot.AnswerInlineQuery(tg.AnswerInlineQueryParams{
		InlineQueryID: u.InlineQuery.ID,
		Results: []tg.InlineQueryResult{
			{
				Type:  "article",
				ID:    "1",
				Title: "oieoioei",
				InputMessageContent: tg.InputMessageContent{
					MessageText: "oioioio",
				},
			},
		},
	})
}

func handleCallbackQuery(bot *tg.Bot, u tg.Update) error {
	var err error

	poll, err := myDB.FindPollByMessage(u.CallbackQuery.Message.MessageID)
	if err != nil {
		return err
	}

	voteNum, err := strconv.Atoi(u.CallbackQuery.Data)
	if err != nil {
		return err
	}

	// TODO: improve this logic
	vote, err := myDB.FindPollVote(poll.ID, u.CallbackQuery.From.ID)
	if errors.Is(err, sql.ErrNoRows) {
		vote = nil
	} else if err != nil {
		return err
	}

	if vote != nil && vote.Vote == voteNum {
		err = myDB.DeletePollVote(vote.PollID, vote.UserID)
	} else {
		err = myDB.SavePollVote(db.PollVote{
			PollID: poll.ID,
			UserID: u.CallbackQuery.From.ID,
			Vote:   voteNum,
		})
	}
	if err != nil {
		return err
	}

	users, err := myDB.FindUsersByTopic(poll.ChatID, poll.Topic)
	if err != nil {
		return err
	}

	found := false
	for _, user := range users {
		if user.ID == u.CallbackQuery.From.ID {
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

	up := "👍 " + fmt.Sprint(positiveCount)
	down := "👎 " + fmt.Sprint(negativeCount)

	_, err = bot.EditMessageText(tg.EditMessageTextParams{
		ChatID:    poll.ChatID,
		MessageID: poll.ResultMessageID,
		Text:      txt,
		ParseMode: "MarkdownV2",
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
		// log.Printf("%d messages remaining", 50-n)
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

func callSubs(bot *tg.Bot, u tg.Update, topic string, quiet bool) error {
	users, err := myDB.FindUsersByTopic(u.Message.Chat.ID, topic)
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

	err = myDB.SavePoll(db.Poll{
		ID:              strconv.Itoa(msg.MessageID),
		ChatID:          u.Message.Chat.ID,
		Topic:           topic,
		ResultMessageID: msg.MessageID,
	})

	return err
}

func prepareTextForGPT(msgs []db.Message) []string {
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
