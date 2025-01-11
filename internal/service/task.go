package service

import (
	"context"
	"github.com/ZnNr/user-reward-controller/internal/errors"
	"github.com/ZnNr/user-reward-controller/internal/models"
	"github.com/ZnNr/user-reward-controller/internal/repository"

	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type TaskService struct {
	repo   repository.TaskRepository
	logger *zap.Logger
}

func NewTaskService(repo repository.TaskRepository, logger *zap.Logger) *TaskService {
	return &TaskService{
		repo:   repo,
		logger: logger,
	}
}

// CreateTask создает новую задачу.
func (s *TaskService) CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.Task, error) {
	s.logger.Info("Creating new task",
		zap.String("title", req.Title),
		zap.String("description", req.Description))

	if err := validateTaskRequest(req); err != nil {
		return nil, err
	}

	task := &models.Task{
		TaskID:      generateTaskID(), // Генерация уникального ID задачи
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   time.Now(), // Установка текущего времени в качестве времени создания
		DueDate:     req.DueDate,
		Status:      req.Status,     // Direct assignment because Status is of type TaskStatus
		AssigneeID:  req.AssigneeID, // Convert UUID to string pointer
	}
	// Устанавливаем значение по умолчанию для статуса, если не было указано
	if req.Status == models.NotStarted || req.Status == models.InProgress || req.Status == models.Completed {
		task.Status = req.Status // Если передано допустимое значение, устанавливаем его.
	} else {
		task.Status = models.NotStarted // Устанавливаем статус по умолчанию
	}
	return s.repo.CreateTask(ctx, task)
}

// UpdateTask обновляет существующую задачу.
func (s *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, req *models.UpdateTaskRequest) (*models.Task, error) {
	s.logger.Info("Updating task",
		zap.Any("task_id", id),
		zap.String("Title", req.Title),
		zap.String("Description", req.Description))

	task, err := s.repo.GetTaskByID(ctx, id)
	if err != nil {
		s.logger.Error("Task not found", zap.Error(err))
		return nil, errors.NewNotFound("task not found", nil)
	}

	updateTaskFields(task, req)
	task.UpdatedAt = time.Now()

	updatedTask, err := s.repo.UpdateTask(ctx, task)
	if err != nil {
		s.logger.Error("Failed to update task", zap.Error(err))
		return nil, errors.NewInternal("failed to update task", err)
	}

	return updatedTask, nil
}

// updateTaskFields обновляет поля задачи на основании запроса
func updateTaskFields(task *models.Task, req *models.UpdateTaskRequest) {
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if !req.DueDate.IsZero() {
		task.DueDate = req.DueDate
	}
	if req.Status != 0 { // Если Status - это 0, значит, он не был установлен
		task.Status = req.Status // Простое присваивание, если Status не равен 0
	}
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		task.AssigneeID = req.AssigneeID
	}
}

