package main

import (
	"context"
	"log"

	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/handler"
	"github.com/igoracmelo/euperturbot/oai"
	"github.com/igoracmelo/euperturbot/tg"
	_ "modernc.org/sqlite"
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
	err = db.Migrate(context.Background(), "./db/migrations")
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
	c.HandleCommand("enable_create_topics", h.RequireAdmin(h.Enable("create_topics")))
	c.HandleCommand("disable_create_topics", h.RequireAdmin(h.Disable("create_topics")))
	c.HandleCommand("enable_audio", h.RequireAdmin(h.Enable("audio")))
	c.HandleCommand("disable_audio", h.RequireAdmin(h.Disable("audio")))
	c.HandleCommand("enable_ask", h.RequireAdmin(h.Enable("ask")))
	c.HandleCommand("disable_ask", h.RequireAdmin(h.Disable("ask")))
	c.HandleCommand("enable_cask", h.RequireAdmin(h.Enable("cask")))
	c.HandleCommand("disable_cask", h.RequireAdmin(h.Disable("cask")))
	c.HandleCommand("enable_sed", h.RequireAdmin(h.Enable("sed")))
	c.HandleCommand("disable_sed", h.RequireAdmin(h.Disable("sed")))

	c.HandleText(h.Text)

	c.Start()
}
