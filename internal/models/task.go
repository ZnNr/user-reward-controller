package models

import (
	"time"
)

// TaskStatus определяет возможные статусы задачи
type TaskStatus int

const (
	NotStarted TaskStatus = iota + 1
	InProgress
	Completed
	Canceled
)

// Task представляет собой задание
type Task struct {
	TaskID      string     `json:"task_id" validate:"required"` // Уникальный идентификатор задания
	Title       string     `json:"title" validate:"required"`   // Заголовок задания
	Description string     `json:"description,omitempty"`       // Описание задания
	CreatedAt   time.Time  `json:"created_at"`                  // Дата создания задания
	UpdatedAt   time.Time  `json:"updated_at"`                  // Дата и время последнего обновления записи
	DueDate     *time.Time `json:"due_date,omitempty"`          // Дедлайн (необязательный)
	Status      TaskStatus `json:"status"`                      // Статус задания
	AssigneeID  *string    `json:"assignee_id,omitempty"`       // Уникальный идентификатор исполнителя (необязательный)
}

// BaseTaskRequest представляет собой базовую структуру для создания и обновления задания
type BaseTaskRequest struct {
	Title       string     `json:"title" validate:"required"`  // Заголовок задания
	Description string     `json:"description,omitempty"`      // Описание задания
	DueDate     *time.Time `json:"due_date,omitempty"`         // Дедлайн (необязательный)
	Status      TaskStatus `json:"status" validate:"required"` // Статус задания, теперь обязательный
	AssigneeID  *string    `json:"assignee_id,omitempty"`      // Уникальный идентификатор исполнителя (необязательный)
}

// CreateTaskRequest представляет собой запрос на создание задания
type CreateTaskRequest struct {
	BaseTaskRequest
}

// UpdateTaskRequest представляет собой запрос на обновление задания
type UpdateTaskRequest struct {
	TaskID string `json:"task_id" validate:"required"` // Уникальный идентификатор задания
	BaseTaskRequest
	UpdatedAt time.Time `json:"updated_at"` // Дата и время последнего обновления записи
}

// TaskFilter используется для фильтрации задач
type TaskFilter struct {
	Title         string     `json:"title,omitempty"` // Фильтрация по заголовку
	Description   string     `json:"description,omitempty"`
	Status        TaskStatus `json:"status,omitempty"`         // Заменили string на TaskStatus для типизации
	AssigneeID    string     `json:"assignee_id,omitempty"`    // Фильтрация по ID исполнителя
	CreatedAfter  *time.Time `json:"created_after,omitempty"`  // Исправлено имя поля
	CreatedBefore *time.Time `json:"created_before,omitempty"` // Исправлено имя поля
	DueAfter      *time.Time `json:"due_after,omitempty"`      // Исправлено имя поля
	DueBefore     *time.Time `json:"due_before,omitempty"`     // Исправлено имя поля
	Page          int        `json:"page"`                     // Номер текущей страницы
	PageSize      int        `json:"page_size"`                // Размер страницы (количество элементов на странице)
}

// String возвращает строковое представление статуса задачи
func (s TaskStatus) String() string {
	switch s {
	case NotStarted:
		return "Not Started"
	case InProgress:
		return "In Progress"
	case Completed:
		return "Completed"
	case Canceled:
		return "Canceled"
	default:
		return "Unknown"
	}
}

// TaskResponse представляет структуру ответа со списком задач и информацией о пагинации.
type TaskResponse struct {
	Tasks      []Task `json:"tasks"`       // Список задач
	Page       int    `json:"page"`        // Номер текущей страницы
	TotalPages int    `json:"total_pages"` // Общее количество страниц
	TotalItems int    `json:"total_items"` // Общее количество задач
	PageSize   int    `json:"page_size"`   // Количество элементов на странице
}

// DescriptionResponse представляет структуру ответа с текстом куплетов и информацией о пагинации.
type DescriptionResponse struct {
	Description string `json:"description,omitempty"` // Описание задания
	CurrentPage int    `json:"current_page"`          // Номер текущей страницы
	TotalPages  int    `json:"total_pages"`           // Общее количество страниц
	PageSize    int    `json:"page_size"`             // Количество элементов на странице
}
