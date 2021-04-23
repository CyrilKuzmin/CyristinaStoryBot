package story

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Story struct {
	ID     int64
	Title  string
	Text   string
	IsLong bool
	Images []string
}

func ReadAllStories() []Story {
	var stories []Story
	var dirName = "./stories"
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Panic(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			st := ReadStoriesFromFile(fmt.Sprintf("%s/%s", dirName, f.Name()))
			stories = append(stories, st)
		}
	}
	return stories
}

func ReadStoriesFromFile(path string) Story {
	st, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
	}
	var data Story
	err = json.Unmarshal(st, &data)
	if err != nil {
		log.Println(err)
	}
	return data
}

func GenerateMessageForStory(chatId int64, st Story) tgbotapi.MessageConfig {
	text := fmt.Sprintf("*%s*\n\n%s", st.Title, st.Text)
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	return msg
}

func GetTitles(stories []Story) []string {
	var titles []string
	for _, s := range stories {
		titles = append(titles, s.Title)
	}
	return titles
}
