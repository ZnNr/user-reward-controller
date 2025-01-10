package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-reward-controler/internal/errors"
	"go.uber.org/zap"
	"net/http"
)

// BaseHandler provides common functionality for HTTP handlers
type BaseHandler struct {
	logger *zap.Logger
}

// handleError handles error responses
func (h *BaseHandler) handleError(w http.ResponseWriter, err error) {
	h.logger.Error("Handling error", zap.Error(err))
	var status int
	var errorResponse map[string]string

	if errors.IsNotFound(err) {
		status = http.StatusNotFound
		errorResponse = map[string]string{"error": err.Error()}
	} else if errors.IsBadRequest(err) {
		status = http.StatusBadRequest
		errorResponse = map[string]string{"error": err.Error()}
	} else {
		status = http.StatusInternalServerError
		errorResponse = map[string]string{"error": "internal server error"}
	}

	h.respondWithJSON(w, status, errorResponse)
}

// respondWithJSON writes a JSON response to the ResponseWriter
func (h *BaseHandler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))
	}
}
