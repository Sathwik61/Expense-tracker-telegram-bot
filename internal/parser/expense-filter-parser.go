package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"expense-tracker/internal/expense"
)

func ParseExpenseFilter(
	input string,
	userID uint,
) (expense.ExpenseFilter, error) {

	filter := expense.ExpenseFilter{
		UserID: userID,
	}

	input = strings.TrimSpace(input)

	// /list
	if input == "" {
		return filter, nil
	}

	// /list today
	if strings.EqualFold(input, "today") {
		return todayFilter(filter)
	}

	// /list week
	if strings.EqualFold(input, "week") {
		return weekFilter(filter)
	}

	// /list month
	if strings.EqualFold(input, "month") {
		return monthFilter(filter)
	}

	// /list year
	if strings.EqualFold(input, "year") {
		return yearFilter(filter)
	}

	// /list quarter 3/2026
	if strings.HasPrefix(
		strings.ToLower(input),
		"quarter ",
	) {
		return quarterFilter(filter, input)
	}

	// /list 01/09/2026 - 30/09/2026
	if strings.Contains(input, " - ") {
		return dateRangeFilter(filter, input)
	}

	// /list >500
	if strings.HasPrefix(input, ">") {
		return minAmountFilter(filter, input)
	}

	// /list <500
	if strings.HasPrefix(input, "<") {
		return maxAmountFilter(filter, input)
	}

	parts := strings.Fields(input)

	// /list 09/2026 food
	if len(parts) == 2 && isMonth(parts[0]) {

		monthFilterResult, err := parseMonth(
			filter,
			parts[0],
		)

		if err != nil {
			return filter, err
		}

		monthFilterResult.Category = parts[1]

		return monthFilterResult, nil
	}

	// /list 03/10/2026
	if len(input) == 10 {

		date, err := time.ParseInLocation(
			"02/01/2006",
			input,
			time.Local,
		)

		if err != nil {
			return filter, fmt.Errorf(
				"invalid date. Use DD/MM/YYYY",
			)
		}

		start := date
		end := date.AddDate(0, 0, 1)

		filter.StartDate = &start
		filter.EndDate = &end

		return filter, nil
	}

	// /list 09/2026
	if len(input) == 7 {
		return parseMonth(filter, input)
	}

	// /list 2026
	if len(input) == 4 {

		year, err := strconv.Atoi(input)

		if err != nil {
			return filter, fmt.Errorf(
				"invalid year",
			)
		}

		start := time.Date(
			year,
			time.January,
			1,
			0, 0, 0, 0,
			time.Local,
		)

		end := start.AddDate(1, 0, 0)

		filter.StartDate = &start
		filter.EndDate = &end

		return filter, nil
	}

	return filter, fmt.Errorf(
		"invalid filter. Try:\n" +
			"/list\n" +
			"/list today\n" +
			"/list week\n" +
			"/list month\n" +
			"/list year\n" +
			"/list 03/10/2026\n" +
			"/list 09/2026\n" +
			"/list 2026\n" +
			"/list quarter 3/2026\n" +
			"/list 01/09/2026 - 30/09/2026\n" +
			"/list >500\n" +
			"/list <500\n" +
			"/list 09/2026 food",
	)
}

func todayFilter(
	filter expense.ExpenseFilter,
) (expense.ExpenseFilter, error) {

	now := time.Now()

	start := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		time.Local,
	)

	end := start.AddDate(0, 0, 1)

	filter.StartDate = &start
	filter.EndDate = &end

	return filter, nil
}

func weekFilter(
	filter expense.ExpenseFilter,
) (expense.ExpenseFilter, error) {

	now := time.Now()

	weekday := int(now.Weekday())

	// Convert Sunday=0, Monday=1... into
	// Monday=0, Tuesday=1... Sunday=6
	if weekday == 0 {
		weekday = 6
	} else {
		weekday--
	}

	start := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		time.Local,
	)

	start = start.AddDate(
		0,
		0,
		-weekday,
	)

	end := start.AddDate(0, 0, 7)

	filter.StartDate = &start
	filter.EndDate = &end

	return filter, nil
}

func monthFilter(
	filter expense.ExpenseFilter,
) (expense.ExpenseFilter, error) {

	now := time.Now()

	start := time.Date(
		now.Year(),
		now.Month(),
		1,
		0, 0, 0, 0,
		time.Local,
	)

	end := start.AddDate(0, 1, 0)

	filter.StartDate = &start
	filter.EndDate = &end

	return filter, nil
}

