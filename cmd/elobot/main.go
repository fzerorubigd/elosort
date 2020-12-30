package main

import (
	"log"
	"os"

	"github.com/fzerorubigd/clictx"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ogier/pflag"

	"github.com/fzerorubigd/elosort/pkg/db/sqlite"
	"github.com/fzerorubigd/elosort/pkg/elobot"
	"github.com/fzerorubigd/elosort/pkg/ranking/elo"
	"github.com/fzerorubigd/elosort/pkg/telegram"
)

func main() {
	ctx := clictx.DefaultContext()

	var (
		token  string
		dbRoot string
		debug  bool
	)
	pflag.StringVar(&token, "token", "", "Telegram bot token, if it's empty, it tries the env TELEGRAM_BOT_TOKEN")
	pflag.StringVar(&dbRoot, "db", "database.sqlite3", "Database path to load")
	pflag.BoolVar(&debug, "debug", false, "Show debug log")

	pflag.Parse()
	if token == "" {
		token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	s, err := sqlite.NewSQLiteStorage(ctx, dbRoot)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = s.Close()
	}()

	ranker := elo.NewEloRankerDefault()
	telegram.InitLibrary(func(cahtID int64) telegram.Menu {
		return elobot.NewChat(cahtID, ranker, s)
	})

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case update := <-updates:
			msg, err := telegram.Update(ctx, update)
			if err != nil {
				log.Println(err)
			}

			for i := range msg {
				if _, err := bot.Send(msg[i]); err != nil {
					log.Println(err)
				}
			}
		case <-ctx.Done():
			log.Print("Done")
			return
		}
	}
}
