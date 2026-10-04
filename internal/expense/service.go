package expense

import (
	"expense-tracker/internal/database"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) GetOrCreateUser(
	telegramId int64, username string, firstname string,
) (*database.User, error) {
	var user database.User

	err := s.db.
		Where("telegram_id = ?", telegramId).First(&user).Error

	if err == nil {
		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	user = database.User{
		TelegramID: telegramId,
		Username:   username,
		FirstName:  firstname,
	}

	err = s.db.Create(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) CreateExpense(
	userID uint,
	amount float64,
	category string,
	description string,
	expenseDate time.Time,
) (*database.Expense, error) {

	expense := database.Expense{
		UserID:      userID,
		Amount:      amount,
		Category:    category,
		Description: description,
		ExpenseDate: expenseDate,
	}

	err := s.db.Create(&expense).Error

	if err != nil {
		return nil, err
	}

	return &expense, nil
}
func (s *Service) GetExpenses(
	filter ExpenseFilter,
) ([]database.Expense, error) {

	var expenses []database.Expense

	query := s.db.
		Where("user_id = ?", filter.UserID)

	if filter.StartDate != nil {
		query = query.Where(
			"expense_date >= ?",
			*filter.StartDate,
		)
	}

	if filter.EndDate != nil {
		query = query.Where(
			"expense_date < ?",
			*filter.EndDate,
		)
	}

	if filter.Category != "" {
		query = query.Where(
			"category = ?",
			filter.Category,
		)
	}

	if filter.MinAmount != nil {
		query = query.Where(
			"amount > ?",
			*filter.MinAmount,
		)
	}

	if filter.MaxAmount != nil {
		query = query.Where(
			"amount < ?",
			*filter.MaxAmount,
		)
	}

	err := query.
		Order("expense_date DESC").
		Find(&expenses).
		Error

	return expenses, err
}

func (s *Service) DeleteAllExpenses(userID uint) (int64, error) {

	result := s.db.
		Where("user_id = ?", userID).
		Delete(&database.Expense{})

	return result.RowsAffected, result.Error
}

func (s *Service) DeleteExpensesByCategory(
	userID uint,
	category string,
) (int64, error) {

	result := s.db.
		Where(
			"user_id = ? AND category = ?",
			userID,
			category,
		).
		Delete(&database.Expense{})

	return result.RowsAffected, result.Error
}
