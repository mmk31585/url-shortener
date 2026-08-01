package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mmk31585/url-shortener/internal/domain"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	addr := os.Getenv("DB_ADDR")
	if addr == "" {
		addr = "postgres://admin:admin@localhost:5430/urlshortener?sslmode=disable"
	}

	db, err := sql.Open("pgx", addr)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("skipping integration tests: postgres not available: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func cleanTable(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), "DELETE FROM urls")
	if err != nil {
		t.Fatalf("failed to clean table: %v", err)
	}
}

func TestPostgresURLRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	url := &domain.URL{
		ShortCode:   "abc123de",
		OriginalURL: "https://example.com",
	}

	created, err := repo.Create(context.Background(), url)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if created.ShortCode != "abc123de" {
		t.Errorf("ShortCode: got %q, want %q", created.ShortCode, "abc123de")
	}
	if created.OriginalURL != "https://example.com" {
		t.Errorf("OriginalURL: got %q, want %q", created.OriginalURL, "https://example.com")
	}
	if created.RedirectCount != 0 {
		t.Errorf("RedirectCount: got %d, want 0", created.RedirectCount)
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
	if created.DeletedAt != nil {
		t.Error("expected nil DeletedAt for new URL")
	}
}

func TestPostgresURLRepository_Create_DuplicateShortCode(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	url := &domain.URL{
		ShortCode:   "dup12345",
		OriginalURL: "https://example.com",
	}

	_, err := repo.Create(context.Background(), url)
	if err != nil {
		t.Fatalf("first Create() failed: %v", err)
	}

	duplicate := &domain.URL{
		ShortCode:   "dup12345",
		OriginalURL: "https://other.com",
	}

	_, err = repo.Create(context.Background(), duplicate)
	if err == nil {
		t.Fatal("expected error for duplicate short_code, got nil")
	}
	if !errors.Is(err, domain.ErrShortCodeCollision) {
		t.Errorf("expected ErrShortCodeCollision, got %v", err)
	}
}

func TestPostgresURLRepository_GetByShortCode(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	url := &domain.URL{
		ShortCode:   "gettest1",
		OriginalURL: "https://example.com/get",
	}
	created, err := repo.Create(context.Background(), url)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	got, err := repo.GetByShortCode(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("GetByShortCode() returned error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID: got %d, want %d", got.ID, created.ID)
	}
	if got.ShortCode != created.ShortCode {
		t.Errorf("ShortCode: got %q, want %q", got.ShortCode, created.ShortCode)
	}
	if got.OriginalURL != created.OriginalURL {
		t.Errorf("OriginalURL: got %q, want %q", got.OriginalURL, created.OriginalURL)
	}
}

func TestPostgresURLRepository_GetByShortCode_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	_, err := repo.GetByShortCode(context.Background(), "nonexist")
	if err == nil {
		t.Fatal("expected error for non-existent shortcode, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestPostgresURLRepository_GetByShortCode_ExcludesDeleted(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	url := &domain.URL{
		ShortCode:   "delcode1",
		OriginalURL: "https://example.com/del",
	}
	created, err := repo.Create(context.Background(), url)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	_, err = repo.SoftDelete(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("SoftDelete() failed: %v", err)
	}

	_, err = repo.GetByShortCode(context.Background(), created.ShortCode)
	if err == nil {
		t.Fatal("expected error for deleted URL, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound for deleted URL, got %v", err)
	}
}

func TestPostgresURLRepository_GetAll(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	urls := []*domain.URL{
		{ShortCode: "all001aa", OriginalURL: "https://example.com/1"},
		{ShortCode: "all002bb", OriginalURL: "https://example.com/2"},
	}
	for _, u := range urls {
		if _, err := repo.Create(context.Background(), u); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}

	all, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() returned error: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(all))
	}
}

func TestPostgresURLRepository_GetAll_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	all, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() returned error: %v", err)
	}

	if all == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(all) != 0 {
		t.Errorf("expected 0 URLs, got %d", len(all))
	}
}

func TestPostgresURLRepository_GetAll_ExcludesDeleted(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	url := &domain.URL{
		ShortCode:   "alldel01",
		OriginalURL: "https://example.com/del",
	}
	created, err := repo.Create(context.Background(), url)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	_, err = repo.SoftDelete(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("SoftDelete() failed: %v", err)
	}

	all, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() returned error: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("expected 0 URLs after soft delete, got %d", len(all))
	}
}

