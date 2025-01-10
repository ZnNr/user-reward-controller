package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-reward-controler/internal/errors"
	"github.com/ZnNr/user-reward-controler/internal/models"
	"github.com/ZnNr/user-reward-controler/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

// ReferralHandler with service interface
type ReferralHandler struct {
	BaseHandler
	service *service.ReferralService
}

// NewReferralHandler returns a new instance of ReferralHandler
func NewReferralHandler(service *service.ReferralService, logger *zap.Logger) *ReferralHandler {
	return &ReferralHandler{
		BaseHandler: BaseHandler{logger: logger},
		service:     service,
	}
}

// CreateReferral handles creation of a referral code
func (h *ReferralHandler) CreateReferral(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling CreateReferral request")

	var req models.CreateReferralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid request body", err))
		return
	}

	referral, err := h.service.CreateReferral(req.UserID, req.Code)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, referral)
}

// GetReferral handles getting a referral code by ID
func (h *ReferralHandler) GetReferral(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetReferral request")

	vars := mux.Vars(r)
	idStr := vars["referral_id"]
	referralID, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid referral ID", err))
		return
	}

	referral, err := h.service.GetReferral(referralID.String())
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, referral)
}

// UpdateReferral handles updating an existing referral code
func (h *ReferralHandler) UpdateReferral(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling UpdateReferral request")

	vars := mux.Vars(r)
	idStr := vars["referral_id"]
	referralID, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid referral ID", err))
		return
	}

	var req models.UpdateReferralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, errors.NewBadRequest("Invalid request body", err))
		return
	}

	referral, err := h.service.UpdateReferral(referralID.String(), req.Code)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, referral)
}

// DeleteReferral handles deletion of a referral code by ID
func (h *ReferralHandler) DeleteReferral(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling DeleteReferral request")

	vars := mux.Vars(r)
	idStr := vars["referral_id"]

	if err := h.service.DeleteReferral(idStr); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReferralsByUserID handles fetching all referrals for a specific user
func (h *ReferralHandler) GetReferralsByUserID(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling GetReferralsByUserID request")

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.handleError(w, errors.NewBadRequest("User ID is required", nil))
		return
	}

	referrals, err := h.service.GetReferralsByUserID(userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, referrals)
}
