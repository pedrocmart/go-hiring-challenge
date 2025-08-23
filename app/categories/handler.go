package categories

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// Repository defines the required behaviour for category storage.
type Repository interface {
	List(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, c *models.Category) error
}

type Handler struct {
	repo Repository
}

func NewHandler(r Repository) *Handler {
	return &Handler{repo: r}
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.List(r.Context())
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]categoryResponse, len(categories))
	for i, c := range categories {
		resp[i] = categoryResponse{Code: c.Code, Name: c.Name}
	}

	api.OKResponse(w, resp)
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req categoryResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid body")
		return
	}

	if strings.TrimSpace(req.Code) == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code is required")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "name is required")
		return
	}

	cat := models.Category{Code: req.Code, Name: req.Name}
	if err := h.repo.Create(r.Context(), &cat); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.OKResponse(w, categoryResponse{Code: cat.Code, Name: cat.Name})
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
