package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ParsedExpense struct {
	Amount      float64
	Description string
	ExpenseDate time.Time
}

func ParseExpense(input string) (*ParsedExpense, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return nil, fmt.Errorf("expense cannot be empty")
	}

	parts := strings.Fields(input)

	if len(parts) < 2 {
		return nil, fmt.Errorf(
			"invalid format: amount and description are required",
		)
	}

	// First value must be amount
	amount, err := strconv.ParseFloat(parts[0], 64)

	if err != nil {
		return nil, fmt.Errorf(
			"invalid amount: %s",
			parts[0],
		)
	}

	if amount <= 0 {
		return nil, fmt.Errorf(
			"amount must be greater than zero",
		)
	}

	// Default date = today
	expenseDate := time.Now()

	descriptionParts := parts[1:]

	// Check whether last field is a date
	if isDate(parts[len(parts)-1]) {

		parsedDate, err := parseDate(parts[len(parts)-1])

		if err != nil {
			return nil, err
		}

		expenseDate = parsedDate

		descriptionParts = parts[1 : len(parts)-1]
	}

	if len(descriptionParts) == 0 {
		return nil, fmt.Errorf(
			"description is required",
		)
	}

	description := strings.Join(descriptionParts, " ")

	return &ParsedExpense{
		Amount:      amount,
		Description: description,
		ExpenseDate: expenseDate,
	}, nil
}

func isDate(value string) bool {
	_, err := time.ParseInLocation(
		"02/01/2006",
		value,
		time.Local,
	)

	return err == nil
}

func parseDate(value string) (time.Time, error) {

	date, err := time.ParseInLocation(
		"02/01/2006",
		value,
		time.Local,
	)

	if err != nil {
		return time.Time{}, fmt.Errorf(
			"invalid date %q. Use DD/MM/YYYY",
			value,
		)
	}

	return date, nil
}
