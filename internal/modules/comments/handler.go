package comments

import (
	"encoding/json"
	"net/http"

	"papertrail/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/papers/{id}/comments", h.list)
	r.Post("/papers/{id}/comments", h.create)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	paperID := chi.URLParam(r, "id")
	comments, err := h.service.ListByPaper(r.Context(), paperID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, comments)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	paperID := chi.URLParam(r, "id")

	var req struct {
		UserID string `json:"user_id"`
		Body   string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	comment, err := h.service.Create(r.Context(), Comment{
		PaperID: paperID,
		UserID:  req.UserID,
		Body:    req.Body,
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusCreated, comment)
}
