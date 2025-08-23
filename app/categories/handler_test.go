package categories

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mytheresa/go-hiring-challenge/models"
)

type stubRepo struct {
	listFunc   func(ctx context.Context) ([]models.Category, error)
	createFunc func(ctx context.Context, c *models.Category) error
}

func (s *stubRepo) List(ctx context.Context) ([]models.Category, error) {
	return s.listFunc(ctx)
}

func (s *stubRepo) Create(ctx context.Context, c *models.Category) error {
	return s.createFunc(ctx, c)
}

func TestHandleList(t *testing.T) {
	repo := &stubRepo{
		listFunc: func(ctx context.Context) ([]models.Category, error) {
			return []models.Category{{Code: "CLOTHING", Name: "Clothing"}}, nil
		},
	}

	h := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expected := `[{"code":"CLOTHING","name":"Clothing"}]`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestHandleCreate(t *testing.T) {
	var created models.Category
	repo := &stubRepo{
		createFunc: func(ctx context.Context, c *models.Category) error {
			created = *c
			return nil
		},
		listFunc: nil,
	}

	h := NewHandler(repo)

	body := bytes.NewBufferString(`{"code":"NEW","name":"New Cat"}`)
	req := httptest.NewRequest(http.MethodPost, "/categories", body)
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "NEW", created.Code)
	assert.Equal(t, "New Cat", created.Name)
	expected := `{"code":"NEW","name":"New Cat"}`
	assert.JSONEq(t, expected, w.Body.String())
}
