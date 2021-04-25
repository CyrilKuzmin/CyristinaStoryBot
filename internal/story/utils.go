package story

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetAllStories() map[string]Story {
	var allStories map[string]Story = make(map[string]Story)
	var stories = readAllStories()
	var titles = getAllTitles()

	log.Printf("Stories titles: %#v", titles)

	for _, t := range titles {
		for _, s := range stories {
			if t == s.Title {
				allStories[t] = s
			}
		}
	}

	return allStories
}

func GetTitlesKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	titles := getAllTitles()
	for _, title := range titles {
		if len(title) > 60 {
			title = title[0:60] + "..."
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(title, title)
		row := tgbotapi.NewInlineKeyboardRow(btn)
		rows = append(rows, row)
	}

	var storiesKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return storiesKeyboard
}

func readAllStories() []Story {
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
	data.Parts = len(data.Content)
	return data
}

func GenerateMessageForStory(chatId int64, st Story, part int) (tgbotapi.PhotoConfig, error) {
	if part >= len(st.Content) {
		errorFile := fmt.Sprintf("%s/%s", dirName, "error.jpg")
		return tgbotapi.NewPhotoUpload(chatId, errorFile), errors.New("story is over")
	}
	file := fmt.Sprintf("%s/%s", dirName, st.Content[part].Image)
	msg := tgbotapi.NewPhotoUpload(chatId, file)
	msg.ParseMode = "markdown"
	msg.Caption = st.Content[part].Caption
	if part < len(st.Content)-1 {
		msg.ReplyMarkup = singleInlineButton("Продолжить", "NEXT")
	}
	if part == len(st.Content)-1 {
		msg.ReplyMarkup = singleInlineButton("Открыть меню", "OPEN_MENU")
	}
	return msg, nil
}

func getAllTitles() []string {
	var titles []string
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Panic(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			st := readStoryFromFile(fmt.Sprintf("%s/%s", dirName, f.Name()))
			titles = append(titles, st.Title)
		}
	}
	return titles
}

func singleInlineButton(text, data string) tgbotapi.InlineKeyboardMarkup {
	bs := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text, data)}
	kb := tgbotapi.NewInlineKeyboardMarkup(bs)
	return kb
}
