package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-reward-controller/internal/errors"
	"github.com/ZnNr/user-reward-controller/internal/models"
	"github.com/ZnNr/user-reward-controller/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/mail"
	"strings"
	"time"
)

// UserService представляет собой службу управления пользователями
type UserService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

// NewUserService создает новый экземпляр UserService
func NewUserService(repo repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

// GetUsers возвращает список пользователей, соответствующих заданному фильтру
func (u *UserService) GetUsers(ctx context.Context, filter *models.User) (*models.UsersResponse, error) {
	users, err := u.repo.GetUsers(ctx, filter)
	if err != nil {
		u.logger.Error("Error getting users", zap.Error(err))
		return nil, err
	}
	return users, nil
}

// GetUserByID получает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if ctx == nil {
		return nil, errors.NewInvalidArgument("context is required", nil)
	}
	if err := validateUUID(id); err != nil {
		s.logger.Error("Invalid user ID", zap.Error(err))
		return nil, err
	}

	userID := uuid.MustParse(id)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("User not found", zap.String("userID", id), zap.Error(err))
		return nil, errors.NewNotFound("user not found", err)
	}
	return user, nil
}

// CreateUser создает нового пользователя
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	s.logger.Info("Creating new user", zap.String("username", req.Username), zap.String("email", req.Email))

	if err := validateUserRequest(req); err != nil {
		s.logger.Error("User request validation failed", zap.Error(err))
		return nil, err
	}

	user := &models.User{
		ID:        generateUserID(),
		Email:     req.Email,
		CreatedAt: time.Now(),
		Username:  req.Username,
	}

	createdUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("User created successfully", zap.String("userID", createdUser.ID))
	return createdUser, nil
}

// validateUserRequest проверяет корректность данных запроса на создание пользователя
func validateUserRequest(req *models.CreateUserRequest) error {
	if req.Username == "" {
		return errors.NewBadRequest("username cannot be empty", nil)
	}
	if req.Email == "" || !isValidEmail(req.Email) {
		return errors.NewBadRequest("invalid email", nil)
	}
	return nil
}

// isValidEmail проверяет корректность email-адреса
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil && strings.Contains(email, "@")
}

// generateUserID генерирует уникальный ID для пользователя.
func generateUserID() string {
	// Реализация генерации уникального идентификатора пользователя
	return uuid.New().String() // Пример использования UUID
}

// UpdateUser обновляет информацию о пользователе
func (s *UserService) UpdateUser(ctx context.Context, req *models.UpdateUserRequest) (*models.User, error) {
	if err := validateUUID(req.UserID); err != nil {
		s.logger.Error("Invalid user ID", zap.Error(err))
		return nil, err
	}

	userID := uuid.MustParse(req.UserID)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("User not found", zap.String("userID", req.UserID), zap.Error(err))
		return nil, errors.NewNotFound("user not found", nil)
	}

	if err := updateUserFields(user, req); err != nil {
		s.logger.Error("Failed to update user fields", zap.Error(err))
		return nil, err
	}

	updatedUser, err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, errors.NewInternal("failed to update user", err)
	}

	return updatedUser, nil
}

