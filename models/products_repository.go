package models

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductsRepository provides access to product storage.
type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

// List returns products matching the provided filters along with the total count.
func (r *ProductsRepository) List(ctx context.Context, offset, limit int, category string, priceLessThan *decimal.Decimal) ([]Product, int64, error) {
	applyFilters := func(db *gorm.DB) *gorm.DB {
		if category != "" {
			db = db.Joins("JOIN categories ON categories.id = products.category_id").Where("categories.code = ?", category)
		}
		if priceLessThan != nil {
			db = db.Where("price < ?", *priceLessThan)
		}
		return db
	}

	base := applyFilters(r.db.WithContext(ctx).Model(&Product{}))

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := applyFilters(r.db.WithContext(ctx).Model(&Product{})).
		Order("products.id ASC").
		Preload("Variants").Preload("Category").
		Offset(offset).Limit(limit)
	var products []Product
	if err := query.Preload("Variants").Preload("Category").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetByCode returns a product identified by its code.
func (r *ProductsRepository) GetByCode(ctx context.Context, code string) (*Product, error) {
	var p Product
	if err := r.db.WithContext(ctx).Preload("Variants").Preload("Category").Where("code = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
