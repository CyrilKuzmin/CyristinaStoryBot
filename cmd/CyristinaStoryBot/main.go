package main

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	story "github.com/xxlaefxx/CyristinaStoryBot/internal/story"
)

var token = "1781364855:AAGJJqx0pjhCWG_GqpPTuQgZKmZIhwc9Yh4"

const nextMsgTimeout = 15

func cleanup(chats *map[int64]*NextMessage) {
	log.Print("Cleanuping...")
	for {
		time.Sleep(time.Minute)
		for chat, msg := range *chats {
			if msg.activeUntil.Before(time.Now()) {
				delete(*chats, chat)
			}
		}
	}
}

func helpMessage(chatId int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, "Введите /start чтобы получить список сказок")
	return msg
}

func titlesMessage(chatId int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, "Выбирай сказку:")
	msg.ReplyMarkup = story.GetTitlesKeyboard()
	return msg
}

func sendMsg(bot *tgbotapi.BotAPI, c tgbotapi.Chattable) {
	_, err := bot.Send(c)
	if err != nil {
		log.Print(err)
	}
}

type NextMessage struct {
	title       string
	part        int
	activeUntil time.Time
}

func main() {

	allStories := story.GetAllStories()

	var currentChats = make(map[int64]*NextMessage)
	var expiryTime = time.Duration(time.Minute * nextMsgTimeout) // сколько ждем клиента для продолжения

	go cleanup(&currentChats)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Bot started : %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.CallbackQuery != nil {
			// Обрабатываем нажатия кнопок Inline-клавиатуры
			log.Printf("Callback Query: %s", update.CallbackQuery.Data)
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			chatID := update.CallbackQuery.Message.Chat.ID
			if update.CallbackQuery.Data == "OPEN_MENU" {
				// Открываем менюшку
				sendMsg(bot, titlesMessage(update.CallbackQuery.Message.Chat.ID))
				continue
			}

			st, ok := allStories[update.CallbackQuery.Data]
			if !ok {
				// Нажал "Продолжить" ...или что-то иное
				next, alive := currentChats[update.CallbackQuery.Message.Chat.ID]
				if !alive {
					sendMsg(bot, helpMessage(chatID))
					continue
				}
				msg, err := story.GenerateMessageForStory(chatID, allStories[next.title], next.part)
				if err != nil {
					continue
				}
				currentChats[chatID].part = currentChats[chatID].part + 1
				sendMsg(bot, msg)
				continue
			} else {
				// Клиент нажал кнопку с названием сказки. Шлем первую часть
				msg, _ := story.GenerateMessageForStory(chatID, st, 0)
				currentChats[chatID] = &NextMessage{st.Title, 1, time.Now().Add(expiryTime)}
				sendMsg(bot, msg)
				continue
			}
		}

		if update.Message.Command() != "" {
			// Обрабатываем команды типа /start и остальных
			log.Printf("Command: %s", update.Message.Command())
			switch update.Message.Command() {
			case "start":
				sendMsg(bot, titlesMessage(update.Message.Chat.ID))
				continue
			default:
				sendMsg(bot, helpMessage(update.Message.Chat.ID))
				continue
			}
		}

		if update.Message == nil {
			// Не делаем ничего, если дошли сюда, а сообщения нет. На случай странных callback
			log.Print("Message is nil. Skipping")
			continue
		}

		// Отправляем хелп, если клиент прислал просто текст
		log.Printf("Just a message: [%s] %s", update.Message.From.UserName, update.Message.Text)
		sendMsg(bot, helpMessage(update.Message.Chat.ID))

	}
}
