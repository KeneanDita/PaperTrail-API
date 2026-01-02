package users

import (
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
	r.Get("/users", h.list)
	r.Get("/users/{id}", h.get)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.service.Get(r.Context(), id)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.List(r.Context())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, users)
}
