package main

import (
	"expense-tracker/internal/database"
	"expense-tracker/internal/expense"
	"expense-tracker/internal/telegram"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	// Loading .env
	err := godotenv.Load()

	if err != nil {
		log.Println("WARNING : .env  file not FOUND!")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	if token == "" {
		log.Println("WARNING : TELEGRAM_BOT_TOKEN Not set yet !")
	}

	// Connecting to the database
	db := database.Connect()
	expenseService := expense.NewService(db)

	// Create telegram Bot
	bot, err := telegram.NewBot(token, expenseService)
	if err != nil {
		log.Fatal(err)
	}

	bot.Start()

}
