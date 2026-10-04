package expense

import "time"

type ExpenseFilter struct {
	UserID    uint
	StartDate *time.Time
	EndDate   *time.Time
	Category  string

	MinAmount *float64
	MaxAmount *float64
}
