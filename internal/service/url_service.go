package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmk31585/url-shortener/internal/domain"
	"github.com/mmk31585/url-shortener/internal/repository"
	"github.com/mmk31585/url-shortener/internal/shortener"
	"github.com/mmk31585/url-shortener/internal/validator"
)

var _ URLService = (*urlService)(nil)

type urlService struct {
	repo      repository.URLRepository
	shortener shortener.Shortener
}

func NewURLService(repo repository.URLRepository, shortener shortener.Shortener) *urlService {
	return &urlService{
		repo:      repo,
		shortener: shortener,
	}
}

func (s *urlService) findURL(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	url, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		return domain.URL{}, fmt.Errorf("service: failed to find url: %w", err)
	}
	return url, nil
}

func (s *urlService) CreateURL(ctx context.Context, url string) (domain.URL, error) {
	err := validator.ValidateURL(url)
	if err != nil {
		return domain.URL{}, err
	}
	const maxRetries = 5
	for range maxRetries {
		code, err := s.shortener.Generate()
		if err != nil {
			return domain.URL{}, err
		}
		created, err := s.repo.Create(ctx, &domain.URL{
			ShortCode:   code,
			OriginalURL: url,
		})
		if err == nil {
			return created, nil
		}
		if !errors.Is(err, domain.ErrShortCodeCollision) {
			return domain.URL{}, fmt.Errorf("service: failed to create url: %w", err)
		}
	}
	return domain.URL{}, domain.ErrShortCodeCollision

}
func (s *urlService) GetURL(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	return s.findURL(ctx, code)
}
func (s *urlService) ListURLs(ctx context.Context) ([]domain.URL, error) {
	return s.repo.GetAll(ctx)
}
func (s *urlService) UpdateURL(ctx context.Context, originalUrl string, code domain.ShortCode) (domain.URL, error) {
	err := validator.ValidateURL(originalUrl)
	if err != nil {
		return domain.URL{}, err
	}

	if _, err := s.findURL(ctx, code); err != nil {
		return domain.URL{}, err
	}
	return s.repo.Update(ctx, &domain.URL{
		ShortCode:   code,
		OriginalURL: originalUrl,
	})
}
func (s *urlService) DeleteURL(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	if _, err := s.findURL(ctx, code); err != nil {
		return domain.URL{}, err
	}
	return s.repo.SoftDelete(ctx, code)
}
func (s *urlService) Redirect(ctx context.Context, code domain.ShortCode) (string, error) {
	updated, err := s.repo.IncrementRedirectCount(ctx, code)
	if err != nil {
		return "", err
	}
	return updated.OriginalURL, nil
}
