package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	// "honnef.co/go/tools/config"
)

type newTaskCreation struct {
	Stage         string
	UserID        string
	Description   string
	ScheduledTime string
	IsRecurring   bool
	Interval      string
	// CreatedAt     time.Time
}

type userContextStruct struct {
	Action    string
	UserID    int
	IsPremium bool
}

var newUserTask = make(map[int64]*newTaskCreation)

var userContext = make(map[int64]*userContextStruct)

// var err error

func HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, userContext *userContextStruct) {
	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			userContext.Action = "start"
			userContext.UserID = update.Message.From.ID
			userName := update.Message.From.FirstName
			userID := fmt.Sprintf("%d", update.Message.From.ID)

			switch returnCode := CreateUser(config.DBInfo, userName, userID); returnCode {
			case 0:
				SendMessage(bot, update, "Привет! Я помогу вам управлять задачами.")
			case 1:
				SendMessage(bot, update, "И снова здравствуйте, "+userName+"!")
			}
		case "addtask":
			userContext.Action = "newTask"
			NewTask(bot, update, int64(update.Message.From.ID))
			// SendMessage(bot, update, "Задача добавлена!")
		default:
			SendMessage(bot, update, "Извините, я не могу распознать вашу команду.")
		}
	} else {
		switch userContext.Action {
		case "newTask":
			NewTask(bot, update, int64(update.Message.From.ID))
		default:
			SendMessage(bot, update, "Кажется, мы с вами ничего не делаем... Дайте мне команду из списка. :)")
		}
	}
}

func SendMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, messageText string) {
	message := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	// message.ReplyToMessageID = update.Message.MessageID

	bot.Send(message)
}

func NewTask(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64) {
	if _, exists := newUserTask[chatID]; !exists {
		newUserTask[chatID] = &newTaskCreation{Stage: "addDescription"}
	}

	userData := newUserTask[chatID]
	userData.UserID = fmt.Sprintf("%d", update.Message.From.ID)
	fmt.Printf("Current user action stage: %v", newUserTask[chatID].Stage)
	for {
		switch userData.Stage {
		case "addDescription":
			SendMessage(bot, update, "Отлично! Давайте придумаем название задачи.")
			userData.Stage = "askScheduledTime"
			// break
		case "askScheduledTime":
			if !update.Message.IsCommand() {
				userData.Description = update.Message.Text
				SendMessage(bot, update, "Когда вам напомнить о задаче? Например, 2023-05-26 14:00.")
				userData.Stage = "askIsRecurring"
				// break
			} else {
				SendMessage(bot, update, "Мы должны пройти полный цикл добавления задачи, хотите вы этого или нет.")
				// break
			}

		case "askIsRecurring":
			if !update.Message.IsCommand() {
				userData.ScheduledTime = update.Message.Text + ":00"
				_, err := time.Parse(time.DateTime, userData.ScheduledTime)
				if err != nil {
					log.Printf("Failed to parse user task schedule: %v", err)
					SendMessage(bot, update, "Формат даты и времени неправильный. Правильный формат: 2023-05-26 14:00")
					break
				}

				userID := fmt.Sprintf("%d", update.Message.From.ID)
				isPremiumQuery, err := ExecQuery(config.DBInfo, "SELECT * FROM users WHERE telegram_id='"+userID+"' AND is_premium='TRUE'")
				if err != nil {
					log.Fatalf("Failed to query user premium state: %v", err)
				}
				isPremium, err := isPremiumQuery.RowsAffected()
				if err != nil {
					log.Fatalf("Failed to get query rows affected while checking user premium state: %v", err)
				}
				if isPremium > 0 {
					SendMessage(bot, update, "Ура, вы премиальный пользователь!")

					// ОБРАБОТКА ПЕРИОДИЧНОСТИ ЗАДАЧИ

					userContext[int64(update.Message.From.ID)].IsPremium = true
					SendMessage(bot, update, "Нужно ли периодически напоминать вам о задаче?")
					userData.Stage = "howOftenRecurring"
					// break
				} else {
					SendMessage(bot, update, "Увы, вам не положены повторяющиеся сообщения. Идём дальше.")
					userContext[int64(update.Message.From.ID)].IsPremium = false
					userData.Stage = "howOftenRecurring"
					continue
				}
			} else {
				SendMessage(bot, update, "Мы должны пройти полный цикл добавления задачи, хотите вы этого или нет.")
				continue
			}

		case "howOftenRecurring":
			if !update.Message.IsCommand() {
				if userContext[int64(update.Message.From.ID)].IsPremium {
					switch howOftenRecurringAnswer := update.Message.Text; howOftenRecurringAnswer {
					case "+":
						userData.IsRecurring = true
						SendMessage(bot, update, "Отлично! Записываю себе напоминалку напомнить вам о задаче...")
						userData.Stage = "createTask"
						// break
					case "-":
						userData.IsRecurring = false
						SendMessage(bot, update, "Хорошо! Записываю себе напоминалку напомнить вам о задаче...")
						userData.Stage = "createTask"
						// break
					default:
						SendMessage(bot, update, "Давайте будем более краткими :) Да (+) или нет (-)?")
						continue
					}
				} else {
					SendMessage(bot, update, "Увы, вам снова не положены повторяющиеся сообщения. Идём дальше.")
					userData.Stage = "createTask"
					continue
				}
			} else {
				SendMessage(bot, update, "Мы должны пройти полный цикл добавления задачи, хотите вы этого или нет.")
				continue
			}

		case "createTask":
			log.Println("Array consists of: Stage, UserID, Description, ScheduledTime, IsRecurring, Interval")
			log.Printf("Gathered data: %v", userData)
			isTaskCreated := CreateNewTask(config.DBInfo, userData.UserID, userData.Description, userData.ScheduledTime, strconv.FormatBool(userData.IsRecurring), userData.Interval)
			if isTaskCreated < 1 {
				SendMessage(bot, update, "Что-то пошло не так. Кажется, я забыл все данные... :(")
			} else {
				SendMessage(bot, update, "Записываю себе напоминалку напомнить вам о задаче...")
			}
		}
		break
	}

}
