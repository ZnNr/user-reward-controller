package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-reward-controler/internal/errors"
	"github.com/ZnNr/user-reward-controler/internal/models"
	"github.com/ZnNr/user-reward-controler/internal/repository"
	"github.com/google/uuid"
)

// SQL Queries
const (
	addTaskQuery = `
	INSERT INTO tasks (task_id, title, description, due_date, status, assignee_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING task_id, title, description, created_at, due_date, status, assignee_id`

	getTaskByIDQuery = `
	SELECT task_id, title, description, created_at, updated_at, due_date, status, assignee_id
	FROM tasks
	WHERE task_id = $1`

	updateTaskQuery = `
	UPDATE tasks
	SET title = $1, 
		description = $2, 
		due_date = $3,
		status = $4,
		assignee_id = $5,
		updated_at = NOW()
	WHERE task_id = $6
	RETURNING task_id, title, description, created_at, updated_at, due_date, status, assignee_id`

	deleteTaskQuery = `DELETE FROM tasks WHERE task_id = $1`

	checkTaskExistsQuery = `
	SELECT EXISTS (
	    SELECT 1 
	    FROM tasks 
	    WHERE task_id = $1
	);`

	checkTaskDuplicateQuery = `SELECT COUNT(*) FROM tasks WHERE title = $1 AND description = $2 AND task_id <> $3`

	countTasksQuery = `SELECT COUNT(*) FROM tasks WHERE (title ILIKE COALESCE($1, title) OR $1 IS NULL) AND (status = COALESCE($2, status) OR $2 IS NULL) AND (assignee_id = COALESCE($3, assignee_id) OR $3 IS NULL) AND (created_at >= COALESCE($4, created_at) OR $4 IS NULL) AND (created_at <= COALESCE($5, created_at) OR $5 IS NULL) AND (due_date >= COALESCE($6, due_date) OR $6 IS NULL) AND (due_date <= COALESCE($7, due_date) OR $7 IS NULL);`

	getAllTasksQuery = `
	SELECT task_id, title, description, created_at, updated_at, due_date, status, assignee_id 
	FROM tasks 
	WHERE ($1 IS NULL OR title ILIKE '%' || $1 || '%') AND ($2 IS NULL OR created_at >= $2) AND ($3 IS NULL OR created_at <= $3) AND ($4 IS NULL OR due_date >= $4) AND ($5 IS NULL OR due_date <= $5) AND ($6 IS NULL OR status = $6) AND ($7 IS NULL OR assignee_id = $7)
	ORDER BY created_at DESC
	LIMIT $8 OFFSET $9;`

	userTaskStatusChangeQuery = `UPDATE users SET TasksCompleted = TasksCompleted + 1 WHERE id = $1`
)

type PostgresTaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository with a given database connection.
func NewPostgresTaskRepository(db *sql.DB) repository.TaskRepository {
	return &PostgresTaskRepository{db: db}
}

// CreateTask сохраняет новую задачу в базу данных
func (r *PostgresTaskRepository) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	// проверяем на дубликаты
	if isDuplicate, err := r.checkForDuplicateTask(ctx, task, ""); err != nil {
		return nil, err
	} else if isDuplicate {
		return nil, errors.NewAlreadyExists("a task with the same title and description already exists", nil)
	}

	if err := r.insertTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (r *PostgresTaskRepository) insertTask(ctx context.Context, task *models.Task) error {
	err := r.db.QueryRowContext(
		ctx,
		addTaskQuery,
		task.TaskID,
		task.Title,
		task.Description,
		task.DueDate,
		task.Status,
		task.AssigneeID,
	).Scan(
		&task.TaskID,
		&task.Title,
		&task.Description,
		&task.CreatedAt,
		&task.DueDate,
		&task.Status,
		&task.AssigneeID,
		&task.UpdatedAt,
	)
	if err != nil {
		return errors.NewInternal("failed to insert task", err)
	}
	return nil
}

