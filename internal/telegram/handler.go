package telegram

import (
	"expense-tracker/internal/parser"
	"fmt"
	"html"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) {

	// Ignore updates that don't contain a message
	if update.Message == nil {
		return
	}

	message := update.Message

	log.Printf(
		"[%s] %s",
		message.From.UserName,
		message.Text,
	)

	switch {
	case message.Text == "/start":
		b.handleStart(message)

	case message.Text == "/help":
		b.handleHelp(message)

	case message.Text == "/add":
		b.handleAddHelp(message)

	case strings.HasPrefix(message.Text, "/add "):
		b.handleAdd(message)

	case message.Text == "/list":
		b.handleList(message)

	case strings.HasPrefix(message.Text, "/list "):
		b.handleList(message)

	case message.Text == "/delete":
		b.handleDeleteHelp(message)

	case strings.HasPrefix(message.Text, "/delete "):
		b.handleDelete(message)
	default:
		b.handleUnknown(message)
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) {

	user, err := b.expenseService.GetOrCreateUser(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
	)

	if err != nil {
		log.Println("failed to create user:", err)

		b.sendMessage(
			message.Chat.ID,
			"❌ Something went wrong.",
		)

		return
	}

	text := fmt.Sprintf(
		"👋 Welcome %s!\n\nYour account is ready.\n\nYour user ID is %d.",
		user.FirstName,
		user.ID,
	)

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {

	text := `📖 Available Commands

/start - Start the bot
/help - Show help

Expense commands will be added soon.`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleUnknown(message *tgbotapi.Message) {

	b.sendMessage(
		message.Chat.ID,
		"❓ I don't understand that command yet.\n\nUse /help to see available commands.",
	)
}

func (b *Bot) sendMessage(chatID int64, text string) {

	msg := tgbotapi.NewMessage(chatID, text)

	_, err := b.client.Send(msg)

	if err != nil {
		log.Printf("failed to send telegram message: %v", err)
	}
}

func (b *Bot) handleAddHelp(message *tgbotapi.Message) {

	text := `➕ Add Expense

Single Expense:

/add <amount> <description> [DD/MM/YYYY]

Examples:

/add 340 chicken biryani
/add 120 auto 03/10/2026
/add 500 groceries

If you don't provide a date, today's date will be used.

Bulk Expenses:

/add bulk

Then enter one expense per line:

340 chicken biryani
120 auto 03/10/2026
500 groceries
80 coffee 01/10/2026

Date format: DD/MM/YYYY`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleBulkHelp(message *tgbotapi.Message) {

	text := `📦 Bulk Expense Mode

Send your expenses one per line.

Example:

340 chicken biryani
120 auto 03/10/2026
500 groceries
80 coffee 01/10/2026

📅 Date format: DD/MM/YYYY

When you're finished, send:
/done

All expenses will be added together.`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleAdd(message *tgbotapi.Message) {

	input := strings.TrimSpace(
		strings.TrimPrefix(message.Text, "/add"),
	)

	if strings.EqualFold(input, "bulk") {
		b.handleBulkHelp(message)
		return
	}

	parsed, err := parser.ParseExpense(input)

	if err != nil {
		b.sendMessage(
			message.Chat.ID,
			"❌ "+err.Error(),
		)
		return
	}

	user, err := b.expenseService.GetOrCreateUser(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
	)

	if err != nil {
		log.Println(err)

		b.sendMessage(
			message.Chat.ID,
			"❌ Failed to identify your account.",
		)

		return
	}

	expense, err := b.expenseService.CreateExpense(
		user.ID,
		parsed.Amount,
		"OTHER",
		parsed.Description,
		parsed.ExpenseDate,
	)

	if err != nil {
		log.Println(err)

		b.sendMessage(
			message.Chat.ID,
			"❌ Failed to save expense.",
		)

		return
	}

	text := fmt.Sprintf(
		"✅ Expense Added\n\n"+
			"💰 ₹%.2f\n"+
			"📝 %s\n"+
			"📅 %s",
		expense.Amount,
		expense.Description,
		expense.ExpenseDate.Format("02/01/2006"),
	)

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}
func (b *Bot) handleList(message *tgbotapi.Message) {

	user, err := b.expenseService.GetOrCreateUser(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
	)

	if err != nil {
		log.Println("failed to identify user:", err)

		b.sendMessage(
			message.Chat.ID,
			"❌ Failed to identify your account.",
		)

		return
	}

	input := strings.TrimSpace(
		strings.TrimPrefix(message.Text, "/list"),
	)

	filter, err := parser.ParseExpenseFilter(
		input,
		user.ID,
	)

	if err != nil {
		b.sendMessage(
			message.Chat.ID,
			"❌ "+err.Error(),
		)

		return
	}

	expenses, err := b.expenseService.GetExpenses(filter)

	if err != nil {
		log.Println("failed to get expenses:", err)

		b.sendMessage(
			message.Chat.ID,
			"❌ Failed to retrieve expenses.",
		)

		return
	}

	if len(expenses) == 0 {
		b.sendMessage(
			message.Chat.ID,
			"📭 No expenses found.",
		)

		return
	}

	var text strings.Builder

	text.WriteString("📊 <b>Your Expenses</b>\n\n")

	var total float64

	for _, expense := range expenses {

		total += expense.Amount

		description := html.EscapeString(
			truncateDescription(expense.Description, 35),
		)

		text.WriteString(
			fmt.Sprintf(
				"📅 <b>%s</b>\n"+
					"   💰 ₹%.2f  •  %s\n\n",
				expense.ExpenseDate.Format("02/01/2006"),
				expense.Amount,
				description,
			),
		)
	}

	text.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	text.WriteString(
		fmt.Sprintf(
			"💵 <b>Total: ₹%.2f</b>\n",
			total,
		),
	)

	text.WriteString(
		fmt.Sprintf(
			"📌 %d expense(s)",
			len(expenses),
		),
	)

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		text.String(),
	)

	msg.ParseMode = "HTML"

	_, err = b.client.Send(msg)

	if err != nil {
		log.Println("failed to send expense list:", err)
	}
}

func truncateDescription(
	description string,
	maxLength int,
) string {

	if len(description) <= maxLength {
		return description
	}

	return description[:maxLength-3] + "..."
}

func (b *Bot) handleDeleteHelp(message *tgbotapi.Message) {

	text := `🗑️ Delete Expenses

Delete all your expenses:

/delete -a

Delete expenses by category:

/delete <category>

Examples:

/delete food
/delete travel
/delete shopping

⚠️ Deleting expenses cannot be undone.`

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}

func (b *Bot) handleDelete(message *tgbotapi.Message) {

	input := strings.TrimSpace(
		strings.TrimPrefix(message.Text, "/delete"),
	)

	if input == "" {
		b.handleDeleteHelp(message)
		return
	}

	user, err := b.expenseService.GetOrCreateUser(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
	)

	if err != nil {
		log.Println("failed to identify user:", err)

		b.sendMessage(
			message.Chat.ID,
			"❌ Failed to identify your account.",
		)

		return
	}

	// Delete all expenses
	if strings.EqualFold(input, "-a") {

		count, err := b.expenseService.DeleteAllExpenses(user.ID)

		if err != nil {
			log.Println("failed to delete expenses:", err)

			b.sendMessage(
				message.Chat.ID,
				"❌ Failed to delete expenses.",
			)

			return
		}

		if count == 0 {
			b.sendMessage(
				message.Chat.ID,
				"📭 You don't have any expenses to delete.",
			)

			return
		}

		text := fmt.Sprintf(
			"🗑️ <b>Expenses Deleted</b>\n\n"+
				"%d expense(s) deleted successfully.",
			count,
		)

		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			text,
		)

		msg.ParseMode = "HTML"

		_, err = b.client.Send(msg)

		if err != nil {
			log.Println("failed to send delete message:", err)
		}

		return
	}

	// Delete by category
	category := strings.ToUpper(input)

	count, err := b.expenseService.DeleteExpensesByCategory(
		user.ID,
		category,
	)

	if err != nil {
		log.Println("failed to delete expenses by category:", err)

		b.sendMessage(
			message.Chat.ID,
			"❌ Failed to delete expenses.",
		)

		return
	}

	if count == 0 {
		b.sendMessage(
			message.Chat.ID,
			fmt.Sprintf(
				"📭 No expenses found in category <b>%s</b>.",
				html.EscapeString(category),
			),
		)

		return
	}

	text := fmt.Sprintf(
		"🗑️ <b>Expenses Deleted</b>\n\n"+
			"Category: <b>%s</b>\n"+
			"Deleted: %d expense(s)",
		html.EscapeString(category),
		count,
	)

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		text,
	)

	msg.ParseMode = "HTML"

	_, err = b.client.Send(msg)

	if err != nil {
		log.Println("failed to send delete message:", err)
	}
}