// updateUserFields обновляет поля пользователя на основании запроса
func updateUserFields(user *models.User, req *models.UpdateUserRequest) error {
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		if !isValidEmail(*req.Email) {
			return errors.NewBadRequest("invalid email", nil)
		}
		user.Email = *req.Email
	}
	if req.Balance != nil {
		if err := user.UpdateBalance(*req.Balance - user.Balance); err != nil {
			return err
		}
	}
	if req.ReferralCode != nil {
		user.ReferralCode = *req.ReferralCode
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.TimeZone != nil {
		user.TimeZone = *req.TimeZone
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	user.UpdatedAt = time.Now()
	return nil
}

// DeleteUser удаляет пользователя из базы данных
func (u *UserService) DeleteUser(ctx context.Context, id string) error {
	if err := validateUUID(id); err != nil {
		u.logger.Error("Invalid user ID", zap.Error(err))
		return err
	}

	err := u.repo.DeleteUser(ctx, uuid.MustParse(id))
	if err != nil {
		u.logger.Error("Error deleting user", zap.String("id", id), zap.Error(err))
		return err
	}
	return nil
}

// validateUUID проверяет корректность формата UUID
func validateUUID(id string) error {
	if id == "" {
		return errors.NewBadRequest("user ID cannot be empty", nil)
	}
	if _, err := uuid.Parse(id); err != nil {
		return errors.NewBadRequest("invalid user ID", err)
	}
	return nil
}

// GetUsersByStatus возвращает пользователей по заданному статусу
func (u *UserService) GetUsersByStatus(ctx context.Context, status models.UserStatus) ([]*models.User, error) {
	users, err := u.repo.GetUsersByStatus(ctx, status)
	if err != nil {
		u.logger.Error("error getting users by status", zap.String("status", string(status)), zap.Error(err))
		return nil, err
	}
	return users, nil
}

// GetUserByEmail возвращает пользователя по его электронной почте
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("error getting user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	return user, nil
}

// UpdateBalance обновляет баланс пользователя на заданную сумму
func (s *UserService) UpdateBalance(ctx context.Context, id string, amount float64) error {
	// Конвертация строки id в UUID
	userID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("invalid UUID format", zap.String("id", id), zap.Error(err))
		return errors.NewInternal("invalid user ID format", err)
	}

	// Получаем пользователя по ID
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("error getting user for balance update", zap.String("id", id), zap.Error(err))
		return err
	}

	// Обновление баланса
	if err := user.UpdateBalance(amount); err != nil {
		return fmt.Errorf("error updating balance: %w", err)
	}

	// Обновляем пользователя в репозитории
	if _, err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("error updating user balance", zap.String("id", id), zap.Error(err))
		return err
	}

	// Логирование успешного обновления
	s.logger.Info("user balance updated", zap.String("id", id), zap.Float64("newBalance", user.Balance))
	return nil
}

// GetUserFullInfo получает полную информацию о пользователе по ID
func (s *UserService) GetUserFullInfo(ctx context.Context, id string) (string, error) {
	if id == "" {
		s.logger.Error("User ID cannot be empty")
		return "", errors.NewBadRequest("user ID cannot be empty", nil)
	}

	userID, err := uuid.Parse(id) // Преобразование строкового ID в uuid.UUID
	if err != nil {
		s.logger.Error("Invalid user ID", zap.Error(err))
		return "", errors.NewBadRequest("invalid user ID", err)
	}

	user, err := s.repo.GetUserByID(ctx, userID) // Получаем пользователя по ID
	if err != nil {
		s.logger.Error("User not found", zap.Error(err))
		return "", errors.NewNotFound("user not found", err)
	}

	// Считаем активность
	weeklyActivity := user.GetWeeklyActivity()
	monthlyActivity := user.GetMonthlyActivity()

	// Формируем полную информацию
	userInfo := fmt.Sprintf(
		"User ID: %s\nName: %s\nEmail: %s\nBalance: %f\nReferrals: %d\nReferral Code: %s\n"+
			"Tasks Completed: %d\nCreated At: %s\nUpdated At: %s\nBio: %s\nTime Zone: %s\n"+
			"Weekly Activity: %d\nMonthly Activity: %d\n",
		user.ID,
		user.Username,
		user.Email,
		user.Balance,
		user.Referrals,
		user.ReferralCode,
		user.TasksCompleted,
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
		user.Bio,
		user.TimeZone,
		weeklyActivity,  // Добавлен вывод недельной активности
		monthlyActivity, // Добавлен вывод месячной активности
	)

	return userInfo, nil // Возвращаем строку с информацией о пользователе
}

