package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/go-co-op/gocron/v2"
	 "github.com/joho/godotenv"
)

func notify(bot *gotgbot.Bot, message string, chatId int64) {
	_, err := bot.SendMessage(chatId, message, &gotgbot.SendMessageOpts{})
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("Reminder was Fired Successfully for:" + string(chatId))
}

var scheduler gocron.Scheduler

func main() {
	token:=os.Getenv("TOKEN")
	godotenv.Load()
	scheduler, _ = gocron.NewScheduler()
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
	dispatcher.AddHandler(handlers.NewCommand("start", start))
	dispatcher.AddHandler(handlers.NewCommand("setReminder", setReminder))

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
	scheduler.Start()
	slog.Info("Scheduler Started...")

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
	
	

}

func start(bot *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("Hello, I'm @%s.\nI'm here to help you remind your tasks\nyou can use the following format to set reminders:\n/setReminder \nmessage=message\ntime=time\nwhen=date\nExamples:\n/setReminder\nmessage= don't forget your scrum\ntime=12:00AM\nwhen=weekdays ", bot.User.Username), &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func setReminder(bot *gotgbot.Bot, ctx *ext.Context) error {
	var t time.Time
	str := ctx.EffectiveMessage.Text
	slog.Info("Requesting to Set a Reminder")
	chatId := ctx.EffectiveChat.Id
	str = strings.TrimPrefix(str, "/setReminder")
	list := strings.Split(str, "\n")
	reminder := make(map[string]string)
	for _, part := range list {
		part=strings.TrimSpace(part)
		if strings.Contains(part, "=") {
			keyValue := strings.Split(part, "=")
			if keyValue[0] == "time" {
				keyValue[1]=strings.ToUpper(keyValue[1])
				t, _ = time.Parse("3:04PM",keyValue[1])
				t.Format("15:04")
				continue
			}
			reminder[keyValue[0]] = keyValue[1]
		}
	}
	reminder["when"] = strings.ToLower(reminder["when"])
	slog.Info("Data parsed successfully")

	if reminder["when"] == "weekdays" {
		_, err := scheduler.NewJob(gocron.WeeklyJob(1,
			gocron.NewWeekdays(time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday),
			gocron.NewAtTimes(
				gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second()))),
		), gocron.NewTask(func() {
			notify(bot, reminder["message"], chatId)
		}))
		if err != nil {
			slog.Error(err.Error())
			_, err := ctx.EffectiveMessage.Reply(bot, "Couldn't set the reminder", &gotgbot.SendMessageOpts{})
			if err != nil {
				slog.Error(err.Error())
			}
			return err

		}
		slog.Info(" Reminder was set successfully")
	} else if reminder["when"] == "everyday" {
		_, err := scheduler.NewJob(gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second())))), gocron.NewTask(
			func() {
				notify(bot, reminder["message"], chatId)
			},
		))
		if err != nil {
			slog.Error(err.Error())
			_, err := ctx.EffectiveMessage.Reply(bot, "Couldn't set the reminder", &gotgbot.SendMessageOpts{})
			if err != nil {
				slog.Error(err.Error())
			}
			return err
		}
		slog.Info("Reminder was set successfully")
	}
	_, err := ctx.EffectiveMessage.Reply(bot, "Reminder was set successfully", &gotgbot.SendMessageOpts{})
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}
