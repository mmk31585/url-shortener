package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mmk31585/url-shortener/internal/domain"
	"github.com/mmk31585/url-shortener/internal/repository/mock"
	"github.com/mmk31585/url-shortener/internal/validator"
)

type fakeShortener struct {
	code   domain.ShortCode
	codes  []domain.ShortCode
	idx    int
	genErr error
}

func (f *fakeShortener) Generate() (domain.ShortCode, error) {
	if f.genErr != nil {
		return "", f.genErr
	}
	if len(f.codes) > 0 {
		if f.idx >= len(f.codes) {
			return "", errors.New("fake shortener exhausted")
		}
		code := f.codes[f.idx]
		f.idx++
		return code, nil
	}
	return f.code, nil
}

func newFixedShortener(code domain.ShortCode) *fakeShortener {
	return &fakeShortener{code: code}
}

func newSequenceShortener(codes ...domain.ShortCode) *fakeShortener {
	return &fakeShortener{codes: codes}
}

type failingRepo struct {
	*mock.URLRepository
}

func (f *failingRepo) Create(ctx context.Context, url *domain.URL) (domain.URL, error) {
	return domain.URL{}, errors.New("db down")
}

const (
	validURL   = "https://example.com"
	shortCode1 = domain.ShortCode("abc11111")
	shortCode2 = domain.ShortCode("abc22222")
)

func TestCreateURL_Valid(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() returned error: %v", err)
	}
	if created.ShortCode != shortCode1 {
		t.Errorf("ShortCode: got %q, want %q", created.ShortCode, shortCode1)
	}
	if created.OriginalURL != validURL {
		t.Errorf("OriginalURL: got %q, want %q", created.OriginalURL, validURL)
	}
	if created.RedirectCount != 0 {
		t.Errorf("RedirectCount: got %d, want 0", created.RedirectCount)
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateURL_EmptyURL(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	_, err := svc.CreateURL(context.Background(), "")
	if !errors.Is(err, validator.ErrEmptyURL) {
		t.Errorf("expected ErrEmptyURL, got %v", err)
	}
}

func TestCreateURL_InvalidScheme(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	_, err := svc.CreateURL(context.Background(), "ftp://example.com")
	if !errors.Is(err, validator.ErrInvalidURLScheme) {
		t.Errorf("expected ErrInvalidURLScheme, got %v", err)
	}
}

func TestCreateURL_TooLong(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	longURL := "https://example.com/" + strings.Repeat("a", 3000)
	_, err := svc.CreateURL(context.Background(), longURL)
	if !errors.Is(err, validator.ErrURLTooLong) {
		t.Errorf("expected ErrURLTooLong, got %v", err)
	}
}

func TestCreateURL_CollisionRetries(t *testing.T) {
	repo := mock.NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   shortCode1,
		OriginalURL: "https://taken.com",
	}); err != nil {
		t.Fatalf("seed Create() failed: %v", err)
	}

	svc := NewURLService(repo, newSequenceShortener(shortCode1, shortCode2))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() returned error: %v", err)
	}
	if created.ShortCode != shortCode2 {
		t.Errorf("ShortCode: got %q, want %q (collision should retry)", created.ShortCode, shortCode2)
	}
}

func TestCreateURL_CollisionExhausted(t *testing.T) {
	repo := mock.NewURLRepository()
	if _, err := repo.Create(context.Background(), &domain.URL{
		ShortCode:   shortCode1,
		OriginalURL: "https://taken.com",
	}); err != nil {
		t.Fatalf("seed Create() failed: %v", err)
	}

	svc := NewURLService(repo, newFixedShortener(shortCode1))

	_, err := svc.CreateURL(context.Background(), validURL)
	if !errors.Is(err, domain.ErrShortCodeCollision) {
		t.Errorf("expected ErrShortCodeCollision, got %v", err)
	}
}

func TestCreateURL_ShortenerError(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), &fakeShortener{genErr: errors.New("gen failed")})

	_, err := svc.CreateURL(context.Background(), validURL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateURL_RepoError(t *testing.T) {
	repo := &failingRepo{mock.NewURLRepository()}
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	_, err := svc.CreateURL(context.Background(), validURL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create url") {
		t.Errorf("expected wrapped error, got %v", err)
	}
}

func TestGetURL_Existing(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() failed: %v", err)
	}

	got, err := svc.GetURL(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("GetURL() returned error: %v", err)
	}
	if got.ShortCode != created.ShortCode {
		t.Errorf("ShortCode: got %q, want %q", got.ShortCode, created.ShortCode)
	}
	if got.OriginalURL != validURL {
		t.Errorf("OriginalURL: got %q, want %q", got.OriginalURL, validURL)
	}
}