// GetUserSummary получает сокращенную информацию о пользователе по ID
func (s *UserService) GetUserSummary(ctx context.Context, id string) (*models.UserSummary, error) {
	if id == "" {
		s.logger.Error("User ID cannot be empty")
		return nil, errors.NewBadRequest("user ID cannot be empty", nil)
	}

	userID, err := uuid.Parse(id) // Преобразование строкового ID в uuid.UUID
	if err != nil {
		s.logger.Error("Invalid user ID", zap.Error(err))
		return nil, errors.NewBadRequest("invalid user ID", err)
	}

	user, err := s.repo.GetUserByID(ctx, userID) // userID uuid.UUID
	if err != nil {
		s.logger.Error("User not found", zap.Error(err))
		return nil, errors.NewNotFound("user not found", err)
	}

	userSummary := &models.UserSummary{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	return userSummary, nil
}

// InviteUser позволяет существующему пользователю пригласить нового пользователя.
func (s *UserService) InviteUser(ctx context.Context, inviterID string, inviteeEmail string) error {
	if ctx == nil {
		return errors.NewInvalidArgument("context is required", nil)
	}
	if inviterID == "" {
		return errors.NewInvalidArgument("inviterID cannot be empty", nil)
	}
	if inviteeEmail == "" || !isValidEmail(inviteeEmail) {
		return errors.NewBadRequest("invalid email format", nil)
	}

	return s.repo.WithTransaction(ctx, func(tx *sql.Tx) error {
		inviterUUID, err := uuid.Parse(inviterID)
		if err != nil {
			s.logger.Error("Invalid inviter ID", zap.Error(err))
			return errors.NewBadRequest("invalid inviterID", err)
		}

		inviter, err := s.repo.GetUserByIDTx(ctx, tx, inviterUUID)
		if err != nil {
			s.logger.Error("Failed to retrieve inviter", zap.String("id", inviterUUID.String()), zap.Error(err))
			return errors.NewInternal("failed to retrieve inviter", err)
		}

		if inviter == nil {
			return errors.NewNotFound("inviter not found", nil)
		}

		existingInvitee, err := s.repo.GetUserByEmailTx(ctx, tx, inviteeEmail)
		if err != nil && !errors.IsNotFound(err) {
			s.logger.Error("Failed to check if invitee exists", zap.Error(err))
			return errors.NewInternal("unable to check invitee existence", err)
		}

		if existingInvitee != nil {
			return errors.NewAlreadyExists("user already exists", nil)
		}

		inviteeUsername := inviteeEmail[:strings.Index(inviteeEmail, "@")]
		invitee := models.BrandNewUser(inviteeEmail, inviteeUsername, models.Pending)

		if err, _ := s.repo.CreateUserTx(ctx, tx, invitee); err != nil {
			s.logger.Error("Failed to create new user")
			return errors.NewInternal("failed to create new user", nil)
		}

		const bonusPoints = 10.0
		newBalance := inviter.Balance + bonusPoints
		newReferrals := inviter.Referrals + 1

		if err := s.repo.UpdateBalanceAndReferralsTx(ctx, tx, inviterUUID, newBalance, newReferrals); err != nil {
			s.logger.Error("Failed to update inviter's balance and referrals", zap.Error(err))
			return errors.NewInternal("failed to update inviter's balance and referrals", err)
		}

		s.logger.Info("User invited successfully",
			zap.String("inviterID", inviterID),
			zap.String("inviteeEmail", inviteeEmail),
			zap.Float64("bonusPoints", bonusPoints))

		return nil
	})
}

// Получить лидера по балансу
func (s *UserService) GetLeaderByBalance(ctx context.Context) (*models.TopUser, error) {
	leader, err := s.repo.GetLeaderByBalance(ctx)
	if err != nil {
		s.logger.Error("Error getting leader by balance", zap.Error(err))
		return nil, fmt.Errorf("could not get leader: %w", err)
	}
	if leader == nil {
		return nil, errors.NewNotFound("no leaders found", nil)
	}
	return leader, nil
}

// GetTopUsers возвращает список пользователей с их балансами,
// начиная от лидера до пользователей с наименьшим балансом, с поддержкой пагинации.
func (s *UserService) GetTopUsers(ctx context.Context, limit int, offset int) (*models.TopUsers, error) {
	// Проверка валидности limit и offset
	if limit <= 0 {
		return nil, errors.NewValidation("limit must be greater than 0", nil)
	}
	if offset < 0 {
		return nil, errors.NewValidation("offset cannot be negative", nil)
	}

	// Получение топ пользователей с использованием метода репозитория
	users, err := s.repo.GetTopUsers(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Error getting top users", zap.Error(err))
		return nil, fmt.Errorf("could not get top users: %w", err)
	}

	// Проверка наличия пользователей с указанным limit и offset
	if len(users) == 0 {
		s.logger.Warn("No users found for the specified limit and offset", zap.Int("limit", limit), zap.Int("offset", offset))
		return &models.TopUsers{Users: []models.TopUser{}, Count: 0}, nil // Возвращаем пустой TopUsers
	}

	// Преобразование пользователей в TopUser с учетом их ранга
	topUsers := make([]models.TopUser, len(users))
	for i, user := range users {
		topUsers[i] = models.TopUser{
			User: user.User,      // Извлекаем объект User из user
			Rank: offset + i + 1, // Устанавливаем ранг с учетом offset
		}
	}

	// Формируем и возвращаем результат
	return &models.TopUsers{
		Users: topUsers,
		Count: len(topUsers),
	}, nil
}
