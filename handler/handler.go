package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/openai"
	"github.com/igoracmelo/euperturbot/tg"
	"github.com/igoracmelo/euperturbot/tg/tgh"
	"github.com/igoracmelo/euperturbot/util"
)

type Handler struct {
	DB      *db.DB
	OpenAI  openai.Service
	BotInfo *tg.User
	Config  *config.Config
}

func (h Handler) Start(bot *tg.Bot, u tg.Update) error {
	err := h.DB.SaveChat(context.TODO(), db.Chat{
		ID:    u.Message.Chat.ID,
		Title: u.Message.Chat.Name(),
	})
	if err != nil {
		return err
	}

	_, err = bot.SendMessage(tg.SendMessageParams{
		ChatID:                   u.Message.Chat.ID,
		ReplyToMessageID:         u.Message.MessageID,
		Text:                     "vamo que vamo",
		AllowSendingWithoutReply: true,
	})
	return err
}

func (h Handler) SubToTopic(bot *tg.Bot, u tg.Update) error {
	fields := strings.SplitN(u.Message.Text, " ", 2)
	topics := []string{}
	if len(fields) > 1 {
		topics = strings.Split(fields[1], "\n")
	}

	if len(topics) == 0 {
		return tgh.Reply{
			Text: "cadê o(s) tópico(s)?",
		}
	}

	if len(topics) > 3 {
		return tgh.Reply{
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
			return tgh.Reply{
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

		enablesCreatingTopic, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "create_topics")
		isAdmin, _ := h.isAdmin(bot, u)
		if !exists && !isAdmin && !enablesCreatingTopic {
			return tgh.Reply{
				Text: "você só tem permissão para se inscrever em tópicos existentes",
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
			return tgh.Reply{
				Text: "falha ao salvar tópico " + topic,
			}
		}
	}

	txt := fmt.Sprintf("inscrições adicionadas para %s:\n", user.Name())
	for _, topic := range topics {
		txt += fmt.Sprintf("- %s\n", topic)
	}
	return tgh.Reply{
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
		return tgh.Reply{
			Text: fmt.Sprintf("usuário %s não está inscrito nesse tópico", user.Name()),
		}
	}

	return tgh.Reply{
		Text: "inscrição removida para " + user.Name(),
	}
}

func (h Handler) CreatePoll(bot *tg.Bot, u tg.Update) error {
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

func (h Handler) CallSubs(bot *tg.Bot, u tg.Update) error {
	log.Print(username(u.Message.From) + ": " + u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		return tgh.Reply{
			Text: err.Error(),
		}
	}

	return h.callSubs(bot, u, topic, false)
}

func (h Handler) ListSubs(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	fields := strings.SplitN(u.Message.Text, " ", 2)
	topic := ""
	if len(fields) > 1 {
		topic = fields[1]
	}

	if err := validateTopic(topic); err != nil {
		return tgh.Reply{
			Text: err.Error(),
		}
	}

	users, err := h.DB.FindUsersByTopic(u.Message.Chat.ID, topic)
	if err != nil {
		return tgh.Reply{
			Text: "falha ao listar usuários",
		}
	}

	if len(users) == 0 {
		return tgh.Reply{
			Text: "não tem ninguém inscrito nesse tópico",
		}
	}

	txt := fmt.Sprintf("*inscritos \\(%d\\)*\n", len(users))
	for _, user := range users {
		txt += fmt.Sprintf("\\- %s\n", user.Name())
	}
	return tgh.Reply{
		Text:      txt,
		ParseMode: "MarkdownV2",
	}
}

func (h Handler) ListUserTopics(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	topics, err := h.DB.FindUserChatTopics(u.Message.Chat.ID, u.Message.From.ID)
	if err != nil {
		return tgh.Reply{
			Text: "falha ao listar tópicos",
		}
	}

	if len(topics) == 0 {
		return tgh.Reply{
			Text: "você não está inscrito em nenhum tópico",
		}
	}

	txt := "seus tópicos:\n"
	for _, topic := range topics {
		txt += fmt.Sprintf("(%02d)  %s\n", topic.Subscribers, topic.Topic)
	}

	return tgh.Reply{
		Text: txt,
	}
}

func (h Handler) ListChatTopics(bot *tg.Bot, u tg.Update) error {
	log.Print(u.Message.Text)

	topics, err := h.DB.FindChatTopics(u.Message.Chat.ID)
	if err != nil {
		log.Print(err)
		return tgh.Reply{
			Text: "falha ao listar tópicos",
		}
	}

	if len(topics) == 0 {
		return tgh.Reply{
			Text: "não existe nenhum tópico registrado nesse chat",
		}
	}

	txt := "tópicos:\n"
	for _, topic := range topics {
		txt += fmt.Sprintf("- (%02d)  %s\n", topic.Subscribers, topic.Topic)
	}

	return tgh.Reply{
		Text: txt,
	}
}

func (h Handler) SaveAudio(bot *tg.Bot, u tg.Update) error {
	enables, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "audio")
	if !enables {
		return tgh.Reply{
			Text: "comando desativado. ative com /enable_audio",
		}
	}

	if u.Message.ReplyToMessage == nil {
		return tgh.Reply{
			Text: "responda ao audio que quer salvar",
		}
	}

	if u.Message.ReplyToMessage.Voice == nil {
		return tgh.Reply{
			Text: "tem que ser uma mensagem de voz",
		}
	}

	err := h.DB.SaveVoice(db.Voice{
		FileID: u.Message.ReplyToMessage.Voice.FileID,
		UserID: u.Message.ReplyToMessage.From.ID,
		ChatID: u.Message.Chat.ID,
	})
	if err != nil {
		return err
	}

	return tgh.Reply{
		Text: "áudio salvo",
	}
}

func (h Handler) SendRandomAudio(bot *tg.Bot, u tg.Update) error {
	enables, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "audio")
	if !enables {
		return tgh.Reply{
			Text: "comando desativado. ative com /enable_audio",
		}
	}

	voice, err := h.DB.FindRandomVoice(u.Message.Chat.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return tgh.Reply{
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

func (h Handler) gptCompletion(bot *tg.Bot, u tg.Update, msgs []openai.Message) error {
	msg, err := bot.SendMessage(tg.SendMessageParams{
		ChatID:           u.Message.Chat.ID,
		ReplyToMessageID: u.Message.MessageID,
		Text:             "Carregando...",
	})
	if err != nil {
		return err
	}

	resp, err := h.OpenAI.Completion(&openai.CompletionParams{
		Messages: msgs,
	})

	var rateErr openai.ErrRateLimit
	if errors.As(err, &rateErr) {
		_, err = bot.EditMessageText(tg.EditMessageTextParams{
			ChatID:    u.Message.Chat.ID,
			MessageID: msg.MessageID,
			Text:      fmt.Sprintf("ignorated kk rate limit (%ds)", int(rateErr)),
		})
		go func() {
			deadline := time.Now().Add(time.Duration(rateErr) * time.Second)
			for time.Now().Before(deadline) {
				time.Sleep(time.Second)
				secs := int(time.Until(deadline).Seconds())
				_, _ = bot.EditMessageText(tg.EditMessageTextParams{
					ChatID:    u.Message.Chat.ID,
					MessageID: msg.MessageID,
					Text:      fmt.Sprintf("ignorated kk rate limit (%ds)", secs),
				})
			}
			_, _ = bot.EditMessageText(tg.EditMessageTextParams{
				ChatID:    u.Message.Chat.ID,
				MessageID: msg.MessageID,
				Text:      "manda de novo ae",
			})
		}()
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

	msg, err = bot.EditMessageText(tg.EditMessageTextParams{
		ChatID:    u.Message.Chat.ID,
		MessageID: msg.MessageID,
		Text:      resp.Choices[0].Message.Content,
	})
	if err != nil {
		return err
	}

	txt := strings.TrimPrefix(strings.TrimPrefix(u.Message.Text, "/cask "), "/ask ")

	replyTo := 0
	if u.Message.ReplyToMessage != nil {
		replyTo = u.Message.ReplyToMessage.MessageID
	}
	err = h.DB.SaveMessage(context.TODO(), db.Message{
		ID:               u.Message.MessageID,
		ChatID:           u.Message.Chat.ID,
		Text:             txt,
		Date:             time.Unix(u.Message.Date, 0),
		UserID:           u.Message.From.ID,
		ReplyToMessageID: replyTo,
	})
	if err != nil {
		return err
	}

	err = h.DB.SaveMessage(context.TODO(), db.Message{
		ID:               msg.MessageID,
		ChatID:           msg.Chat.ID,
		Text:             msg.Text,
		Date:             time.Unix(msg.Date, 0),
		UserID:           h.Config.GPTUserID,
		ReplyToMessageID: u.Message.MessageID,
	})
	return err
}

func (h Handler) GPTCompletion(bot *tg.Bot, u tg.Update) error {
	enables, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "ask")
	if !enables {
		return tgh.Reply{
			Text: "comando desativado. ative com /enable_ask",
		}
	}

	chunks := strings.SplitN(u.Message.Text, " ", 2)
	if len(chunks) != 2 {
		return tgh.Reply{
			Text: "faltou a pergunta",
		}
	}

	name := username(u.Message.From)

	msgs := []openai.Message{
		{
			Content: fmt.Sprintf(
				"%s: %s",
				name,
				chunks[1],
			),
		},
	}

	return h.gptCompletion(bot, u, msgs)
}

func (h Handler) GPTChatCompletion(bot *tg.Bot, u tg.Update) error {
	enables, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "cask")
	if !enables {
		return tgh.Reply{
			Text: "comando desativado. ative com /enable_cask\nATENÇÃO! Ao ativar essa opção, as mensagens de texto serão salvas no banco de dados do bot",
		}
	}

	chunks := strings.SplitN(u.Message.Text, " ", 2)
	if len(chunks) != 2 {
		return tgh.Reply{
			Text: "faltou a pergunta",
		}
	}

	date := time.Unix(u.Message.Date, 0)
	if u.Message.ReplyToMessage != nil {
		date = time.Unix(u.Message.ReplyToMessage.Date, 0)
	}

	msgs, err := h.DB.FindMessagesBeforeDate(context.TODO(), u.Message.Chat.ID, date, 100)
	if err != nil {
		return err
	}

	name := username(u.Message.From)
	title := u.Message.Chat.Title
	if title == "" {
		title = u.Message.Chat.FirstName
	}

	prepMsgs := prepareMessagesForGPT(msgs)
	if len(prepMsgs) == 0 {
		return tgh.Reply{
			Text: "ainda não há mensagens salvas para usar o /cask",
		}
	}

	prompts := []openai.Message{
		{
			Content: fmt.Sprintf(
				"Mensagens recentes do chat %s para voce se contextualizar, no formato '<usuario>: <texto>'\n\n%s",
				title,
				strings.Join(prepMsgs, "\n"),
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

	msg, err := bot.SendMessage(tg.SendMessageParams{
		ChatID:           u.Message.Chat.ID,
		ReplyToMessageID: u.Message.MessageID,
		Text:             fmt.Sprintf("Carregando... (usando últimas %d mensagens de contexto)", len(prepMsgs)),
	})
	if err != nil {
		return err
	}

	resp, err := h.OpenAI.Completion(&openai.CompletionParams{
		Messages:    prompts,
		Temperature: 0.5,
	})
	var rateErr openai.ErrRateLimit
	if errors.As(err, &rateErr) {
		_, err = bot.EditMessageText(tg.EditMessageTextParams{
			ChatID:    u.Message.Chat.ID,
			MessageID: msg.MessageID,
			Text:      fmt.Sprintf("ignorated kk rate limit (%ds)", int(rateErr)),
		})
		go func() {
			deadline := time.Now().Add(time.Duration(rateErr) * time.Second)
			for time.Now().Before(deadline) {
				time.Sleep(time.Second)
				secs := int(time.Until(deadline).Seconds())
				_, _ = bot.EditMessageText(tg.EditMessageTextParams{
					ChatID:    u.Message.Chat.ID,
					MessageID: msg.MessageID,
					Text:      fmt.Sprintf("ignorated kk rate limit (%ds)", secs),
				})
			}
		}()
		return err
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

func (h Handler) Enable(opt string) tgh.HandlerFunc {
	return func(bot *tg.Bot, u tg.Update) error {
		err := h.DB.ChatEnable(context.TODO(), u.Message.Chat.ID, opt)
		if err != nil {
			return err
		}

		return tgh.Reply{
			Text: "ativado",
		}
	}
}

func (h Handler) Disable(opt string) tgh.HandlerFunc {
	return func(bot *tg.Bot, u tg.Update) error {
		err := h.DB.ChatDisable(context.TODO(), u.Message.Chat.ID, opt)
		if err != nil {
			return err
		}

		return tgh.Reply{
			Text: "desativado",
		}
	}
}

func (h Handler) Backup(bot *tg.Bot, u tg.Update) error {
	return bot.SendDocument(tg.SendDocumentParams{
		ChatID:   h.Config.GodID,
		FileName: "./euperturbot.db",
	})
}

// WIP
func (h Handler) Xonotic(bot *tg.Bot, u tg.Update) error {
	type XonoticResponse []struct {
		Status        string
		Name          string
		Gametype      string
		Map           string
		Numplayers    int
		Numspectators int
		Ping          int
		Players       []struct {
			Name  string
			Score int
			Ping  int
		}
	}

	resp, err := http.Get("http://dpmaster.deathmask.net/?game=xonotic&server=129.148.22.240:27000&hide=empty&json=1")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data XonoticResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}
	if len(data) < 1 {
		return nil
	}
	server := data[0]

	players := ""
	// if len(server.Players) > 0 {
	// 	players = "*players:*\n"
	// }
	for _, p := range server.Players {
		players += fmt.Sprintf("%s \\- %d ms \\- %d pts\n", util.EscapeMarkdown(p.Name), p.Ping, p.Score)
	}

	txt := fmt.Sprintf(
		"*%s*\n%d players\n%d spectators\n\n%s",
		util.EscapeMarkdown(server.Name),
		server.Numplayers,
		server.Numspectators,
		players,
	)

	_, err = bot.SendMessage(tg.SendMessageParams{
		ChatID:           u.Message.Chat.ID,
		ReplyToMessageID: u.Message.MessageID,
		Text:             txt,
		ParseMode:        "MarkdownV2",
	})
	return err
}

func (h Handler) CallbackQuery(bot *tg.Bot, u tg.Update) error {
	var err error

	poll, err := h.DB.FindPollByMessage(u.CallbackQuery.Message.MessageID)
	if err != nil {
		return err
	}

	voteNum, err := strconv.Atoi(u.CallbackQuery.Data)
	if err != nil {
		return err
	}

	// TODO: improve this logic
	vote, err := h.DB.FindPollVote(poll.ID, u.CallbackQuery.From.ID)
	if errors.Is(err, sql.ErrNoRows) {
		vote = nil
	} else if err != nil {
		return err
	}

	if vote != nil && vote.Vote == voteNum {
		err = h.DB.DeletePollVote(vote.PollID, vote.UserID)
	} else {
		err = h.DB.SavePollVote(db.PollVote{
			PollID: poll.ID,
			UserID: u.CallbackQuery.From.ID,
			Vote:   voteNum,
		})
	}
	if err != nil {
		return err
	}

	if voteNum == db.VoteUp {
		err = h.DB.SaveUserTopic(db.UserTopic{
			ChatID: poll.ChatID,
			UserID: u.CallbackQuery.From.ID,
			Topic:  poll.Topic,
		})
		if err != nil {
			return err
		}
	}

	users, err := h.DB.FindUsersByTopic(poll.ChatID, poll.Topic)
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

		vote, err := h.DB.FindPollVote(poll.ID, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			remainings += mention
			remainingCount++
			continue
		} else if err != nil {
			return err
		}

		if vote.Vote == db.VoteUp {
			positiveCount++
			positives += mention
		} else if vote.Vote == db.VoteDown {
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

func (h Handler) Text(bot *tg.Bot, u tg.Update) error {
	// sed commands
	re := regexp.MustCompile(`^(s|y)\/.*\/`)
	if re.MatchString(u.Message.Text) && u.Message.ReplyToMessage != nil {
		enables, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "sed")
		if !enables {
			return nil
		}

		cmd := exec.CommandContext(context.TODO(), "sed", "--sandbox", "-E", u.Message.Text)
		buf := &bytes.Buffer{}
		cmd.Stdout = buf
		cmd.Stdin = strings.NewReader(u.Message.ReplyToMessage.Text)

		err := cmd.Run()
		if err != nil {
			return err
		}

		_, err = bot.SendMessage(tg.SendMessageParams{
			ChatID:           u.Message.Chat.ID,
			ReplyToMessageID: u.Message.ReplyToMessage.MessageID,
			Text:             buf.String(),
		})
		return err
	}

	// if reply to chatGPT, treat as /ask
	if u.Message.ReplyToMessage != nil && u.Message.ReplyToMessage.From.ID == h.BotInfo.ID {
		enables, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "ask")
		if !enables {
			return nil
		}

		msg, err := h.DB.FindMessage(context.TODO(), u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID)
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}
		if err != nil {
			return nil
		}

		if msg.UserID != h.Config.GPTUserID {
			return nil
		}

		msgs, err := h.DB.FindMessageThread(context.TODO(), u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID)
		if err != nil {
			return err
		}

		oaiMsgs := []openai.Message{}
		for _, msg := range msgs {
			role := "user"
			if msg.UserID == h.Config.GPTUserID {
				role = "assistant"
			}

			oaiMsgs = append(oaiMsgs, openai.Message{
				Role:    role,
				Content: msg.Text,
			})
		}

		name := username(u.Message.From)
		oaiMsgs = append(oaiMsgs, openai.Message{
			Content: fmt.Sprintf(
				"Meu nome é %s. Responda a seguinte mensagem se referindo a mim como @%s.\n%s",
				name,
				name,
				u.Message.Text,
			),
		})

		return h.gptCompletion(bot, u, oaiMsgs)
	}

	// call subscribers
	txt := strings.TrimSpace(u.Message.Text)
	if strings.HasPrefix(txt, "#") {
		if err := validateTopic(txt); err != nil {
			return nil
		}
		return h.callSubs(bot, u, txt, true)
	}

	// save message
	enables, _ := h.DB.ChatEnables(context.TODO(), u.Message.Chat.ID, "cask")
	if !enables {
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

	err := h.DB.SaveMessage(context.TODO(), db.Message{
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

// TODO:
func (h Handler) InlineQuery(bot *tg.Bot, u tg.Update) error {
	var err error

	// TODO: debounce by u.InlineQuery.ID
	util.Debounce(5*time.Second, func() {
		var resp *openai.CompletionResponse
		resp, err = h.OpenAI.Completion(&openai.CompletionParams{
			WaitRateLimit: true,
			Messages: []openai.Message{
				{
					Content: u.InlineQuery.Query,
				},
			},
		})

		if err != nil {
			_ = bot.AnswerInlineQuery(tg.AnswerInlineQueryParams{
				InlineQueryID: u.InlineQuery.ID,
				Results: []tg.InlineQueryResult{
					{
						Type:  "article",
						ID:    "1",
						Title: "Erro ao perguntar ao ChatGPT",
						InputMessageContent: tg.InputMessageContent{
							MessageText: "",
						},
					},
				},
			})
			return
		}

		title := resp.Choices[0].Message.Content
		if len(title) > 100 {
			title = title[:97] + "..."
		}

		err = bot.AnswerInlineQuery(tg.AnswerInlineQueryParams{
			InlineQueryID: u.InlineQuery.ID,
			Results: []tg.InlineQueryResult{
				{
					Type:  "article",
					ID:    fmt.Sprintf("%016X", rand.Int63()),
					Title: title,
					InputMessageContent: tg.InputMessageContent{
						MessageText: resp.Choices[0].Message.Content,
					},
				},
			},
		})
	})()

	return err
}
