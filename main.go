package main

import (
	"context"
	"log"
	"net/http"

	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/handler"
	"github.com/igoracmelo/euperturbot/openai"
	"github.com/igoracmelo/euperturbot/tg"
	"github.com/igoracmelo/euperturbot/tg/tgh"
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
	defer db.Close()

	err = db.Migrate(context.Background(), "./db/migrations")
	if err != nil {
		panic(err)
	}

	oai := openai.NewService(conf.OpenAIKey, http.DefaultClient)

	bot := tg.NewBot(conf.BotToken)
	if err != nil {
		panic(err)
	}

	botInfo, err := bot.GetMe()
	if err != nil {
		panic(err)
	}

	h := handler.Handler{
		DB:      db,
		OpenAI:  oai,
		BotInfo: botInfo,
		Config:  &conf,
	}

	updates := bot.GetUpdatesChannel()
	c := tgh.NewUpdateController(bot, updates)

	c.Middleware(h.EnsureStarted(), tgh.AnyMessage)
	c.Middleware(h.IgnoreForwardedCommand(), tgh.AnyCommand)

	c.Handle(tgh.Command("start"), h.RequireAdmin(h.Start))
	c.Handle(tgh.Command("suba"), h.SubToTopic)
	c.Handle(tgh.Command("desca"), h.UnsubTopic)
	c.Handle(tgh.Command("pollo"), h.CreatePoll)
	c.Handle(tgh.Command("bora"), h.CallSubs)
	c.Handle(tgh.Command("quem"), h.ListSubs)
	c.Handle(tgh.Command("lista"), h.ListUserTopics)
	c.Handle(tgh.Command("listudo"), h.ListChatTopics)
	// c.Handle(tgh.Command("conta"), h.CountEvent)
	// c.Handle(tgh.Command("desconta"), h.UncountEvent)
	c.Handle(tgh.Command("a"), h.SaveAudio)
	c.Handle(tgh.Command("arand"), h.SendRandomAudio)
	c.Handle(tgh.Command("ask"), h.GPTCompletion)
	c.Handle(tgh.Command("cask"), h.GPTChatCompletion)
	c.Handle(tgh.Command("backup"), h.RequireGod(h.Backup))
	c.Handle(tgh.Command("xonotic"), h.Xonotic)
	c.Handle(tgh.AnyCallbackQuery, h.CallbackQuery)
	c.Handle(tgh.AnyInlineQuery, h.InlineQuery)

	// switches
	c.Handle(tgh.Command("enable_create_topics"), h.RequireAdmin(h.Enable("create_topics")))
	c.Handle(tgh.Command("disable_create_topics"), h.RequireAdmin(h.Disable("create_topics")))
	c.Handle(tgh.Command("enable_audio"), h.RequireAdmin(h.Enable("audio")))
	c.Handle(tgh.Command("disable_audio"), h.RequireAdmin(h.Disable("audio")))
	c.Handle(tgh.Command("enable_ask"), h.RequireAdmin(h.Enable("ask")))
	c.Handle(tgh.Command("disable_ask"), h.RequireAdmin(h.Disable("ask")))
	c.Handle(tgh.Command("enable_cask"), h.RequireAdmin(h.Enable("cask")))
	c.Handle(tgh.Command("disable_cask"), h.RequireAdmin(h.Disable("cask")))
	c.Handle(tgh.Command("enable_sed"), h.RequireAdmin(h.Enable("sed")))
	c.Handle(tgh.Command("disable_sed"), h.RequireAdmin(h.Disable("sed")))

	c.Handle(tgh.AnyText, h.Text)

	c.Start()
}
