package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var config = LoadConfig()

func main() {
	config := LoadConfig()

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	InitDB(config.DBInfo)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	log.Println("Created updates channel: ", updates)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := int64(update.Message.From.ID)
		if _, exists := newUserTask[chatID]; !exists {
			userContext[chatID] = &userContextStruct{Action: ""}
		}

		HandleCommand(bot, update, userContext[chatID])
		// if update.Message.IsCommand() {
		// 	HandleCommand(bot, update, userContext[chatID])
		// } else {
		// 	SendMessage(bot, update, "Извините, я не могу распознать ваше сообщение как команду.")
		// }
	}
}
