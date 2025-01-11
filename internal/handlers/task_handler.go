package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-reward-controller/internal/errors"
	"github.com/ZnNr/user-reward-controller/internal/models"
	"github.com/ZnNr/user-reward-controller/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// TaskHandler with service interface
type TaskHandler struct {
	BaseHandler
	service *service.TaskService
}

// NewTaskHandler returns a new instance of TaskHandler
func NewTaskHandler(service *service.TaskService, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		BaseHandler: BaseHandler{logger: logger},
		service:     service,
	}
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetTaskByID request")

	vars := mux.Vars(r)
	idStr := vars["task_id"]
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid task ID", err))
		return
	}

	task, err := h.service.GetTaskByID(r.Context(), taskID.String())
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetTasks request")

	// Создание фильтра задач
	filter := &models.TaskFilter{
		Title:       r.URL.Query().Get("title"),
		Status:      r.URL.Query().Get("status"),
		Description: r.URL.Query().Get("description"),
		AssigneeID:  r.URL.Query().Get("assignee_id"),
	}

	// Получение параметров страницы и размера страницы
	if err := h.getPaginationParams(r, filter); err != nil {
		h.handleError(w, err)
		return
	}

	// Получение параметров даты
	if err := h.getDateParams(r, filter); err != nil {
		h.handleError(w, err)
		return
	}

	// Получение задач из сервиса
	response, err := h.service.GetTasks(r.Context(), filter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// Получение параметров страницы и размера страницы
func (h *TaskHandler) getPaginationParams(r *http.Request, filter *models.TaskFilter) error {
	var err error
	if filter.Page, err = getQueryParamInt(r, "page", 0); err != nil {
		return errors.NewBadRequest("Invalid page number", err)
	}
	if filter.PageSize, err = getQueryParamInt(r, "page_size", 10); err != nil {
		return errors.NewBadRequest("Invalid page size", err)
	}
	return nil
}

// Получение параметров дат
func (h *TaskHandler) getDateParams(r *http.Request, filter *models.TaskFilter) error {
	var err error
	if filter.CreatedAfter, err = getQueryParamDate(r, "createdAfter"); err != nil {
		return errors.NewBadRequest("Invalid createdAfter format", err)
	}
	if filter.CreatedBefore, err = getQueryParamDate(r, "createdBefore"); err != nil {
		return errors.NewBadRequest("Invalid createdBefore format", err)
	}
	if filter.DueAfter, err = getQueryParamDate(r, "dueAfter"); err != nil {
		return errors.NewBadRequest("Invalid dueAfter format", err)
	}
	if filter.DueBefore, err = getQueryParamDate(r, "dueBefore"); err != nil {
		return errors.NewBadRequest("Invalid dueBefore format", err)
	}
	return nil
}

// Утилитные функции для получения значений параметров
func getQueryParamInt(r *http.Request, param string, defaultValue int) (int, error) {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(valueStr)
}

func getQueryParamDate(r *http.Request, param string) (*time.Time, error) {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return nil, nil // Возвращаем nil, если параметр отсутствует
	}
	t, err := time.Parse(time.RFC3339, valueStr) // Модифицируйте под ваш формат даты
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling CreateTask request")

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid request body", err))
		return
	}

	task, err := h.service.CreateTask(r.Context(), &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling UpdateTask request")

	vars := mux.Vars(r)
	idStr := vars["task_id"]
	taskId, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid task ID", err))
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid request body", err))
		return
	}

	task, err := h.service.UpdateTask(r.Context(), taskId, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling UpdateTaskStatus request")

	vars := mux.Vars(r)
	idStr := vars["task_id"]     // Получаем id задачи
	userIdStr := vars["user_id"] // Получаем id пользователя

	taskId, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid task ID", err))
		return
	}

	userID, err := uuid.Parse(userIdStr) // Конвертируем строку userId в uuid
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid user ID", err))
		return
	}

	newStatusStr := r.URL.Query().Get("status")
	newStatus, err := strconv.Atoi(newStatusStr)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid status value", err))
		return
	}

	updatedTask, err := h.service.UpdateTaskStatus(r.Context(), taskId.String(), newStatus, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, updatedTask)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling DeleteTask request")

	vars := mux.Vars(r)
	idStr := vars["task_id"]

	if err := h.service.DeleteTask(r.Context(), idStr); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) GetDescription(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetDescription request")

	vars := mux.Vars(r)
	idStr := vars["task_id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid task ID", err))
		return
	}

	page, pageSize := 0, 10
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil {
			page = parsedPage
		} else {
			h.handleError(w, errors.NewBadRequest("Invalid page query", err))
			return
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = parsedPageSize
		} else {
			h.handleError(w, errors.NewBadRequest("Invalid page size query", err))
			return
		}
	}

	response, err := h.service.GetDescription(r.Context(), id, page, pageSize)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}
