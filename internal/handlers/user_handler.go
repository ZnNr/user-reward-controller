package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-reward-controler/internal/errors"
	"github.com/ZnNr/user-reward-controler/internal/models"
	"github.com/ZnNr/user-reward-controler/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// UserHandler with service interface
type UserHandler struct {
	BaseHandler
	service *service.UserService
}

// NewUserHandler returns a new instance of UserHandler
func NewUserHandler(service *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		BaseHandler: BaseHandler{logger: logger},
		service:     service,
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetUsers request")

	filter := &models.User{} // Задайте фильтр для поиска пользователей

	response, err := h.service.GetUsers(r.Context(), filter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetUserByID request")

	vars := mux.Vars(r)
	id := vars["user_id"]

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling CreateUser request")

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid request body", err))
		return
	}

	user, err := h.service.CreateUser(r.Context(), &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling UpdateUser request")

	vars := mux.Vars(r)
	id := vars["user_id"]

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid request body", err))
		return
	}

	req.UserID = id // Установите UserID в запрос

	updatedUser, err := h.service.UpdateUser(r.Context(), &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, updatedUser)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling DeleteUser request")

	vars := mux.Vars(r)
	id := vars["user_id"]

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetUserByEmail request")

	email := r.URL.Query().Get("email")

	user, err := h.service.GetUserByEmail(r.Context(), email)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling UpdateBalance request")

	vars := mux.Vars(r)
	id := vars["user_id"]

	amountStr := r.URL.Query().Get("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid amount value", err))
		return
	}

	if err := h.service.UpdateBalance(r.Context(), id, amount); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetUserFullInfo(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetUserFullInfo request")

	vars := mux.Vars(r)
	id := vars["user_id"]

	userInfo, err := h.service.GetUserFullInfo(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, userInfo)
}

func (h *UserHandler) GetUserSummary(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetUserSummary request")

	vars := mux.Vars(r)
	id := vars["user_id"]

	summary, err := h.service.GetUserSummary(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, summary)
}

func (h *UserHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling InviteUser request")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var inviteRequest struct {
		InviterID    string `json:"inviter_id"`
		InviteeEmail string `json:"invitee_email"`
	}

	// Декодируем JSON-запрос в структуру
	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ctx := r.Context()
	if err := h.service.InviteUser(ctx, inviteRequest.InviterID, inviteRequest.InviteeEmail); err != nil {
		h.logger.Error("Failed to invite user", zap.Error(err))
		http.Error(w, "Failed to invite user", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User invited successfully"))
}

func (h *UserHandler) GetLeaderByBalance(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetLeaderByBalance request")

	leader, err := h.service.GetLeaderByBalance(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, leader)
}

func (h *UserHandler) GetTopUsers(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetTopUsers request")

	limit, err := getQueryParamInt(r, "limit", 10)
	if err != nil {
		h.handleError(w, err)
		return
	}

	offset, err := getQueryParamInt(r, "offset", 0)
	if err != nil {
		h.handleError(w, err)
		return
	}

	topUsers, err := h.service.GetTopUsers(r.Context(), limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, topUsers)
}
