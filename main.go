package main

import (
	"log"
	"os"

	// Alias for local directory
	bot "github.com/royarchl/D2-server-bot/bot"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file could not be loaded")
	}

	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.Run()
}
