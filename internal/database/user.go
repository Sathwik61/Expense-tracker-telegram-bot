package database

import "time"

type User struct {
	ID         uint  `gorm:"primarykey"`
	TelegramID int64 `gorm:"uniqueIndex;not null"`
	Username   string
	FirstName  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
