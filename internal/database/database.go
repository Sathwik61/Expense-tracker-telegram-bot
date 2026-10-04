package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	db, err := gorm.Open(
		sqlite.Open("data/expense.db"),
		&gorm.Config{},
	)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(
		&User{},
		&Expense{},
	)

	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	fmt.Println("Datbase connected successfully")

	return db
}
