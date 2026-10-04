package database

import "time"

type Expense struct {
	ID          uint    `gorm:"primaryKey"`
	UserID      uint    `gorm:"not null;index"`
	Amount      float64 `gorm:"not null"`
	Category    string  `gorm:"not null"`
	Description string
	ExpenseDate time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
