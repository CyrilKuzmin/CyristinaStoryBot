package app

import (
	"log"
	"time"

	story "github.com/xxlaefxx/CyristinaStoryBot/internal/story"
)

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
		time.Sleep(time.Minute)
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
	currentChats[chatID] = &NextMessage{title, 1, time.Now().Add(cfg.TG.NextMsgTimeout)}
}

func updateCurrentChatCache(chatID int64) {
	currentChats[chatID].part++
	currentChats[chatID].activeUntil = time.Now().Add(cfg.TG.NextMsgTimeout)
}
