package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-reward-controller/internal/models"
	"github.com/ZnNr/user-reward-controller/internal/repository"

	"github.com/google/uuid"
)

// SQL Queries
const (
	// Получение пользователей с фильтрацией
	GetUsersQuery = `SELECT ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt, LastVisit, VisitCount, Bio, TimeZone, Status 
	FROM Users
	WHERE (Username ILIKE COALESCE($1, Username) OR $1 IS NULL) 
	  AND (Status = COALESCE($2, Status) OR $2 IS NULL);`

	// Получение пользователя по ID
	GetUserByIDQuery = `SELECT ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt, LastVisit, VisitCount, Bio, TimeZone, Status 
	FROM Users 
	WHERE ID = $1;`

	// Создание нового пользователя
	CreateUserQuery = `INSERT INTO Users (ID, Username, Email, Status) VALUES ($1, $2, $3, $4) RETURNING ID, Username, Email, Status, CreatedAt;`

	// Обновление пользователя
	UpdateUserQuery = `UPDATE Users
	SET Username = COALESCE($1, Username),
	    Email = COALESCE($2, Email),
	    Balance = COALESCE($3, Balance),
	    Referrals = COALESCE($4, Referrals),
	    ReferralCode = COALESCE($5, ReferralCode),
	    TasksCompleted = COALESCE($6, TasksCompleted),
	    Bio = COALESCE($7, Bio),
	    TimeZone = COALESCE($8, TimeZone),
	    Status = COALESCE($9, Status),
	    UpdatedAt = CURRENT_TIMESTAMP
	WHERE ID = $10
	RETURNING ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt;`

	// Удаление пользователя
	DeleteUserQuery = `DELETE FROM Users 
	WHERE ID = $1;`

	// Добавление записи в журнал активности пользователя
	AddUserActivityLogQuery = `INSERT INTO UserActivityLog (UserID, ActivityTime)
	VALUES ($1, CURRENT_TIMESTAMP);`

	// Добавление записи о посещении пользователя
	AddUserVisitLogQuery = `INSERT INTO UserVisits (UserID, VisitDate)
	VALUES ($1, CURRENT_TIMESTAMP)
	ON CONFLICT (UserID, VisitDate) DO NOTHING;`

	// Получение пользователей по статусу
	GetUsersByStatusQuery = `SELECT ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt, LastVisit, VisitCount, Bio, TimeZone, Status 
	FROM Users 
	WHERE Status = $1;`

	UpdateBalanceAndReferralsTxQuery = `UPDATE Users SET Balance = Balance + $1, Referrals = Referrals + $2 WHERE ID = $3`

	GetUserByEmailTxQuery = `SELECT ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt, LastVisit, VisitCount, Bio, TimeZone, Status FROM Users WHERE Email = $1`

	GetUserByEmailQuery = `SELECT ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt, LastVisit, VisitCount, Bio, TimeZone, Status FROM Users WHERE Email = $1`

	UpdateBalanceQuery = `UPDATE Users SET Balance = Balance + $1 WHERE ID = $2`

	// Получение лидера по балансу
	GetLeaderByBalanceQuery = `SELECT ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt, LastVisit, VisitCount, Bio, TimeZone, Status 
    FROM Users 
    ORDER BY Balance DESC 
    LIMIT 1;`

	GetUserRankQuery = `
        SELECT COUNT(*) + 1 
        FROM users
        WHERE balance > (SELECT balance FROM users WHERE username = ?)
    `
)

// UserRepository для работы с пользователями
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository создает новый репозиторий пользователей с указанным соединением с БД.
func NewPostgresUserRepository(db *sql.DB) repository.UserRepository {
	return &PostgresUserRepository{db: db}
}

// scanUser сканирует пользователя из строки и возвращает его.
func scanUser(row *sql.Row) (*models.User, error) {
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Balance, &user.Referrals,
		&user.ReferralCode, &user.TasksCompleted, &user.CreatedAt, &user.UpdatedAt,
		&user.LastVisit, &user.VisitCount, &user.Bio, &user.TimeZone, &user.Status)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// scanTopUser сканирует данные о пользователе в структуру TopUser.
func scanTopUser(row *sql.Row) (*models.TopUser, error) {
	var topUser models.TopUser
	err := row.Scan(&topUser.ID, &topUser.Username, &topUser.Email, &topUser.Balance,
		&topUser.Referrals, &topUser.ReferralCode, &topUser.TasksCompleted,
		&topUser.CreatedAt, &topUser.UpdatedAt, &topUser.LastVisit,
		&topUser.VisitCount, &topUser.Bio, &topUser.TimeZone,
		&topUser.Status, &topUser.Rank) // Добавляем поле Rank
	if err != nil {
		return nil, err
	}
	return &topUser, nil
}

// scanUsers сканирует множество пользователей из строк.
func scanUsers(rows *sql.Rows) ([]*models.User, error) {
	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Balance, &user.Referrals,
			&user.ReferralCode, &user.TasksCompleted, &user.CreatedAt, &user.UpdatedAt,
			&user.LastVisit, &user.VisitCount, &user.Bio, &user.TimeZone, &user.Status); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

// Загрузить всех пользователей по фильтру
func (r *PostgresUserRepository) GetUsers(ctx context.Context, filter *models.User) (*models.UsersResponse, error) {
	rows, err := r.db.QueryContext(ctx, GetUsersQuery, filter.Username, filter.Status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := scanUsers(rows)
	if err != nil {
		return nil, err
	}

	response := &models.UsersResponse{
		Users: users,
		Count: len(users), //populate Count based on row count
	}
	return response, nil
}

// Получить пользователя по ID
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, GetUserByIDQuery, id)
	return scanUser(row)
}