// GetTasks retrieves a paginated list of tasks based on filters.
func (r *PostgresTaskRepository) GetTasks(ctx context.Context, filter *models.TaskFilter) (*models.TaskResponse, error) {
	setDefaultFilterValues(filter)

	totalItems, err := r.countTotalTasks(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Handling pagination
	totalPages := (totalItems + filter.PageSize - 1) / filter.PageSize
	if filter.Page > totalPages {
		return nil, errors.NewNotFound(fmt.Sprintf("page %d does not exist, total pages: %d", filter.Page, totalPages), nil)
	}

	offset := (filter.Page - 1) * filter.PageSize
	tasks, err := r.getTasksByPage(ctx, filter, offset)
	if err != nil {
		return nil, err
	}

	return &models.TaskResponse{
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Tasks:      tasks,
	}, nil
}

// getTasksByPage retrieves tasks for a specific page.
func (r *PostgresTaskRepository) getTasksByPage(ctx context.Context, filter *models.TaskFilter, offset int) ([]models.Task, error) {
	rows, err := r.db.QueryContext(ctx, getAllTasksQuery,
		filter.Title,
		filter.CreatedAfter,
		filter.CreatedBefore,
		filter.DueAfter,
		filter.DueBefore,
		filter.Status,
		filter.AssigneeID,
		filter.PageSize,
		offset,
	)
	if err != nil {
		return nil, errors.NewInternal("failed to query tasks", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(
			&task.TaskID,
			&task.Title,
			&task.Description,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DueDate,
			&task.Status,
			&task.AssigneeID,
		); err != nil {
			return nil, errors.NewInternal("failed to scan task", err)
		}
		tasks = append(tasks, task)
	}

	// Handle row iteration error
	if err = rows.Err(); err != nil {
		return nil, errors.NewInternal("error occurred while iterating over tasks", err)
	}

	return tasks, nil
}

// setDefaultFilterValues sets default values for the filter.
func setDefaultFilterValues(filter *models.TaskFilter) {
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
}

// countTotalTasks counts the total number of tasks matching the filter.
func (r *PostgresTaskRepository) countTotalTasks(ctx context.Context, filter *models.TaskFilter) (int, error) {
	var totalItems int
	err := r.db.QueryRowContext(
		ctx,
		countTasksQuery,
		filter.Title,
		filter.Status,
		filter.AssigneeID,
		filter.CreatedAfter,
		filter.CreatedBefore,
		filter.DueAfter,
		filter.DueBefore,
	).Scan(&totalItems)

	if err != nil {
		return 0, errors.NewInternal("failed to count tasks", err)
	}
	return totalItems, nil
}

// GetTaskByID retrieves task information by its ID from PostgreSQL.
func (r *PostgresTaskRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.QueryRowContext(ctx, getTaskByIDQuery, id).Scan(&task.TaskID,
		&task.Title,
		&task.Description,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DueDate,
		&task.Status,
		&task.AssigneeID)
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFound("task not found", nil)
	} else if err != nil {
		return nil, errors.NewInternal("failed to get task", err)
	}
	return &task, nil
}

// UpdateTask updates an existing task in the database.
func (r *PostgresTaskRepository) UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	if exists, err := r.checkTaskExists(ctx, task.TaskID); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.NewNotFound("task not found", nil)
	}

	if isDuplicate, err := r.checkForDuplicateTask(ctx, task, task.TaskID); err != nil {
		return nil, err
	} else if isDuplicate {
		return nil, errors.NewAlreadyExists("a task with the same title and description already exists", nil)
	}

	if err := r.updateTask(ctx, task); err != nil {
		return nil, errors.NewInternal("failed to update task", err)
	}
	return task, nil
}

// updateTask обновляет существующую задачу в базе данных.
func (r *PostgresTaskRepository) updateTask(ctx context.Context, task *models.Task) error {
	err := r.db.QueryRowContext(ctx, updateTaskQuery,
		task.Title,
		task.Description,
		task.DueDate,
		task.Status,
		task.AssigneeID,
		task.TaskID,
	).Scan(
		&task.TaskID,
		&task.Title,
		&task.Description,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DueDate,
		&task.Status,
		&task.AssigneeID,
	)
	if err != nil {
		return errors.NewInternal("failed to update task", err)
	}
	return nil
}

// checkTaskExists verifies whether a task exists by its ID.
func (r *PostgresTaskRepository) checkTaskExists(ctx context.Context, taskID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, checkTaskExistsQuery, taskID).Scan(&exists)
	if err != nil {
		return false, errors.NewInternal("failed to check task existence", err)
	}
	return exists, nil
}

// checkForDuplicateTask checks for duplicate tasks based on title and description.
func (r *PostgresTaskRepository) checkForDuplicateTask(ctx context.Context, task *models.Task, excludeID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		checkTaskDuplicateQuery,
		task.Title,
		task.Description,
		excludeID).Scan(&count)
	if err != nil {
		return false, errors.NewInternal("failed to check for duplicate task", err)
	}
	return count > 0, nil
}

