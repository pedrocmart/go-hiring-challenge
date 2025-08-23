package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIsUniqueConstraintError(t *testing.T) {
	t.Run("returns true for 23505", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505"}
		if !isUniqueConstraintError(err) {
			t.Fatalf("expected true for 23505, got false")
		}
	})

	t.Run("returns false for non-23505", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"} // foreign key violation
		if isUniqueConstraintError(err) {
			t.Fatalf("expected false for 23503, got true")
		}
	})

	t.Run("unwraps wrapped PgError", func(t *testing.T) {
		base := &pgconn.PgError{Code: "23505"}
		wrapped := fmt.Errorf("insert failed: %w", base)
		if !isUniqueConstraintError(wrapped) {
			t.Fatalf("expected true for wrapped 23505, got false")
		}
	})
}

func TestCategoriesRepository_CreateAndList_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&Category{}); err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}

	repo := NewCategoriesRepository(db)

	ctx := context.Background()
	cat := &Category{
		Code: "ELECT",
		Name: "Electronics",
	}

	// Create should succeed.
	if err := repo.Create(ctx, cat); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List should return the inserted category.
	got, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 category, got %d", len(got))
	}
	if got[0].Code != "ELECT" || got[0].Name != "Electronics" {
		t.Fatalf("unexpected category: %+v", got[0])
	}
}
