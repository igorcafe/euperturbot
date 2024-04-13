package main

import (
	"context"
	"log"
	"net/http"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/db"
	"github.com/igoracmelo/euperturbot/handler"
	"github.com/igoracmelo/euperturbot/openai"
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

	bot := bot.NewService(conf.BotToken)
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
	c := bh.NewUpdateController(bot, updates)

	c.Middleware(h.EnsureStarted(), bh.AnyMessage)
	c.Middleware(h.IgnoreForwardedCommand(), bh.AnyCommand)

	c.Handle(bh.Command("start"), h.RequireAdmin(h.Start))
	c.Handle(bh.Command("suba"), h.SubToTopic)
	c.Handle(bh.Command("desca"), h.UnsubTopic)
	c.Handle(bh.Command("pollo"), h.CreatePoll)
	c.Handle(bh.Command("bora"), h.CallSubs)
	c.Handle(bh.Command("quem"), h.ListSubs)
	c.Handle(bh.Command("lista"), h.ListUserTopics)
	c.Handle(bh.Command("listudo"), h.ListChatTopics)
	// c.Handle(tgh.Command("conta"), h.CountEvent)
	// c.Handle(tgh.Command("desconta"), h.UncountEvent)
	c.Handle(bh.Command("a"), h.SaveAudio)
	c.Handle(bh.Command("arand"), h.SendRandomAudio)
	c.Handle(bh.Command("ask"), h.GPTCompletion)
	c.Handle(bh.Command("cask"), h.GPTChatCompletion)
	c.Handle(bh.Command("backup"), h.RequireGod(h.Backup))
	c.Handle(bh.Command("xonotic"), h.Xonotic)
	c.Handle(bh.AnyCallbackQuery, h.CallbackQuery)
	c.Handle(bh.AnyInlineQuery, h.InlineQuery)

	// switches
	c.Handle(bh.Command("enable_create_topics"), h.RequireAdmin(h.Enable("create_topics")))
	c.Handle(bh.Command("disable_create_topics"), h.RequireAdmin(h.Disable("create_topics")))
	c.Handle(bh.Command("enable_audio"), h.RequireAdmin(h.Enable("audio")))
	c.Handle(bh.Command("disable_audio"), h.RequireAdmin(h.Disable("audio")))
	c.Handle(bh.Command("enable_ask"), h.RequireAdmin(h.Enable("ask")))
	c.Handle(bh.Command("disable_ask"), h.RequireAdmin(h.Disable("ask")))
	c.Handle(bh.Command("enable_cask"), h.RequireAdmin(h.Enable("cask")))
	c.Handle(bh.Command("disable_cask"), h.RequireAdmin(h.Disable("cask")))
	c.Handle(bh.Command("enable_sed"), h.RequireAdmin(h.Enable("sed")))
	c.Handle(bh.Command("disable_sed"), h.RequireAdmin(h.Disable("sed")))

	c.Handle(bh.AnyText, h.Text)

	c.Start()
}
