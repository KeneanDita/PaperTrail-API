package reviews

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
	r.Get("/papers/{id}/reviews", h.list)
	r.Post("/papers/{id}/reviews", h.create)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	paperID := chi.URLParam(r, "id")
	reviews, err := h.service.ListByPaper(r.Context(), paperID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, reviews)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	paperID := chi.URLParam(r, "id")

	var req struct {
		ReviewerID string `json:"reviewer_id"`
		Rating     int    `json:"rating"`
		Comments   string `json:"comments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	review, err := h.service.Create(r.Context(), Review{
		PaperID:    paperID,
		ReviewerID: req.ReviewerID,
		Rating:     req.Rating,
		Comments:   req.Comments,
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusCreated, review)
}
