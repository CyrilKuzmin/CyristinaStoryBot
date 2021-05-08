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

var conf *config.Config
var currentChats = make(map[int64]*NextMessage)
var dbClient *story.StoryMongoClient
var storyBot bot.StoryBot
var cacheTimeout = time.Minute * 15

func init() {
	var err error

	conf, err = config.Init("config/main")
	if err != nil {
		log.Panicf("Cannot read config: %v", err)
	}

	dbClient, err = story.NewClient(
		conf.Mongo.URI,
		conf.Mongo.Database,
		conf.Mongo.Collection)
	if err != nil {
		log.Panicf("Cannot connect to Mongo: %v", err)
	}
	log.Printf("Connected to MongoDB : %s", dbClient.DB)

	err = dbClient.GetAllTitles() // первоначальная загрузка тайтлов
	if err != nil {
		log.Panicf("Cannot get titles: %v", err)
	}

	storyBot, err = bot.NewStoryBot(conf.TG.Token, conf.TG.ImagesDir)
	if err != nil {
		log.Panicf("Cannot connect to Teegram: %v", err)
	}

	go cleanup()
	go updateTitles(dbClient)

}

func Run() {
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
				delete(currentChats, chatID)
				// Шлем нужную часть клавиатуры
				titles := getTitlesForKB(dbClient, chatID)
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
					titles := getTitlesForKB(dbClient, chatID)
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
			delete(currentChats, chatID) // на всякий случай
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
				titles := getTitlesForKB(dbClient, chatID)
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

func cleanup() {
	log.Print("Cleanuping...")
	for {
		time.Sleep(time.Minute)
		for chat, msg := range currentChats {
			if msg.activeUntil.Before(time.Now()) {
				delete(currentChats, chat)
			}
		}
	}
}

func updateTitles(smc *story.StoryMongoClient) {
	log.Print("Updating titles cache")
	for {
		err := smc.GetAllTitles()
		if err != nil {
			log.Printf("Error while titles update: %v", err)
		}
		log.Printf("Total titles: %v\n", smc.TitlesCount)
		time.Sleep(time.Minute * 30)
	}
}

func getTitlesForKB(smc *story.StoryMongoClient, chatID int64) (titles []string) {
	next, exist := currentChats[chatID]

	if !exist {
		titles = smc.GetTitlesPart(0)
		createCurrentChatCache(chatID, "TITLES_KEYBOARD")
	} else {
		if next.part > smc.TitlesCount/10 {
			next.part = 0
		}
		titles = smc.GetTitlesPart(next.part)
		updateCurrentChatCache(chatID)
	}
	log.Print(titles)
	return titles
}

func createCurrentChatCache(chatID int64, title string) {
	currentChats[chatID] = &NextMessage{title, 1, time.Now().Add(cacheTimeout)}
}

func updateCurrentChatCache(chatID int64) {
	currentChats[chatID].part++
	currentChats[chatID].activeUntil = time.Now().Add(cacheTimeout)
}