// DeleteTask deletes a task by its ID.
func (r *PostgresTaskRepository) DeleteTask(ctx context.Context, taskId uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, deleteTaskQuery, taskId)
	if err != nil {
		return errors.NewInternal("failed to execute delete query", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternal("failed to retrieve affected rows after delete", err)
	} else if rowsAffected == 0 {
		return errors.NewNotFound("task not found", nil)
	}
	return nil
}

// Выполнение функции в рамках транзакции
func (r *PostgresTaskRepository) WithTransaction(ctx context.Context, f func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	if err := f(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %v", rbErr, err)
		}
		return err
	}
	return tx.Commit()
}

// UpdateTaskStatus updates the status of a task and handles user task completion count.
func (r *PostgresTaskRepository) UpdateTaskStatus(ctx context.Context, taskID string, newStatus int, userID uuid.UUID) (*models.Task, error) {
	var updatedTask models.Task

	err := r.WithTransaction(ctx, func(tx *sql.Tx) error {
		if exists, err := r.checkTaskExistsTx(ctx, tx, taskID); err != nil || !exists {
			if err != nil {
				return err
			}
			return errors.NewNotFound("task not found", nil)
		}

		currentStatus, err := r.getCurrentTaskStatus(ctx, tx, taskID)
		if err != nil {
			return err
		}

		if err := r.validateNewStatus(ctx, tx, newStatus); err != nil {
			return err
		}

		if newStatus == 3 && currentStatus != 3 {
			if err := r.incrementUserTaskCount(ctx, tx, userID); err != nil {
				return err
			}
		}

		if err := r.updateTaskStatusInDB(ctx, tx, newStatus, taskID); err != nil {
			return err
		}

		return r.fetchUpdatedTask(ctx, tx, taskID, &updatedTask)
	})

	if err != nil {
		return nil, err
	}

	return &updatedTask, nil
}

func (r *PostgresTaskRepository) getCurrentTaskStatus(ctx context.Context, tx *sql.Tx, taskID string) (int, error) {
	var currentStatus int
	if err := tx.QueryRowContext(ctx, `SELECT status FROM tasks WHERE task_id = $1`, taskID).Scan(&currentStatus); err != nil {
		return 0, errors.NewInternal("failed to get task status", err)
	}
	return currentStatus, nil
}

func (r *PostgresTaskRepository) validateNewStatus(ctx context.Context, tx *sql.Tx, newStatus int) error {
	var statusExists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM task_status WHERE id = $1)`, newStatus).Scan(&statusExists); err != nil {
		return errors.NewValidation("failed to check if status exists", err)
	}
	if !statusExists {
		return fmt.Errorf("status ID %d does not exist", newStatus)
	}
	return nil
}

func (r *PostgresTaskRepository) incrementUserTaskCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, userTaskStatusChangeQuery, userID); err != nil {
		return errors.NewInternal("failed to update user's completed tasks count", err)
	}
	return nil
}

func (r *PostgresTaskRepository) updateTaskStatusInDB(ctx context.Context, tx *sql.Tx, newStatus int, taskID string) error {
	if _, err := tx.ExecContext(ctx, `UPDATE tasks SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE task_id = $2`, newStatus, taskID); err != nil {
		return errors.NewInternal("failed to update task status", err)
	}
	return nil
}

func (r *PostgresTaskRepository) fetchUpdatedTask(ctx context.Context, tx *sql.Tx, taskID string, updatedTask *models.Task) error {
	return tx.QueryRowContext(ctx, `SELECT task_id, status, title, description, created_at, updated_at, due_date, assignee_id 
	FROM tasks WHERE task_id = $1`, taskID).Scan(&updatedTask.TaskID,
		&updatedTask.Status, &updatedTask.Title, &updatedTask.Description,
		&updatedTask.CreatedAt, &updatedTask.UpdatedAt, &updatedTask.DueDate,
		&updatedTask.AssigneeID)
}

func (r *PostgresTaskRepository) checkTaskExistsTx(ctx context.Context, tx *sql.Tx, taskID string) (bool, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, checkTaskExistsQuery, taskID).Scan(&exists); err != nil {
		return false, errors.NewInternal("failed to check task existence", err)
	}
	return exists, nil
}
