package app

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	config "github.com/xxlaefxx/CyristinaStoryBot/internal/config"
	story "github.com/xxlaefxx/CyristinaStoryBot/internal/story"
	storybot "github.com/xxlaefxx/CyristinaStoryBot/internal/storybot"
)

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

type NextMessage struct {
	title       string
	part        int
	activeUntil time.Time
}

func Run(configFile string) {

	cfg, err := config.Init("config/main")
	if err != nil {
		log.Panicf("Cannot read config: %v", err)
	}

	var currentChats = make(map[int64]*NextMessage)

	go cleanup(&currentChats)

	dbClient, err := story.NewClient(
		cfg.Mongo.URI,
		cfg.Mongo.Database,
		cfg.Mongo.Collection)
	if err != nil {
		log.Panicf("Cannot connect to Mongo: %v", err)
	}

	log.Printf("Connected to MongoDB : %s", dbClient.DB)

	storyBot, err := storybot.NewStoryBot(cfg.TG.Token, cfg.TG.ImagesDir)
	if err != nil {
		log.Panicf("Cannot connect to Teegram: %v", err)
	}

	log.Printf("Bot started : %s", storyBot.Bot.Self.UserName)

	for update := range storyBot.Updates {

		if update.CallbackQuery != nil {
			// Обрабатываем нажатия кнопок Inline-клавиатуры
			log.Printf("Callback Query: %s", update.CallbackQuery.Data)
			storyBot.Bot.AnswerCallbackQuery(
				tgbotapi.NewCallback(
					update.CallbackQuery.ID,
					update.CallbackQuery.Data))
			chatID := update.CallbackQuery.Message.Chat.ID
			if update.CallbackQuery.Data == "OPEN_MENU" {
				// Открываем менюшку
				titles, _ := dbClient.GetAllTitles()
				storyBot.SendTitlesMessage(chatID, titles)
				continue
			}

			if update.CallbackQuery.Data == "NEXT" {
				// Шлем нужную часть
				next, alive := currentChats[chatID]
				if !alive {
					storyBot.SendHelpMessage(chatID)
					continue
				}
				cp, err := dbClient.GetStoryPart(next.title, next.part)
				if err != nil {
					continue
				}
				err = storyBot.SendContentMessage(chatID, cp.Image, cp.Caption)
				if err != nil {
					log.Printf(
						"Cannot send %v part of story: %v",
						next.part,
						err)
					continue
				}
				currentChats[chatID].part = currentChats[chatID].part + 1
				currentChats[chatID].activeUntil = time.Now().Add(cfg.TG.NextMsgTimeout)
				continue
			}
			// Клиент нажал кнопку с названием сказки. Шлем первую часть
			cp, err := dbClient.GetStoryPart(update.CallbackQuery.Data, 0)
			if err != nil {
				log.Printf("Cannot get 1st part of story: %v", err)
				continue
			}
			err = storyBot.SendContentMessage(chatID, cp.Image, cp.Caption)
			if err != nil {
				log.Printf("Cannot send 1st part of story: %v", err)
				continue
			}
			currentChats[chatID] = &NextMessage{
				update.CallbackQuery.Data, 1,
				time.Now().Add(cfg.TG.NextMsgTimeout)}
			continue
		}

		if update.Message.Command() != "" {
			// Обрабатываем команды типа /start и остальных
			log.Printf("Command: %s", update.Message.Command())
			switch update.Message.Command() {
			case "start":
				titles, _ := dbClient.GetAllTitles()
				storyBot.SendTitlesMessage(update.Message.Chat.ID, titles)
				continue
			default:
				storyBot.SendHelpMessage(update.Message.Chat.ID)
				continue
			}
		}

		if update.Message == nil {
			// Не делаем ничего, если дошли сюда, а сообщения нет.
			// На случай странных callback
			log.Print("Message is nil. Skipping")
			continue
		}

		// Отправляем хелп, если клиент прислал просто текст
		log.Printf(
			"Just a message: [%s] %s",
			update.Message.From.UserName,
			update.Message.Text)
		storyBot.SendHelpMessage(update.Message.Chat.ID)
	}
}