// Создать нового пользователя
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.db.QueryRowContext(ctx,
		CreateUserQuery,
		user.ID,
		user.Username,
		user.Email,
		user.Status).
		Scan(&user.ID,
			&user.Username,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
		)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Обновить пользователя
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, UpdateUserQuery, user.Username, user.Email, user.Balance, user.Referrals,
		user.ReferralCode, user.TasksCompleted, user.Bio, user.TimeZone, user.Status, user.ID)
	return scanUser(row)
}

// Удалить пользователя
func (r *PostgresUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, DeleteUserQuery, id)
	return err
}

// Получить пользователей по статусу
func (r *PostgresUserRepository) GetUsersByStatus(ctx context.Context, status models.UserStatus) ([]*models.User, error) {
	rows, err := r.db.QueryContext(ctx, GetUsersByStatusQuery, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUsers(rows)
}

// Получить пользователя по электронной почте
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, GetUserByEmailQuery, email)
	return scanUser(row)
}

// Обновить баланс пользователя
func (r *PostgresUserRepository) UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error {
	_, err := r.db.ExecContext(ctx, UpdateBalanceQuery, amount, id)
	return err
}

// Получить полную информацию о пользователе
func (r *PostgresUserRepository) GetUserFullInfo(ctx context.Context, id uuid.UUID) (string, error) {
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("user not found")
	}
	return fmt.Sprintf("User Info: %+v", user), nil
}

// Получить сводную информацию о пользователе
func (r *PostgresUserRepository) GetUserSummary(ctx context.Context, id uuid.UUID) (*models.UserSummary, error) {
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return &models.UserSummary{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Balance:        user.Balance,
		Referrals:      user.Referrals,
		TasksCompleted: user.TasksCompleted,
	}, nil
}

// Обновление баланса и рефералов в рамках транзакции
func (r *PostgresUserRepository) UpdateBalanceAndReferralsTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, balance float64, referrals int) error {
	_, err := tx.ExecContext(ctx, UpdateBalanceAndReferralsTxQuery, balance, referrals, id)
	return err
}

// Выполнение функции в рамках транзакции
func (r *PostgresUserRepository) WithTransaction(ctx context.Context, f func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %v", rbErr, err)
		}
		return err
	}
	return tx.Commit()
}

// Получение пользователя по ID в рамках транзакции
func (r *PostgresUserRepository) GetUserByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.User, error) {
	row := tx.QueryRowContext(ctx, GetUserByIDQuery, id)
	return scanUser(row)
}

// Получение пользователя по электронной почте в рамках транзакции
func (r *PostgresUserRepository) GetUserByEmailTx(ctx context.Context, tx *sql.Tx, email string) (*models.User, error) {
	row := tx.QueryRowContext(ctx, GetUserByEmailTxQuery, email)
	return scanUser(row)
}

// Создание нового пользователя в рамках транзакции
func (r *PostgresUserRepository) CreateUserTx(ctx context.Context, tx *sql.Tx, user *models.User) (*models.User, error) {
	err := tx.QueryRowContext(ctx, CreateUserQuery, user.ID, user.Username, user.Email, user.Balance,
		user.Referrals, user.ReferralCode, user.TasksCompleted, user.Bio, user.TimeZone, user.Status).
		Scan(&user.ID, &user.Username, &user.Email, &user.Balance, &user.Referrals,
			&user.ReferralCode, &user.TasksCompleted, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Получить лидера по балансу
func (r *PostgresUserRepository) GetLeaderByBalance(ctx context.Context) (*models.TopUser, error) {
	row := r.db.QueryRowContext(ctx, GetLeaderByBalanceQuery)
	topUser, err := scanTopUser(row)
	if err != nil {
		return nil, err
	}
	return topUser, nil
}

func getUserRank(db *sql.DB, username string) (int, error) {
	// Получаем количество завершённых задач для указанного пользователя
	var tasksCompleted int
	err := db.QueryRow("SELECT tasks_completed FROM users WHERE username = ?", username).Scan(&tasksCompleted)
	if err != nil {
		return 0, err
	}

	// Теперь вычисляем ранг, подсчитывая количество пользователей с задачами >= tasksCompleted
	var rank int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE tasks_completed >= ?", tasksCompleted).Scan(&rank)
	if err != nil {
		return 0, err
	}

	// Возвращаем ранг пользователя
	return rank, nil
}

// / Получить топ пользователей с лимитом по количеству
func (r *PostgresUserRepository) GetTopUsers(ctx context.Context, limit int, offset int) ([]models.TopUser, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT ID, Username, Email, Balance, Referrals, ReferralCode, TasksCompleted, CreatedAt, UpdatedAt, LastVisit, VisitCount, Bio, TimeZone, Status
        FROM Users
        ORDER BY Balance DESC, TasksCompleted DESC
        LIMIT $1;
    `, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topUsers []models.TopUser

	for rows.Next() {
		var user models.User
		err = rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Balance,
			&user.Referrals,
			&user.ReferralCode,
			&user.TasksCompleted,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastVisit,
			&user.VisitCount,
			&user.Bio,
			&user.TimeZone,
			&user.Status,
		)
		if err != nil {
			return nil, err
		}

		// Получаем ранг пользователя
		rank, err := getUserRank(r.db, user.Username)
		if err != nil {
			return nil, err // Возвращаем ошибку, если не удалось получить ранг
		}

		// Создаем объект TopUser и добавляем в срез
		topUser := models.TopUser{
			User: user, // Используем user как User
			Rank: rank,
		}
		topUsers = append(topUsers, topUser)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topUsers, nil
}
