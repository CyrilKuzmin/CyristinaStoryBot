package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func ReadFables() string {
	fable, _ := ioutil.ReadFile("fables/story_1.json")
	var data interface{}
	err := json.Unmarshal(fable, &data)
	if err != nil {
		log.Panic(err)
	}
	log.Println(data)
	return fmt.Sprintln(data)
}

func main() {
	var fable = ReadFables()

	bot, err := tgbotapi.NewBotAPI("1781364855:AAGJJqx0pjhCWG_GqpPTuQgZKmZIhwc9Yh4")
	if err != nil {
		log.Panic(err)
	}
	log.Println("Bot started :" + bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fable)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
