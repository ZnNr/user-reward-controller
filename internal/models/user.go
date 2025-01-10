package models

import (
	"fmt"
	"time"
)

// UserStatus определяет возможные статусы пользователя
type UserStatus int

const (
	Active UserStatus = iota
	Suspended
	Banned
	Pending // Новый статус ожидания
)

// User представляет модель данных пользователя
type User struct {
	ID             string      `json:"id" validate:"required"`
	Username       string      `json:"username" validate:"required"`
	Email          string      `json:"email" validate:"required,email"`
	Balance        float64     `json:"balance" validate:"gte=0"`
	Referrals      int         `json:"referrals" validate:"gte=0"`
	ReferralCode   string      `json:"referralCode"`
	TasksCompleted int         `json:"tasksCompleted" validate:"gte=0"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
	LastVisit      time.Time   `json:"lastVisit,omitempty"`   // Время последнего посещения
	VisitCount     int         `json:"visitCount"`            // Общее количество посещений
	ActivityLog    []time.Time `json:"activityLog,omitempty"` // Лог времени посещений
	Bio            string      `json:"bio,omitempty"`
	TimeZone       string      `json:"timeZone,omitempty"`
	Status         UserStatus  `json:"status"`
}

// NewUser представляет модель для нового пользователя перед активацией
type NewUser struct {
	Username     string     `json:"username" validate:"required"`    // Имя пользователя, обязательное поле
	Email        string     `json:"email" validate:"required,email"` // Электронная почта, обязательное поле с валидацией на корректность
	ReferralCode string     `json:"referralCode,omitempty"`          // Реферальный код, не обязательное поле
	Bio          string     `json:"bio,omitempty"`                   // Биография, не обязательное поле
	TimeZone     string     `json:"timeZone,omitempty"`              // Часовой пояс, не обязательное поле
	Status       UserStatus `json:"status,omitempty"`                // Статус пользователя, по умолчанию Pending
}

// CreateUserRequest представляет модель запроса на создание нового пользователя
type CreateUserRequest struct {
	Username     string `json:"username" validate:"required"`    // Имя пользователя, обязательное поле
	Email        string `json:"email" validate:"required,email"` // Электронная почта, обязательное поле с валидацией на корректность
	ReferralCode string `json:"referralCode,omitempty"`          // Реферальный код, не обязательное поле
	Bio          string `json:"bio,omitempty"`                   // Биография, не обязательное поле
	TimeZone     string `json:"timeZone,omitempty"`              // Часовой пояс, не обязательное поле
}

// TopUser представляет пользователя с высшими показателями и использует User
type TopUser struct {
	User     // Встраиваем все поля User
	Rank int `json:"rank"` // Ранг пользователя в топе
}

// TopUsers отвечает за представление списка пользователей в топе
type TopUsers struct {
	Users []TopUser `json:"users"` // Список пользователей в топе
	Count int       `json:"count"` // Общее количество пользователей в топе
}

// UpdateUserRequest представляет модель запроса на обновление информации о пользователе.
type UpdateUserRequest struct {
	UserID       string      `json:"userId" validate:"required"`       // Идентификатор пользователя, обязательное поле
	Username     *string     `json:"username,omitempty"`               // Имя пользователя, может быть пустым
	Email        *string     `json:"email,omitempty" validate:"email"` // Электронная почта, может быть пустым, но если присутствует – должна соответствовать валидации email
	Balance      *float64    `json:"balance,omitempty"`                // Баланс, может быть пустым
	ReferralCode *string     `json:"referralCode,omitempty"`           // Реферальный код, может быть пустым
	Bio          *string     `json:"bio,omitempty"`                    // Биография, может быть пустым
	TimeZone     *string     `json:"timeZone,omitempty"`               // Часовой пояс, может быть пустым
	Status       *UserStatus `json:"status,omitempty"`                 // Статус пользователя, может быть пустым
}

// Структура краткой информации о пользователе
type UserSummary struct {
	ID             string  // Уникальный идентификатор
	Username       string  // Имя пользователя
	Email          string  // Адрес электронной почты
	Balance        float64 // Баланс пользователя
	Referrals      int
	TasksCompleted int
	CreatedAt      time.Time // Дата создания аккаунта
}

// UpdateBalance Метод для обновления баланса пользователя
func (u *User) UpdateBalance(amount float64) error {
	newBalance := u.Balance + amount
	if newBalance < 0 {
		return fmt.Errorf("invalid balance: cannot go below zero")
	}
	now := time.Now()
	u.Balance = newBalance
	u.UpdatedAt = now
	return nil
}

// Метод для добавления задания
func (u *User) CompleteTask() {
	now := time.Now()
	u.TasksCompleted++ // увеличиваем количество выполненных заданий
	u.UpdatedAt = now  // обновляем время последнего изменения
}

// Метод для обновления последнего времени посещения
func (u *User) UpdateLastVisit() {
	now := time.Now()
	u.LastVisit = now
	u.VisitCount++
	u.ActivityLog = append(u.ActivityLog, now)
}

func (u *User) GetWeeklyActivity() int {
	// Определяем текущую дату и дату недели назад
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)

	count := 0
	// Проходим по логам активности и считаем количество посещений за последнюю неделю
	for _, visit := range u.ActivityLog {
		if visit.After(weekAgo) && visit.Before(now) {
			count++
		}
	}
	return count
}

func (u *User) GetMonthlyActivity() int {
	// Определяем текущую дату и дату месяц назад
	now := time.Now()
	monthAgo := now.AddDate(0, -1, 0)

	count := 0
	// Проходим по логам активности и считаем количество посещений за последний месяц
	for _, visit := range u.ActivityLog {
		if visit.After(monthAgo) && visit.Before(now) {
			count++
		}
	}
	return count
}

func BrandNewUser(email, username string, status UserStatus) *User {
	return &User{
		Email:    email,
		Username: username,
		Status:   status,
	}
}

// UsersResponse представляет структуру ответа.
type UsersResponse struct {
	Users []*User `json:"users"`
	Count int     `json:"count"` // total user count can be included
}