func yearFilter(
	filter expense.ExpenseFilter,
) (expense.ExpenseFilter, error) {

	now := time.Now()

	start := time.Date(
		now.Year(),
		time.January,
		1,
		0, 0, 0, 0,
		time.Local,
	)

	end := start.AddDate(1, 0, 0)

	filter.StartDate = &start
	filter.EndDate = &end

	return filter, nil
}

func quarterFilter(
	filter expense.ExpenseFilter,
	input string,
) (expense.ExpenseFilter, error) {

	value := strings.TrimSpace(
		strings.TrimPrefix(
			strings.ToLower(input),
			"quarter",
		),
	)

	parts := strings.Split(value, "/")

	if len(parts) != 2 {
		return filter, fmt.Errorf(
			"invalid quarter. Use: /list quarter 3/2026",
		)
	}

	quarter, err1 := strconv.Atoi(parts[0])
	year, err2 := strconv.Atoi(parts[1])

	if err1 != nil ||
		err2 != nil ||
		quarter < 1 ||
		quarter > 4 {

		return filter, fmt.Errorf(
			"invalid quarter. Use: /list quarter 3/2026",
		)
	}

	startMonth := time.Month(
		(quarter-1)*3 + 1,
	)

	start := time.Date(
		year,
		startMonth,
		1,
		0, 0, 0, 0,
		time.Local,
	)

	end := start.AddDate(0, 3, 0)

	filter.StartDate = &start
	filter.EndDate = &end

	return filter, nil
}

func dateRangeFilter(
	filter expense.ExpenseFilter,
	input string,
) (expense.ExpenseFilter, error) {

	parts := strings.Split(input, " - ")

	if len(parts) != 2 {
		return filter, fmt.Errorf(
			"invalid date range. Use DD/MM/YYYY - DD/MM/YYYY",
		)
	}

	start, err := time.ParseInLocation(
		"02/01/2006",
		strings.TrimSpace(parts[0]),
		time.Local,
	)

	if err != nil {
		return filter, fmt.Errorf(
			"invalid start date",
		)
	}

	end, err := time.ParseInLocation(
		"02/01/2006",
		strings.TrimSpace(parts[1]),
		time.Local,
	)

	if err != nil {
		return filter, fmt.Errorf(
			"invalid end date",
		)
	}

	// Make end date inclusive.
	end = end.AddDate(0, 0, 1)

	filter.StartDate = &start
	filter.EndDate = &end

	return filter, nil
}

func minAmountFilter(
	filter expense.ExpenseFilter,
	input string,
) (expense.ExpenseFilter, error) {

	value := strings.TrimSpace(
		strings.TrimPrefix(input, ">"),
	)

	amount, err := strconv.ParseFloat(value, 64)

	if err != nil || amount < 0 {
		return filter, fmt.Errorf(
			"invalid amount. Example: /list >500",
		)
	}

	filter.MinAmount = &amount

	return filter, nil
}

func maxAmountFilter(
	filter expense.ExpenseFilter,
	input string,
) (expense.ExpenseFilter, error) {

	value := strings.TrimSpace(
		strings.TrimPrefix(input, "<"),
	)

	amount, err := strconv.ParseFloat(value, 64)

	if err != nil || amount < 0 {
		return filter, fmt.Errorf(
			"invalid amount. Example: /list <500",
		)
	}

	filter.MaxAmount = &amount

	return filter, nil
}

func parseMonth(
	filter expense.ExpenseFilter,
	input string,
) (expense.ExpenseFilter, error) {

	parts := strings.Split(input, "/")

	if len(parts) != 2 {
		return filter, fmt.Errorf(
			"invalid month. Use MM/YYYY",
		)
	}

	month, err1 := strconv.Atoi(parts[0])
	year, err2 := strconv.Atoi(parts[1])

	if err1 != nil ||
		err2 != nil ||
		month < 1 ||
		month > 12 {

		return filter, fmt.Errorf(
			"invalid month. Use MM/YYYY",
		)
	}

	start := time.Date(
		year,
		time.Month(month),
		1,
		0, 0, 0, 0,
		time.Local,
	)

	end := start.AddDate(0, 1, 0)

	filter.StartDate = &start
	filter.EndDate = &end

	return filter, nil
}

func isMonth(value string) bool {

	if len(value) != 7 {
		return false
	}

	parts := strings.Split(value, "/")

	if len(parts) != 2 {
		return false
	}

	month, err1 := strconv.Atoi(parts[0])
	year, err2 := strconv.Atoi(parts[1])

	return err1 == nil &&
		err2 == nil &&
		month >= 1 &&
		month <= 12 &&
		year >= 1
}
