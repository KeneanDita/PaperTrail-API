package papers

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
	r.Post("/papers", h.create)
	r.Get("/papers", h.list)
	r.Get("/papers/{id}", h.get)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		Abstract string `json:"abstract"`
		AuthorID string `json:"author_id"`
		PdfURL   string `json:"pdf_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	created, err := h.service.Create(r.Context(), Paper{
		Title:    req.Title,
		Abstract: req.Abstract,
		AuthorID: req.AuthorID,
		PdfURL:   req.PdfURL,
	})
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusCreated, created)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	paper, err := h.service.Get(r.Context(), id)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, paper)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	papers, err := h.service.List(r.Context())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, papers)
}
