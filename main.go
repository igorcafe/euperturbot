package main

import (
	"context"
	"log"

	_ "github.com/glebarez/go-sqlite"
	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/handler"
	"github.com/igoracmelo/euperturbot/oai"
	"github.com/igoracmelo/euperturbot/tg"
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
	c := tg.NewUpdateController(bot, updates)

	h := handler.Handler{
		DB:      myDB,
		OAI:     myOAI,
		BotInfo: botInfo,
		Config:  &conf,
	}

	c.HandleCommand("suba", h.SubToTopic)
	c.HandleCommand("desca", h.UnsubTopic)
	c.HandleCommand("pollo", h.CreatePoll)
	c.HandleCommand("bora", h.CallSubs)
	c.HandleCommand("quem", h.ListSubs)
	c.HandleCommand("lista", h.ListUserTopics)
	c.HandleCommand("listudo", h.ListChatTopics)
	c.HandleCommand("conta", h.CountEvent)
	c.HandleCommand("desconta", h.UncountEvent)
	c.HandleCommand("a", h.SaveAudio)
	c.HandleCommand("arand", h.SendRandomAudio)
	c.HandleCommand("ask", h.GPTCompletion)
	c.HandleCommand("cask", h.GPTChatCompletion)
	c.HandleCallbackQuery(h.CallbackQuery)
	c.HandleInlineQuery(h.InlineQuery)
	c.HandleText(h.Text)
	c.Start()
}
