package main

import (
	"log"
	"log/slog"
	"os"

	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"notify-bot/pkg"
)




func main() {
	godotenv.Load("../.env")
	token := os.Getenv("TOKEN")
	botHandlers.Scheduler, _ = gocron.NewScheduler()
	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		log.Fatal(err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)
	dispatcher.AddHandler(handlers.NewCommand("start", botHandlers.Start))
	dispatcher.AddHandler(handlers.NewCommand("setReminder", botHandlers.SetReminder))

	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		slog.Error(err.Error())
		return
	}
	log.Printf("%s has been started...\n", bot.Username)
	botHandlers.Scheduler.Start()
	slog.Info("Scheduler Started...")

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()

}


