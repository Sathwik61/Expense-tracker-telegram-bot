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

	// handle bulk messages
	// Handle messages from users currently in bulk mode.
	if b.isBulkMode(message.From.ID) {

		// Allow /done and /cancel while in bulk mode.
		if message.Text == "/done" {
			b.handleBulkDone(message)
			return
		}

		if message.Text == "/cancel" {
			b.handleBulkCancel(message)
			return
		}

		b.handleBulkInput(message)
		return
	}

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

	text := `📦 <b>Bulk Expense Mode</b>

Enter multiple expenses separated by commas.

<b>Example:</b>

<pre>
340 chicken biryani 04/10/2026, 120 auto 03/10/2026, 500 groceries, 80 coffee 01/10/2026
</pre>

Each expense must contain:

<b>amount description [date]</b>

📅 Date format: DD/MM/YYYY

When you're finished:

/done

To cancel:

/cancel`

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		text,
	)

	msg.ParseMode = "HTML"

	_, err := b.client.Send(msg)

	if err != nil {
		log.Println("failed to send bulk help:", err)
	}
}
func (b *Bot) handleAdd(message *tgbotapi.Message) {

	input := strings.TrimSpace(
		strings.TrimPrefix(message.Text, "/add"),
	)

	if strings.EqualFold(input, "bulk") {
		b.startBulkMode(message)
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

	const maxMessageLength = 3500

	var total float64

	// Calculate total first.
	for _, expense := range expenses {
		total += expense.Amount
	}

	// Create table header.
	createHeader := func() string {

		var header strings.Builder

		header.WriteString("📊 <b>Your Expenses</b>\n\n")
		header.WriteString("<pre>")

		header.WriteString(
			fmt.Sprintf(
				"%-4s %-10s %9s  %-12s %-25s\n",
				"ID",
				"Date",
				"Amount",
				"Category",
				"Description",
			),
		)

		header.WriteString(
			"-------------------------------------------------------------\n",
		)

		return header.String()
	}

	var messages []string

	currentMessage := createHeader()

	for _, expense := range expenses {

		description := truncateDescription(
			expense.Description,
			25,
		)

		category := truncateDescription(
			expense.Category,
			12,
		)

		row := fmt.Sprintf(
			"%-4d %-10s %9s  %-12s %-25s\n",
			expense.ID,
			expense.ExpenseDate.Format("02/01/2006"),
			fmt.Sprintf("₹%.2f", expense.Amount),
			category,
			description,
		)

		// Check if adding this row would make
		// the Telegram message too large.
		if len(currentMessage)+len(row)+50 > maxMessageLength {

			currentMessage += "</pre>"

			messages = append(
				messages,
				currentMessage,
			)

			currentMessage = createHeader()
		}

		currentMessage += row
	}

	// Add final table footer.
	currentMessage +=
		"-------------------------------------------------------------\n"

	currentMessage += fmt.Sprintf(
		"%44s\n",
		fmt.Sprintf("TOTAL: ₹%.2f", total),
	)

	currentMessage += fmt.Sprintf(
		"%44s",
		fmt.Sprintf("%d expense(s)", len(expenses)),
	)

	currentMessage += "</pre>"

	messages = append(
		messages,
		currentMessage,
	)

	// Send messages one by one.
	for i, text := range messages {

		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			text,
		)

		msg.ParseMode = "HTML"

		_, err := b.client.Send(msg)

		if err != nil {
			log.Printf(
				"failed to send expense list part %d/%d: %v",
				i+1,
				len(messages),
				err,
			)

			return
		}
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

func (b *Bot) startBulkMode(message *tgbotapi.Message) {

	userID := message.From.ID

	b.bulkMutex.Lock()

	b.bulkUsers[userID] = []string{}

	b.bulkMutex.Unlock()

	text := `📦 <b>Bulk Expense Mode</b>

Send one expense per line.

Example:

<pre>
340 chicken biryani 04/10/2026
120 auto 03/10/2026
250 groceries 03/10/2026
</pre>

When you're finished, send:

/done

To cancel:

/cancel`

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		text,
	)

	msg.ParseMode = "HTML"

	_, err := b.client.Send(msg)

	if err != nil {
		log.Println("failed to send bulk help:", err)
	}
}

func (b *Bot) isBulkMode(userID int64) bool {

	b.bulkMutex.Lock()
	defer b.bulkMutex.Unlock()

	_, exists := b.bulkUsers[userID]

	return exists
}
func (b *Bot) handleBulkInput(message *tgbotapi.Message) {

	input := strings.TrimSpace(message.Text)

	if input == "" {
		b.sendMessage(
			message.Chat.ID,
			"❌ Please enter at least one expense.",
		)
		return
	}

	if strings.HasPrefix(input, "/") {
		b.sendMessage(
			message.Chat.ID,
			"❌ You are currently in bulk mode.\n\n"+
				"Use /done to save or /cancel to cancel.",
		)
		return
	}

	// Split expenses by newline
	lines := strings.Split(input, "\n")

	added := 0

	b.bulkMutex.Lock()

	for _, line := range lines {

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		b.bulkUsers[message.From.ID] =
			append(
				b.bulkUsers[message.From.ID],
				line,
			)

		added++
	}

	total := len(b.bulkUsers[message.From.ID])

	b.bulkMutex.Unlock()

	if added == 0 {
		b.sendMessage(
			message.Chat.ID,
			"❌ No expenses found.",
		)
		return
	}

	b.sendMessage(
		message.Chat.ID,
		fmt.Sprintf(
			"✅ Added %d expense(s).\n\n"+
				"📦 Total pending: %d\n\n"+
				"Send more or /done to save.",
			added,
			total,
		),
	)
}
func (b *Bot) handleBulkCancel(message *tgbotapi.Message) {

	userID := message.From.ID

	b.bulkMutex.Lock()

	delete(b.bulkUsers, userID)

	b.bulkMutex.Unlock()

	b.sendMessage(
		message.Chat.ID,
		"❌ Bulk entry cancelled.\n\n"+
			"Nothing was saved.",
	)
}

func (b *Bot) handleBulkDone(message *tgbotapi.Message) {

	userID := message.From.ID

	b.bulkMutex.Lock()

	lines, exists := b.bulkUsers[userID]

	// Remove bulk mode immediately.
	delete(b.bulkUsers, userID)

	b.bulkMutex.Unlock()

	if !exists || len(lines) == 0 {

		b.sendMessage(
			message.Chat.ID,
			"📭 No expenses were entered.",
		)

		return
	}

	// Get/create application user.
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

	// First validate EVERYTHING.
	parsedExpenses := make([]*parser.ParsedExpense, 0, len(lines))

	for i, line := range lines {

		parsed, err := parser.ParseExpense(line)

		if err != nil {

			b.sendMessage(
				message.Chat.ID,
				fmt.Sprintf(
					"❌ Bulk entry failed.\n\n"+
						"Line %d:\n"+
						"%s\n\n"+
						"Error: %s\n\n"+
						"Nothing was saved.",
					i+1,
					line,
					err.Error(),
				),
			)

			return
		}

		parsedExpenses = append(
			parsedExpenses,
			parsed,
		)
	}

	// Save all validated expenses.
	var total float64

	for _, parsed := range parsedExpenses {

		_, err := b.expenseService.CreateExpense(
			user.ID,
			parsed.Amount,
			"OTHER",
			parsed.Description,
			parsed.ExpenseDate,
		)

		if err != nil {

			log.Println("failed to save bulk expense:", err)

			b.sendMessage(
				message.Chat.ID,
				"❌ Failed to save bulk expenses.",
			)

			return
		}

		total += parsed.Amount
	}

	// Build result.
	var text strings.Builder

	text.WriteString("✅ <b>Bulk Expenses Added</b>\n\n")

	for i, parsed := range parsedExpenses {

		text.WriteString(
			fmt.Sprintf(
				"%d. ₹%.2f — %s\n",
				i+1,
				parsed.Amount,
				html.EscapeString(parsed.Description),
			),
		)
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")

	text.WriteString(
		fmt.Sprintf(
			"💰 <b>Total: ₹%.2f</b>\n",
			total,
		),
	)

	text.WriteString(
		fmt.Sprintf(
			"📌 %d expense(s) added",
			len(parsedExpenses),
		),
	)

	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		text.String(),
	)

	msg.ParseMode = "HTML"

	_, err = b.client.Send(msg)

	if err != nil {
		log.Println("failed to send bulk result:", err)
	}
}
