package main

import (
	"context"
	"log"
	"net/http"

	"github.com/igoracmelo/euperturbot/bot"
	bh "github.com/igoracmelo/euperturbot/bot/bothandler"
	"github.com/igoracmelo/euperturbot/config"
	"github.com/igoracmelo/euperturbot/controller"
	"github.com/igoracmelo/euperturbot/openai"
	"github.com/igoracmelo/euperturbot/repo/sqliterepo"
	_ "modernc.org/sqlite"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error

	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	repo, err := sqliterepo.Open(context.TODO(), "euperturbot.db", "./repo/sqliterepo/migrations")
	if err != nil {
		panic(err)
	}
	defer repo.Close()

	oai := openai.NewService(conf.OpenAIKey, http.DefaultClient)
	myBot := bot.NewService(conf.BotToken)

	botInfo, err := myBot.GetMe()
	if err != nil {
		panic(err)
	}

	c := controller.Controller{
		Repo:    repo,
		OpenAI:  oai,
		BotInfo: botInfo,
		Config:  &conf,
	}

	updates := myBot.GetUpdatesChannel()
	uh := bh.NewUpdateHandler(myBot, updates)

	uh.Middleware(c.EnsureStarted(), bh.AnyMessage)
	uh.Middleware(c.IgnoreForwardedCommand(), bh.AnyCommand)

	uh.Handle(bh.Command("start"), c.RequireAdmin(c.Start))

	uh.Handle(bh.Command("suba"), func(s bot.Service, u bot.Update) error {
		return subscribeToTopic(context.TODO(), repo.DB(), u)
	})

	uh.Handle(bh.Command("desca"), c.UnsubTopic)
	uh.Handle(bh.Command("pollo"), c.CreatePoll)
	uh.Handle(bh.Command("bora"), c.CallSubs)
	uh.Handle(bh.Command("quem"), c.ListSubs)
	uh.Handle(bh.Command("lista"), c.ListUserTopics)
	uh.Handle(bh.Command("listudo"), c.ListChatTopics)
	// c.Handle(tgh.Command("conta"), h.CountEvent)
	// c.Handle(tgh.Command("desconta"), h.UncountEvent)
	uh.Handle(bh.Command("a"), c.SaveAudio)
	uh.Handle(bh.Command("arand"), c.SendRandomAudio)
	uh.Handle(bh.Command("ask"), c.GPTCompletion)
	uh.Handle(bh.Command("cask"), c.GPTChatCompletion)
	uh.Handle(bh.Command("backup"), c.RequireGod(c.Backup))
	uh.Handle(bh.Command("xonotic"), c.Xonotic)
	uh.Handle(bh.AnyCallbackQuery, c.CallbackQuery)
	uh.Handle(bh.AnyInlineQuery, c.InlineQuery)

	// switches
	uh.Handle(bh.Command("enable_create_topics"), c.RequireAdmin(c.Enable("create_topics")))
	uh.Handle(bh.Command("disable_create_topics"), c.RequireAdmin(c.Disable("create_topics")))
	uh.Handle(bh.Command("enable_audio"), c.RequireAdmin(c.Enable("audio")))
	uh.Handle(bh.Command("disable_audio"), c.RequireAdmin(c.Disable("audio")))
	uh.Handle(bh.Command("enable_ask"), c.RequireAdmin(c.Enable("ask")))
	uh.Handle(bh.Command("disable_ask"), c.RequireAdmin(c.Disable("ask")))
	uh.Handle(bh.Command("enable_cask"), c.RequireAdmin(c.Enable("cask")))
	uh.Handle(bh.Command("disable_cask"), c.RequireAdmin(c.Disable("cask")))
	uh.Handle(bh.Command("enable_sed"), c.RequireAdmin(c.Enable("sed")))
	uh.Handle(bh.Command("disable_sed"), c.RequireAdmin(c.Disable("sed")))

	// TODO: text containing #topic
	uh.Handle(bh.AnyText, func(s bot.Service, u bot.Update) error {
		if regexp.MustCompile(`^#[a-z0-9_]{1,}$`).MatchString(u.Message.Text) {
			return callSubscribers(context.TODO(), repo.DB(), u, u.Message.Text)
		}
		return nil
	})

	uh.Start()
}
