package repository

import (
	"context"

	"github.com/mmk31585/url-shortener/internal/domain"
)

type URLRepository interface {
	Create(ctx context.Context, url *domain.URL) (*domain.URL, error)
	GetByShortCode(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
	GetAll(ctx context.Context) ([]domain.URL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error)
	Update(ctx context.Context, url *domain.URL) (*domain.URL, error)
	SoftDelete(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
	IncrementRedirectCount(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
}
