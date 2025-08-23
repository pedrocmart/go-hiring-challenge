package catalog

import (
	"context"
	"net/http"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type ProductRepository interface {
	List(ctx context.Context, offset, limit int, category string, priceLessThan *decimal.Decimal) ([]models.Product, int64, error)
	GetByCode(ctx context.Context, code string) (*models.Product, error)
}

type CatalogHandler struct {
	repo ProductRepository
}

func NewCatalogHandler(r ProductRepository) *CatalogHandler {
	return &CatalogHandler{repo: r}
}

func (h *CatalogHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	limit := 10
	if l, err := strconv.Atoi(q.Get("limit")); err == nil {
		limit = l
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	category := q.Get("category")

	var priceFilter *decimal.Decimal
	if v := q.Get("priceLessThan"); v != "" {
		p, err := decimal.NewFromString(v)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid priceLessThan")
			return
		}
		priceFilter = &p
	}

	products, total, err := h.repo.List(r.Context(), offset, limit, category, priceFilter)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := struct {
		Products []productResponse `json:"products"`
		Total    int64             `json:"total"`
	}{
		Products: make([]productResponse, len(products)),
		Total:    total,
	}

	for i, p := range products {
		resp.Products[i] = mapProduct(p)
	}

	api.OKResponse(w, resp)
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	p, err := h.repo.GetByCode(r.Context(), code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "product not found")
		return
	}

	pr := mapProduct(*p)
	pr.Variants = make([]variantResponse, len(p.Variants))
	for i, v := range p.Variants {
		price := v.Price
		if price.IsZero() {
			price = p.Price
		}
		pr.Variants[i] = variantResponse{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}

	api.OKResponse(w, pr)
}

type productResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category categoryResponse  `json:"category"`
	Variants []variantResponse `json:"variants,omitempty"`
}

type variantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func mapProduct(p models.Product) productResponse {
	return productResponse{
		Code:  p.Code,
		Price: p.Price.InexactFloat64(),
		Category: categoryResponse{
			Code: p.Category.Code,
			Name: p.Category.Name,
		},
	}
}
