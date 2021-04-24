package story

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var dirName = "./stories"

type Story struct {
	ID      int64
	Title   string
	Content []ContentPart
	IsLong  bool
}

type ContentPart struct {
	Image   string
	Caption string
}

func ReadAllStories() []Story {
	var stories []Story
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Panic(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			st := readStoryFromFile(fmt.Sprintf("%s/%s", dirName, f.Name()))
			stories = append(stories, st)
		}
	}
	return stories
}

func readStoryFromFile(path string) Story {
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

func GenerateMessagesForStory(chatId int64, st Story) []tgbotapi.PhotoConfig {
	var messages []tgbotapi.PhotoConfig
	for _, cp := range st.Content {
		file := fmt.Sprintf("%s/%s", dirName, cp.Image)
		msg := tgbotapi.NewPhotoUpload(chatId, file)
		msg.ParseMode = "markdown"
		msg.Caption = cp.Caption
		messages = append(messages, msg)
	}
	return messages
}

func GetTitles(stories []Story) []string {
	var titles []string
	for _, s := range stories {
		titles = append(titles, s.Title)
	}
	return titles
}