func TestGetURL_NotFound(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	_, err := svc.GetURL(context.Background(), "missing01")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestListURLs_Empty(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	urls, err := svc.ListURLs(context.Background())
	if err != nil {
		t.Fatalf("ListURLs() returned error: %v", err)
	}
	if urls == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(urls) != 0 {
		t.Errorf("expected 0 URLs, got %d", len(urls))
	}
}

func TestListURLs_All(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newSequenceShortener(
		shortCode1, shortCode2, "abc33333",
	))

	for _, url := range []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	} {
		if _, err := svc.CreateURL(context.Background(), url); err != nil {
			t.Fatalf("CreateURL() failed: %v", err)
		}
	}

	urls, err := svc.ListURLs(context.Background())
	if err != nil {
		t.Fatalf("ListURLs() returned error: %v", err)
	}
	if len(urls) != 3 {
		t.Errorf("expected 3 URLs, got %d", len(urls))
	}
}

func TestUpdateURL_Valid(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() failed: %v", err)
	}

	newURL := "https://new-destination.com/path"
	updated, err := svc.UpdateURL(context.Background(), newURL, created.ShortCode)
	if err != nil {
		t.Fatalf("UpdateURL() returned error: %v", err)
	}
	if updated.OriginalURL != newURL {
		t.Errorf("OriginalURL: got %q, want %q", updated.OriginalURL, newURL)
	}
	if updated.ShortCode != created.ShortCode {
		t.Errorf("ShortCode changed: got %q, want %q", updated.ShortCode, created.ShortCode)
	}
}

func TestUpdateURL_InvalidURL(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() failed: %v", err)
	}

	_, err = svc.UpdateURL(context.Background(), "not-a-url", created.ShortCode)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestUpdateURL_NotFound(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	_, err := svc.UpdateURL(context.Background(), validURL, "missing01")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestDeleteURL_Valid(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() failed: %v", err)
	}

	deleted, err := svc.DeleteURL(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("DeleteURL() returned error: %v", err)
	}
	if deleted.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}

	if _, err := svc.GetURL(context.Background(), created.ShortCode); !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound after delete, got %v", err)
	}
}

func TestDeleteURL_NotFound(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	_, err := svc.DeleteURL(context.Background(), "missing01")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestRedirect_Valid(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() failed: %v", err)
	}

	dest, err := svc.Redirect(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("Redirect() returned error: %v", err)
	}
	if dest != validURL {
		t.Errorf("destination: got %q, want %q", dest, validURL)
	}
}

func TestRedirect_IncrementsCount(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() failed: %v", err)
	}

	if _, err := svc.Redirect(context.Background(), created.ShortCode); err != nil {
		t.Fatalf("first Redirect() failed: %v", err)
	}
	if _, err := svc.Redirect(context.Background(), created.ShortCode); err != nil {
		t.Fatalf("second Redirect() failed: %v", err)
	}

	stats, err := svc.GetStats(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}
	if stats != 2 {
		t.Errorf("redirect count: got %d, want 2", stats)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	_, err := svc.Redirect(context.Background(), "missing01")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

func TestGetStats_Valid(t *testing.T) {
	repo := mock.NewURLRepository()
	svc := NewURLService(repo, newFixedShortener(shortCode1))

	created, err := svc.CreateURL(context.Background(), validURL)
	if err != nil {
		t.Fatalf("CreateURL() failed: %v", err)
	}

	count, err := svc.GetStats(context.Background(), created.ShortCode)
	if err != nil {
		t.Fatalf("GetStats() returned error: %v", err)
	}
	if count != 0 {
		t.Errorf("redirect count: got %d, want 0", count)
	}
}

func TestGetStats_NotFound(t *testing.T) {
	svc := NewURLService(mock.NewURLRepository(), newFixedShortener(shortCode1))

	_, err := svc.GetStats(context.Background(), "missing01")
	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}
