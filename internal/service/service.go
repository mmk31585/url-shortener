package service

import (
	"context"

	"github.com/mmk31585/url-shortener/internal/domain"
)

type URLService interface {
	CreateURL(ctx context.Context, url string) (domain.URL, error)
	GetURL(ctx context.Context, code domain.ShortCode) (domain.URL, error)
	ListURLs(ctx context.Context) ([]domain.URL, error)
	UpdateURL(ctx context.Context, originalUrl string, code domain.ShortCode) (domain.URL, error)
	DeleteURL(ctx context.Context, code domain.ShortCode) (domain.URL, error)
	Redirect(ctx context.Context, code domain.ShortCode) (string, error)
}