// UpdateTaskStatus обновляет статус существующей задачи.
func (s *TaskService) UpdateTaskStatus(ctx context.Context, taskID string, newStatus int, userID uuid.UUID) (*models.Task, error) {
	s.logger.Info("Updating task status",
		zap.String("taskID", taskID),
		zap.Int("newStatus", newStatus))

	taskIDParsed, err := uuid.Parse(taskID)
	if err != nil {
		s.logger.Error("Invalid task ID", zap.Error(err))
		return nil, errors.NewBadRequest("invalid task ID", err)
	}

	updatedTask, err := s.repo.UpdateTaskStatus(ctx, taskIDParsed.String(), newStatus, userID)
	if err != nil {
		s.logger.Error("Failed to update task status", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Successfully updated task status", zap.String("taskID", taskID), zap.Int("newStatus", newStatus))
	return updatedTask, nil
}

// GetTasks получает все задачи с возможностью фильтрации
func (s *TaskService) GetTasks(ctx context.Context, filter *models.TaskFilter) (*models.TaskResponse, error) {
	s.logger.Info("Fetching tasks with filter", zap.Any("filter", filter))

	s.logger.Info("Getting tasks with filter",
		zap.Any("Title", filter.Title),
		zap.Any("Status", filter.Status),
		zap.Any("Assignee", filter.AssigneeID),
		zap.Any("CreatedAfter", filter.CreatedAfter),
		zap.Any("CreatedBefore", filter.CreatedBefore),
		zap.Any("DueAfter", filter.DueAfter),
		zap.Any("DueBefore", filter.DueBefore),
		zap.Int("page", filter.Page),
		zap.Int("pageSize", filter.PageSize))

	// Валидация параметров фильтра
	if err := validateFilter(filter); err != nil {
		s.logger.Warn("Invalid filter", zap.Error(err))
		return nil, err
	}

	tasks, err := s.repo.GetTasks(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to fetch tasks", zap.Error(err))
		return nil, errors.NewInternal("failed to fetch tasks", err)
	}

	return tasks, nil
}

// GetTaskByID получает задачу по ID
func (s *TaskService) GetTaskByID(ctx context.Context, id string) (*models.Task, error) {
	if id == "" {
		s.logger.Error("Task ID cannot be empty")
		return nil, errors.NewBadRequest("Task ID cannot be empty", nil)
	}
	taskID, err := uuid.Parse(id) // Parsing string ID into uuid.UUID
	if err != nil {
		s.logger.Error("Invalid task ID", zap.Error(err))
		return nil, errors.NewBadRequest("invalid task ID", err)
	}

	task, err := s.repo.GetTaskByID(ctx, taskID) // taskID uuid.UUID
	if err != nil {
		s.logger.Error("Task not found", zap.Error(err))
		return nil, errors.NewNotFound("task not found", err)
	}

	return task, nil
}

// DeleteTask удаляет задачу по ID
func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		s.logger.Error("Task ID cannot be empty")
		return errors.NewBadRequest("invalid task ID", nil)
	}
	taskID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("Invalid task ID", zap.Error(err))
		return errors.NewBadRequest("invalid task ID", err)
	}

	if err := s.repo.DeleteTask(ctx, taskID); err != nil {
		s.logger.Error("Failed to delete task", zap.Error(err))
		return errors.NewInternal("failed to delete task", err)
	}

	s.logger.Info("Task deleted successfully", zap.String("taskID", id))
	return nil
}

// GetDescription получает описание задачи с опциональным фильтром.
func (s *TaskService) GetDescription(ctx context.Context, id uuid.UUID, page, pageSize int) (*models.DescriptionResponse, error) {
	s.logger.Info("Getting description",
		zap.Any("taskId", id),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize))

	task, err := s.repo.GetTaskByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFound("task not found", err)
	}

	if task.Description == "" {
		return nil, errors.NewNotFound("Description not found", nil)
	}

	verses := strings.Split(task.Description, "\n\n")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := (len(verses) + pageSize - 1) / pageSize
	if page > totalPages {
		return nil, errors.NewNotFound("page out of range", nil)
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(verses) {
		end = len(verses)
	}

	return &models.DescriptionResponse{
		Description: strings.Join(verses[start:end], "\n\n"),
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    pageSize,
	}, nil
}

// validateTaskRequest выполняет проверку валидности запроса на создание или обновление задачи.
func validateTaskRequest(req *models.CreateTaskRequest) error {
	if req.Title == "" {
		return errors.NewValidation("task title cannot be empty", nil)
	}

	//if req.DueDate.IsZero() {
	//	return errors.NewValidation("due date cannot be empty", nil)
	//}
	if req.DueDate.Before(time.Now()) {
		return errors.NewValidation("due date cannot be in the past", nil)
	}
	return nil
}

// generateTaskID создает уникальный идентификатор задачи
func generateTaskID() string {
	return uuid.New().String() // Генерация нового UUID
}

// validateFilter выполняет проверку валидации фильтра задач.
func validateFilter(filter *models.TaskFilter) error {
	if filter.Page < 1 {
		return errors.NewValidation("page must be at least 1", nil)
	}
	if filter.PageSize < 1 {
		return errors.NewValidation("page size must be at least 1", nil)
	}
	//доп проверки
	return nil
}
