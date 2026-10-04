package telegram

import (
	"expense-tracker/internal/expense"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	client         *tgbotapi.BotAPI
	expenseService *expense.Service
}

func NewBot(token string, expenseService *expense.Service) (*Bot, error) {

	client, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &Bot{
		client:         client,
		expenseService: expenseService,
	}, nil
}

func (b *Bot) Start() {

	fmt.Printf("Telegram bot started: @%s\n", b.client.Self.UserName)

	config := tgbotapi.NewUpdate(0)
	config.Timeout = 30

	updates := b.client.GetUpdatesChan(config)

	for update := range updates {

		b.handleUpdate(update)
	}
}
