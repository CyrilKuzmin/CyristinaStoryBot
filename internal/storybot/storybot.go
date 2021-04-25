package storybot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type StoryBot struct {
	Bot             *tgbotapi.BotAPI
	Updates         tgbotapi.UpdatesChannel
	ImagesDirectory string
}

func NewStoryBot(token, imagesDir string) (StoryBot, error) {
	var sb StoryBot
	var err error
	sb.ImagesDirectory = imagesDir
	sb.Bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return sb, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	sb.Updates, err = sb.Bot.GetUpdatesChan(u)

	if err != nil {
		return sb, err
	}

	return sb, nil
}

func (sb StoryBot) SendMsg(c tgbotapi.Chattable) error {
	_, err := sb.Bot.Send(c)
	if err != nil {
		return err
	}
	return nil
}

func (sb StoryBot) SendHelpMessage(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, "Введите /start чтобы получить список сказок")
	err := sb.SendMsg(msg)
	return err
}

func (sb StoryBot) SendTitlesMessage(chatId int64, titles []string) error {
	msg := tgbotapi.NewMessage(chatId, "Выбирай сказку:")
	msg.ReplyMarkup = getTitlesKeyboard(titles)
	err := sb.SendMsg(msg)
	return err
}

func (sb StoryBot) SendContentMessage(chatId int64, image, text string) error {
	var err error = nil
	if image != "n/a" {
		file := fmt.Sprintf("%s/%s", sb.ImagesDirectory, image)
		err = sb.SendMsg(getPhotoMessage(chatId, file, text))
	} else {
		err = sb.SendMsg(getTextMessage(chatId, text))
	}
	return err
}

func getPhotoMessage(chatId int64, pathToImage, caption string) tgbotapi.PhotoConfig {
	msg := tgbotapi.NewPhotoUpload(chatId, pathToImage)
	msg.Caption = caption
	msg.ParseMode = "markdown"
	ps := strings.Split(pathToImage, "/")
	file := ps[len(ps)-1]
	if file != "end.jpg" {
		msg.ReplyMarkup = singleInlineButton("Продолжить", "NEXT")
	} else {
		msg.ReplyMarkup = singleInlineButton("Открыть меню", "OPEN_MENU")
	}
	return msg
}

func getTextMessage(chatId int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ReplyMarkup = singleInlineButton("Продолжить", "NEXT")
	return msg
}

func getTitlesKeyboard(titles []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
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

func singleInlineButton(text, data string) tgbotapi.InlineKeyboardMarkup {
	bs := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text, data)}
	kb := tgbotapi.NewInlineKeyboardMarkup(bs)
	return kb
}
