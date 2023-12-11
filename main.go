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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error

	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	db, err := db.NewSqlite("euperturbot.db")
	if err != nil {
		panic(err)
	}
	err = db.Migrate(context.Background())
	if err != nil {
		panic(err)
	}

	oai := oai.NewClient(conf.OpenAIKey)

	bot := tg.NewBot(conf.BotToken)
	if err != nil {
		panic(err)
	}

	botInfo, err := bot.GetMe()
	if err != nil {
		panic(err)
	}

	updates := bot.GetUpdatesChannel()
	c := tg.NewUpdateController(bot, updates)

	h := handler.Handler{
		DB:      db,
		OAI:     oai,
		BotInfo: botInfo,
		Config:  &conf,
	}

	c.Middleware(h.EnsureStarted(), tg.AnyMessage)
	c.Middleware(h.IgnoreForwardedCommand(), tg.AnyCommand)

	c.HandleCommand("start", h.RequireAdmin(h.Start))
	c.HandleCommand("suba", h.SubToTopic)
	c.HandleCommand("desca", h.UnsubTopic)
	c.HandleCommand("pollo", h.CreatePoll)
	c.HandleCommand("bora", h.CallSubs)
	c.HandleCommand("quem", h.ListSubs)
	c.HandleCommand("lista", h.ListUserTopics)
	c.HandleCommand("listudo", h.ListChatTopics)
	// c.HandleCommand("conta", h.CountEvent)
	// c.HandleCommand("desconta", h.UncountEvent)
	c.HandleCommand("a", h.SaveAudio)
	c.HandleCommand("arand", h.SendRandomAudio)
	c.HandleCommand("ask", h.GPTCompletion)
	c.HandleCommand("cask", h.GPTChatCompletion)
	c.HandleCommand("backup", h.RequireGod(h.Backup))
	c.HandleCallbackQuery(h.CallbackQuery)
	c.HandleInlineQuery(h.InlineQuery)

	// switches
	c.HandleCommand("enablecreatetopics", h.RequireAdmin(h.Enable("create_topics")))
	c.HandleCommand("disablecreatetopics", h.RequireAdmin(h.Disable("create_topics")))
	c.HandleCommand("enableaudio", h.RequireAdmin(h.Enable("audio")))
	c.HandleCommand("disableaudio", h.RequireAdmin(h.Disable("audio")))
	c.HandleCommand("enablecask", h.RequireAdmin(h.Enable("cask")))
	c.HandleCommand("disablecask", h.RequireAdmin(h.Disable("cask")))
	c.HandleCommand("enablesed", h.RequireAdmin(h.Enable("sed")))
	c.HandleCommand("disablesed", h.RequireAdmin(h.Disable("sed")))

	c.HandleText(h.Text)

	c.Start()
}
