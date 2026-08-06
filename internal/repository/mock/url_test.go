package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/mmk31585/url-shortener/internal/domain"
)

func TestURLRepository_Create(t *testing.T) {
	repo := NewURLRepository()
	created, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   "abc11111",
		OriginalURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("ID: got %d, want 1", created.ID)
	}
	if created.ShortCode != "abc11111" {
		t.Errorf("ShortCode: got %q", created.ShortCode)
	}
	if created.RedirectCount != 0 {
		t.Errorf("RedirectCount: got %d, want 0", created.RedirectCount)
	}
	if created.CreatedAt.IsZero() {
		t.Error("CreatedAt must be set")
	}
}

func TestURLRepository_CreateCollision(t *testing.T) {
	repo := NewURLRepository()
	_, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://a.com"})
	if err != nil {
		t.Fatalf("first Create() failed: %v", err)
	}
	_, err = repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://b.com"})
	if !errors.Is(err, domain.ErrShortCodeCollision) {
		t.Errorf("expected ErrShortCodeCollision, got %v", err)
	}
}

func TestURLRepository_GetByShortCode(t *testing.T) {
	repo := NewURLRepository()
	created, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://a.com"})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := repo.GetByShortCode(context.Background(), "abc11111")
	if err != nil {
		t.Fatalf("GetByShortCode() returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID: got %d, want %d", got.ID, created.ID)
	}
}

func TestURLRepository_GetByShortCodeNotFound(t *testing.T) {
	repo := NewURLRepository()
	_, err := repo.GetByShortCode(context.Background(), "missing1")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestURLRepository_GetByShortCodeExcludesDeleted(t *testing.T) {
	repo := NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://a.com"}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if _, err := repo.SoftDelete(context.Background(), "abc11111"); err != nil {
		t.Fatalf("SoftDelete() failed: %v", err)
	}
	_, err := repo.GetByShortCode(context.Background(), "abc11111")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("deleted record should not be found, got %v", err)
	}
}

func TestURLRepository_GetAll(t *testing.T) {
	repo := NewURLRepository()
	for _, code := range []string{"abc11111", "abc22222"} {
		if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: domain.ShortCode(code), OriginalURL: "https://x.com"}); err != nil {
			t.Fatalf("Create() returned error: %v", err)
		}
	}

	urls, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() returned error: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("got %d urls, want 2", len(urls))
	}
}

func TestURLRepository_GetAllEmpty(t *testing.T) {
	repo := NewURLRepository()
	urls, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() returned error: %v", err)
	}
	if urls == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(urls) != 0 {
		t.Errorf("got %d urls, want 0", len(urls))
	}
}

func TestURLRepository_GetAllExcludesDeleted(t *testing.T) {
	repo := NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://a.com"}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc22222", OriginalURL: "https://b.com"}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if _, err := repo.SoftDelete(context.Background(), "abc11111"); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	urls, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Errorf("got %d urls, want 1 (deleted excluded)", len(urls))
	}
	if urls[0].ShortCode != "abc22222" {
		t.Errorf("got %q, want abc22222", urls[0].ShortCode)
	}
}

func TestURLRepository_GetByOriginalURL(t *testing.T) {
	repo := NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://unique.com"}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := repo.GetByOriginalURL(context.Background(), "https://unique.com")
	if err != nil {
		t.Fatalf("GetByOriginalURL() returned error: %v", err)
	}
	if got.ShortCode != "abc11111" {
		t.Errorf("ShortCode: got %q", got.ShortCode)
	}
}

func TestURLRepository_GetByOriginalURLNotFound(t *testing.T) {
	repo := NewURLRepository()
	_, err := repo.GetByOriginalURL(context.Background(), "https://nope.com")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestURLRepository_Update(t *testing.T) {
	repo := NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://old.com"}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	updated, err := repo.Update(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://new.com"})
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}
	if updated.OriginalURL != "https://new.com" {
		t.Errorf("OriginalURL: got %q", updated.OriginalURL)
	}
}

func TestURLRepository_UpdateNotFound(t *testing.T) {
	repo := NewURLRepository()
	_, err := repo.Update(context.Background(), &domain.URL{ShortCode: "missing1", OriginalURL: "https://x.com"})
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestURLRepository_SoftDelete(t *testing.T) {
	repo := NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://x.com"}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	deleted, err := repo.SoftDelete(context.Background(), "abc11111")
	if err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}
	if deleted.DeletedAt == nil {
		t.Error("DeletedAt must be set")
	}
}

func TestURLRepository_SoftDeleteNotFound(t *testing.T) {
	repo := NewURLRepository()
	_, err := repo.SoftDelete(context.Background(), "missing1")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestURLRepository_IncrementRedirectCount(t *testing.T) {
	repo := NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{ShortCode: "abc11111", OriginalURL: "https://x.com"}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := repo.IncrementRedirectCount(context.Background(), "abc11111")
	if err != nil {
		t.Fatalf("IncrementRedirectCount() returned error: %v", err)
	}
	if got.RedirectCount != 1 {
		t.Errorf("RedirectCount: got %d, want 1", got.RedirectCount)
	}

	got, err = repo.IncrementRedirectCount(context.Background(), "abc11111")
	if err != nil {
		t.Fatalf("second IncrementRedirectCount() failed: %v", err)
	}
	if got.RedirectCount != 2 {
		t.Errorf("RedirectCount: got %d, want 2", got.RedirectCount)
	}
}

func TestURLRepository_IncrementRedirectCountNotFound(t *testing.T) {
	repo := NewURLRepository()
	_, err := repo.IncrementRedirectCount(context.Background(), "missing1")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}
