package main

import (
	app "github.com/xxlaefxx/CyristinaStoryBot/internal/app"
)

const configFile = "config/app.yml"

func main() {
	app.Run(configFile)
}
