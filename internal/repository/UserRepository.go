package repository

import (
	"context"
	"database/sql"
	"github.com/ZnNr/user-reward-controller/internal/models"

	"github.com/google/uuid"
)

// UserRepository определяет методы для взаимодействия с данными пользователя в базе данных
type UserRepository interface {
	// GetUsers возвращает список пользователей, соответствующих заданному фильтру
	GetUsers(ctx context.Context, filter *models.User) (*models.UsersResponse, error)

	// GetUserByID возвращает пользователя по его уникальному идентификатору
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// CreateUser создает нового пользователя в базе данных
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)

	// UpdateUser обновляет информацию о существующем пользователе
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)

	// DeleteUser удаляет пользователя из базы данных
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// GetUsersByStatus возвращает пользователей по заданному статусу
	GetUsersByStatus(ctx context.Context, status models.UserStatus) ([]*models.User, error)

	// GetUserByEmail возвращает пользователя по его электронной почте
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// UpdateBalance увеличивает/уменьшает баланс пользователя на заданную сумму
	UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error

	// GetUserFullInfo возвращает детальную информацию о пользователе в виде строки
	GetUserFullInfo(ctx context.Context, id uuid.UUID) (string, error)

	// GetUserSummary возвращает сводную информацию о пользователе
	GetUserSummary(ctx context.Context, id uuid.UUID) (*models.UserSummary, error)

	// UpdateBalanceTx обновляет баланс пользователя в рамках транзакции
	//UpdateBalanceTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, amount float64) error

	// WithTransaction выполняет функцию в рамках транзакции
	WithTransaction(ctx context.Context, f func(tx *sql.Tx) error) error

	// GetUserByIDTx возвращает пользователя по идентификатору в рамках транзакции
	GetUserByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.User, error)

	// GetUserByEmailTx возвращает пользователя по электронной почте в рамках транзакции
	GetUserByEmailTx(ctx context.Context, tx *sql.Tx, email string) (*models.User, error)

	// CreateUserTx создает нового пользователя в рамках транзакции
	CreateUserTx(ctx context.Context, tx *sql.Tx, user *models.User) (*models.User, error)

	// UpdateBalanceAndReferralsTx обновляет баланс и количество рефералов в рамках транзакции
	UpdateBalanceAndReferralsTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, balance float64, referrals int) error

	GetTopUsers(ctx context.Context, limit int, offset int) ([]models.TopUser, error)

	GetLeaderByBalance(ctx context.Context) (*models.TopUser, error) // Новый метод
}