func TestPostgresURLRepository_GetByOriginalURL(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	original := "https://example.com/original"
	url := &domain.URL{
		ShortCode:   "orig001a",
		OriginalURL: original,
	}
	created, err := repo.Create(context.Background(), url)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	got, err := repo.GetByOriginalURL(context.Background(), original)
	if err != nil {
		t.Fatalf("GetByOriginalURL() returned error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID: got %d, want %d", got.ID, created.ID)
	}
}

func TestPostgresURLRepository_GetByOriginalURL_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	_, err := repo.GetByOriginalURL(context.Background(), "https://nonexistent.com")
	if err == nil {
		t.Fatal("expected error for non-existent URL, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestPostgresURLRepository_Update(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	created, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   "upd001aa",
		OriginalURL: "https://example.com/old",
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	updated, err := repo.Update(context.Background(), &domain.URL{
		ShortCode:   created.ShortCode,
		OriginalURL: "https://example.com/new",
	})
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	if updated.OriginalURL != "https://example.com/new" {
		t.Errorf("OriginalURL: got %q, want %q", updated.OriginalURL, "https://example.com/new")
	}
	if updated.ShortCode != created.ShortCode {
		t.Errorf("ShortCode changed: got %q, want %q", updated.ShortCode, created.ShortCode)
	}
	if updated.CreatedAt != created.CreatedAt {
		t.Errorf("CreatedAt changed: got %v, want %v", updated.CreatedAt, created.CreatedAt)
	}
	if updated.UpdatedAt.Equal(created.UpdatedAt) {
		t.Error("expected UpdatedAt to change after update")
	}
}

func TestPostgresURLRepository_Update_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	_, err := repo.Update(context.Background(), &domain.URL{
		ShortCode:   "nonexist",
		OriginalURL: "https://example.com",
	})
	if err == nil {
		t.Fatal("expected error for non-existent shortcode, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestPostgresURLRepository_Update_ExcludesDeleted(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	created, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   "upddel01",
		OriginalURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	_, err = repo.SoftDelete(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("SoftDelete() failed: %v", err)
	}

	_, err = repo.Update(context.Background(), &domain.URL{
		ShortCode:   created.ShortCode,
		OriginalURL: "https://example.com/updated",
	})
	if err == nil {
		t.Fatal("expected error when updating deleted URL, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound for deleted URL, got %v", err)
	}
}

func TestPostgresURLRepository_SoftDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	created, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   "del002aa",
		OriginalURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	deleted, err := repo.SoftDelete(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	if deleted.DeletedAt == nil {
		t.Fatal("expected DeletedAt to be set")
	}
}

func TestPostgresURLRepository_SoftDelete_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	_, err := repo.SoftDelete(context.Background(), "nonexist")
	if err == nil {
		t.Fatal("expected error for non-existent shortcode, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestPostgresURLRepository_SoftDelete_AlreadyDeleted(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	created, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   "deldel01",
		OriginalURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	_, err = repo.SoftDelete(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("first SoftDelete() failed: %v", err)
	}

	_, err = repo.SoftDelete(context.Background(), created.ShortCode)
	if err == nil {
		t.Fatal("expected error when deleting already deleted URL, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound for already deleted URL, got %v", err)
	}
}

func TestPostgresURLRepository_IncrementRedirectCount(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	created, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   "inc001aa",
		OriginalURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	updated, err := repo.IncrementRedirectCount(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("IncrementRedirectCount() returned error: %v", err)
	}

	if updated.RedirectCount != 1 {
		t.Errorf("RedirectCount: got %d, want 1", updated.RedirectCount)
	}

	updated, err = repo.IncrementRedirectCount(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("second IncrementRedirectCount() returned error: %v", err)
	}

	if updated.RedirectCount != 2 {
		t.Errorf("RedirectCount after second increment: got %d, want 2", updated.RedirectCount)
	}
}

func TestPostgresURLRepository_IncrementRedirectCount_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	_, err := repo.IncrementRedirectCount(context.Background(), "nonexist")
	if err == nil {
		t.Fatal("expected error for non-existent shortcode, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestPostgresURLRepository_IncrementRedirectCount_ExcludesDeleted(t *testing.T) {
	db := newTestDB(t)
	repo := NewPostgresURLRepository(db)
	cleanTable(t, db)

	created, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   "incdel01",
		OriginalURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	_, err = repo.SoftDelete(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("SoftDelete() failed: %v", err)
	}

	_, err = repo.IncrementRedirectCount(context.Background(), created.ShortCode)
	if err == nil {
		t.Fatal("expected error for deleted URL, got nil")
	}
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound for deleted URL, got %v", err)
	}
}
