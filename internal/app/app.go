package app

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	bot "github.com/xxlaefxx/CyristinaStoryBot/internal/bot"
	config "github.com/xxlaefxx/CyristinaStoryBot/internal/config"
	story "github.com/xxlaefxx/CyristinaStoryBot/internal/story"
)

type NextMessage struct {
	title       string
	part        int
	activeUntil time.Time
}

var cfg = new(config.Config)
var currentChats = make(map[int64]*NextMessage)

func Run(configFile string) {

	cfg, err := config.Init("config/main")
	if err != nil {
		log.Panicf("Cannot read config: %v", err)
	}

	go cleanup()

	dbClient, err := story.NewClient(
		cfg.Mongo.URI,
		cfg.Mongo.Database,
		cfg.Mongo.Collection)
	if err != nil {
		log.Panicf("Cannot connect to Mongo: %v", err)
	}

	log.Printf("Connected to MongoDB : %s", dbClient.DB)
	dbClient.GetAllTitles() // первоначальная загрузка тайтлов
	go updateTitles(&dbClient)

	storyBot, err := bot.NewStoryBot(cfg.TG.Token, cfg.TG.ImagesDir)
	if err != nil {
		log.Panicf("Cannot connect to Teegram: %v", err)
	}

	log.Printf("Bot started : %s", storyBot.Bot.Self.UserName)

	for update := range storyBot.Updates {
		// Варианты событий. что мы обрабатываем
		if update.CallbackQuery != nil {
			// Обрабатываем нажатия кнопок Inline-клавиатуры
			log.Printf("Callback Query: %s", update.CallbackQuery.Data)
			// Шлем ответ о том, что все ок, обрабатываем-с
			storyBot.Bot.AnswerCallbackQuery(
				tgbotapi.NewCallback(
					update.CallbackQuery.ID,
					update.CallbackQuery.Data))
			chatID := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data
			if data == "OPEN_ALL_TITLES_MENU" {
				// Шлем нужную часть клавиатуры
				titles := getTitlesForKB(&dbClient, chatID)
				if len(titles) > 0 {
					storyBot.SendTitlesMessage(chatID, titles)
				}
				continue
			}

			if data == "NEXT" {
				// Шлем следующую часть чего-то там
				next, alive := currentChats[chatID]
				if !alive {
					// просрочился кэш, шлем занова help
					storyBot.SendHelpMessage(chatID)
					continue
				}
				if next.title == "TITLES_KEYBOARD" {
					// Шлем нужную часть клавиатуры
					titles := getTitlesForKB(&dbClient, chatID)
					if len(titles) > 0 {
						storyBot.SendTitlesMessage(chatID, titles)
					}
					continue
				}
				// Шлем нужную часть сказки
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
				updateCurrentChatCache(chatID)
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
			createCurrentChatCache(chatID, data)
			continue
		} else if update.Message.Command() != "" {
			// Обрабатываем команды типа /start и остальных
			chatID := update.Message.Chat.ID
			log.Printf("Command: %s", update.Message.Command())
			switch update.Message.Command() {
			case "start":
				titles := getTitlesForKB(&dbClient, chatID)
				if len(titles) > 0 {
					storyBot.SendTitlesMessage(chatID, titles)
				}
				continue
			default:
				storyBot.SendHelpMessage(update.Message.Chat.ID)
				continue
			}
		} else if update.Message == nil {
			// Не делаем ничего, если дошли сюда, а сообщения нет.
			// На случай странных callback
			log.Print("Message is nil. Skipping")
			continue
		} else if dbClient.Titles[update.Message.Text] {
			// Клиент написал точное название сказки
			chatID := update.Message.Chat.ID
			cp, err := dbClient.GetStoryPart(update.Message.Text, 0)
			if err != nil {
				log.Printf("Cannot get 1st part of story: %v", err)
				continue
			}
			err = storyBot.SendContentMessage(chatID, cp.Image, cp.Caption)
			if err != nil {
				log.Printf("Cannot send 1st part of story: %v", err)
				continue
			}
		} else {
			// Отправляем хелп, если клиент прислал просто текст
			log.Printf(
				"Just a message: [%s] %s",
				update.Message.From.UserName,
				update.Message.Text)
			storyBot.SendHelpMessage(update.Message.Chat.ID)
		}
	}
}
