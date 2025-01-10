package repository

import (
	"context"
	"github.com/ZnNr/user-reward-controler/internal/models"
	"github.com/google/uuid"
)

// Интерфейс репозитория задач
type TaskRepository interface {
	// GetTasks Получить все задачи с возможностью фильтрации
	GetTasks(ctx context.Context, filter *models.TaskFilter) (*models.TaskResponse, error)

	// GetTaskByID Получить задачу по ID
	GetTaskByID(ctx context.Context, taskId uuid.UUID) (*models.Task, error)

	// CreateTask Создать новую задачу
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)

	// UpdateTask Обновить существующую задачу
	UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error)

	// UpdateTaskStatus Обновить статус существующей задачи по ID
	UpdateTaskStatus(ctx context.Context, taskID string, newStatus int, userID uuid.UUID) (*models.Task, error)

	// DeleteTask Удалить задачу по ID
	DeleteTask(ctx context.Context, taskId uuid.UUID) error
}
