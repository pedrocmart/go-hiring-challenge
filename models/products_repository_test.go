package models

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProductsRepository_List_And_GetByCode(t *testing.T) {
	db := newTestDB(t)
	repo := NewProductsRepository(db)
	ctx := context.Background()

	// Seed: 2 categories
	catA := Category{Code: "ELEC", Name: "Electronics"}
	catB := Category{Code: "BOOK", Name: "Books"}
	if err := db.Create(&catA).Error; err != nil {
		t.Fatalf("seed catA: %v", err)
	}
	if err := db.Create(&catB).Error; err != nil {
		t.Fatalf("seed catB: %v", err)
	}

	// Seed: 3 products (2 in ELEC, 1 in BOOK) with variants
	p1 := Product{Code: "P-001", Price: decimal.NewFromFloat(199.99), CategoryID: catA.ID}
	p2 := Product{Code: "P-002", Price: decimal.NewFromFloat(49.50), CategoryID: catA.ID}
	p3 := Product{Code: "P-003", Price: decimal.NewFromFloat(9.99), CategoryID: catB.ID}
	if err := db.Create(&p1).Error; err != nil {
		t.Fatalf("seed p1: %v", err)
	}
	if err := db.Create(&p2).Error; err != nil {
		t.Fatalf("seed p2: %v", err)
	}
	if err := db.Create(&p3).Error; err != nil {
		t.Fatalf("seed p3: %v", err)
	}

	// Variants
	v := []Variant{
		{ProductID: p1.ID, Name: "Black", SKU: "SKU-001-B"},
		{ProductID: p1.ID, Name: "White", SKU: "SKU-001-W"},
		{ProductID: p3.ID, Name: "Paperback", SKU: "SKU-003-P"},
	}
	if err := db.Create(&v).Error; err != nil {
		t.Fatalf("seed variants: %v", err)
	}

	// 1) No filters: should see all 3, total=3
	{
		products, total, err := repo.List(ctx, 0, 50, "", nil)
		if err != nil {
			t.Fatalf("List no filters: %v", err)
		}
		if total != 3 {
			t.Fatalf("expected total=3, got %d", total)
		}
		if len(products) != 3 {
			t.Fatalf("expected 3 products, got %d", len(products))
		}
		// Preloads present
		for _, p := range products {
			if p.Category.ID == 0 {
				t.Fatalf("expected Category to be preloaded for %s", p.Code)
			}
		}
	}

	// 2) Category filter: ELEC -> 2 items
	{
		products, total, err := repo.List(ctx, 0, 50, "ELEC", nil)
		if err != nil {
			t.Fatalf("List category ELEC: %v", err)
		}
		if total != 2 {
			t.Fatalf("expected total=2 for ELEC, got %d", total)
		}
		if len(products) != 2 {
			t.Fatalf("expected 2 products, got %d", len(products))
		}
		for _, p := range products {
			if p.Category.Code != "ELEC" {
				t.Fatalf("expected category ELEC, got %s", p.Category.Code)
			}
		}
	}

	// 3) Price filter: price < 50.00 -> P-002 (49.50) and P-003 (9.99) => 2 items
	{
		pl := decimal.NewFromFloat(50.00)
		products, total, err := repo.List(ctx, 0, 50, "", &pl)
		if err != nil {
			t.Fatalf("List price < 50: %v", err)
		}
		if total != 2 {
			t.Fatalf("expected total=2 for price<50, got %d", total)
		}
		if len(products) != 2 {
			t.Fatalf("expected 2 products, got %d", len(products))
		}
	}

	// 4) Combined filters: category ELEC and price < 100 -> only P-002 => 1 item
	{
		pl := decimal.NewFromFloat(100.00)
		products, total, err := repo.List(ctx, 0, 50, "ELEC", &pl)
		if err != nil {
			t.Fatalf("List ELEC & price<100: %v", err)
		}
		if total != 1 {
			t.Fatalf("expected total=1, got %d", total)
		}
		if len(products) != 1 || products[0].Code != "P-002" {
			t.Fatalf("expected only P-002, got %+v", products)
		}
		// Variants preloaded for returned rows
		if len(products[0].Variants) != 0 {
			t.Fatalf("expected Variants to be preloaded for P-002 (even if none exist, preload returns empty slice)")
		}
	}

	// 5) Pagination: offset=1, limit=1, no filters -> returns 1 row, total stays 3
	{
		products, total, err := repo.List(ctx, 1, 1, "", nil)
		if err != nil {
			t.Fatalf("List pagination: %v", err)
		}
		if total != 3 {
			t.Fatalf("expected total=3 with pagination, got %d", total)
		}
		if len(products) != 1 {
			t.Fatalf("expected 1 product page size, got %d", len(products))
		}
	}

	// 6) GetByCode success + preloads
	{
		got, err := repo.GetByCode(ctx, "P-001")
		if err != nil {
			t.Fatalf("GetByCode P-001: %v", err)
		}
		if got.Code != "P-001" {
			t.Fatalf("expected P-001, got %s", got.Code)
		}
		if got.Category.Code != "ELEC" {
			t.Fatalf("expected category ELEC, got %s", got.Category.Code)
		}
		if len(got.Variants) != 2 {
			t.Fatalf("expected 2 variants for P-001, got %d", len(got.Variants))
		}
	}

	// 7) GetByCode not found
	{
		_, err := repo.GetByCode(ctx, "DOES-NOT-EXIST")
		if err == nil {
			t.Fatalf("expected error for missing product")
		}
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Category{}, &Product{}, &Variant{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}
