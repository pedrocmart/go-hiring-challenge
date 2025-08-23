package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/mytheresa/go-hiring-challenge/models"
)

type stubProductRepo struct {
	listFunc      func(ctx context.Context, offset, limit int, category string, priceLessThan *decimal.Decimal) ([]models.Product, int64, error)
	getByCodeFunc func(ctx context.Context, code string) (*models.Product, error)
}

func (s *stubProductRepo) List(ctx context.Context, offset, limit int, category string, priceLessThan *decimal.Decimal) ([]models.Product, int64, error) {
	return s.listFunc(ctx, offset, limit, category, priceLessThan)
}

func (s *stubProductRepo) GetByCode(ctx context.Context, code string) (*models.Product, error) {
	return s.getByCodeFunc(ctx, code)
}

func TestHandleList(t *testing.T) {
	repo := &stubProductRepo{
		listFunc: func(ctx context.Context, offset, limit int, category string, priceLessThan *decimal.Decimal) ([]models.Product, int64, error) {
			assert.Equal(t, 0, offset)
			assert.Equal(t, 10, limit)
			assert.Equal(t, "", category)
			assert.Nil(t, priceLessThan)
			return []models.Product{
				{
					Code:     "PROD001",
					Price:    decimal.NewFromFloat(10.5),
					Category: models.Category{Code: "CLOTHING", Name: "Clothing"},
				},
				{
					Code:     "PROD002",
					Price:    decimal.NewFromFloat(12),
					Category: models.Category{Code: "SHOES", Name: "Shoes"},
				},
			}, 2, nil
		},
	}

	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expected := `{"products":[{"code":"PROD001","price":10.5,"category":{"code":"CLOTHING","name":"Clothing"}},{"code":"PROD002","price":12,"category":{"code":"SHOES","name":"Shoes"}}],"total":2}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestHandleListWithFilters(t *testing.T) {
	var gotOffset, gotLimit int
	var gotCategory string
	var gotPrice *decimal.Decimal

	repo := &stubProductRepo{
		listFunc: func(ctx context.Context, offset, limit int, category string, priceLessThan *decimal.Decimal) ([]models.Product, int64, error) {
			gotOffset, gotLimit, gotCategory, gotPrice = offset, limit, category, priceLessThan
			return []models.Product{}, 0, nil
		},
	}

	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog?offset=2&limit=5&category=CLOTHING&priceLessThan=20", nil)
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 2, gotOffset)
	assert.Equal(t, 5, gotLimit)
	assert.Equal(t, "CLOTHING", gotCategory)
	if assert.NotNil(t, gotPrice) {
		assert.True(t, gotPrice.Equal(decimal.NewFromInt(20)))
	}
}

func TestHandleGet(t *testing.T) {
	repo := &stubProductRepo{
		getByCodeFunc: func(ctx context.Context, code string) (*models.Product, error) {
			assert.Equal(t, "PROD001", code)
			return &models.Product{
				Code:     "PROD001",
				Price:    decimal.NewFromFloat(10),
				Category: models.Category{Code: "CLOTHING", Name: "Clothing"},
				Variants: []models.Variant{{Name: "Variant A", SKU: "SKU001", Price: decimal.Zero}},
			}, nil
		},
	}

	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
	req.SetPathValue("code", "PROD001")
	w := httptest.NewRecorder()
	h.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expected := `{"code":"PROD001","price":10,"category":{"code":"CLOTHING","name":"Clothing"},"variants":[{"name":"Variant A","sku":"SKU001","price":10}]}`
	assert.JSONEq(t, expected, w.Body.String())
}
