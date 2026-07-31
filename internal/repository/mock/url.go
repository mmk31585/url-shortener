package mock

import (
	"context"
	"time"

	"github.com/mmk31585/url-shortener/internal/domain"
)

type URLRepository struct {
	lock chan struct{}
	urls map[int64]*domain.URL
	seq  int64
}

func NewURLRepository() *URLRepository {
	return &URLRepository{
		lock: make(chan struct{}, 1),
		urls: make(map[int64]*domain.URL),
	}
}

func (m *URLRepository) acquire(ctx context.Context) error {
	select {
	case m.lock <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *URLRepository) release() {
	<-m.lock
}

func (m *URLRepository) Create(ctx context.Context, url *domain.URL) (*domain.URL, error) {
	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	for _, u := range m.urls {
		if u.ShortCode == url.ShortCode {
			return nil, domain.ErrShortCodeCollision
		}
	}

	m.seq++
	now := time.Now().UTC()
	created := &domain.URL{
		ID:            m.seq,
		ShortCode:     url.ShortCode,
		OriginalURL:   url.OriginalURL,
		RedirectCount: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	m.urls[created.ID] = created
	return created, nil
}

func (m *URLRepository) GetByShortCode(ctx context.Context, code domain.ShortCode) (*domain.URL, error) {
	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	for _, u := range m.urls {
		if u.ShortCode == code && u.DeletedAt == nil {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrURLNotFound
}

func (m *URLRepository) GetAll(ctx context.Context) ([]domain.URL, error) {
	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	var result []domain.URL
	for _, u := range m.urls {
		if u.DeletedAt == nil {
			result = append(result, *u)
		}
	}
	if result == nil {
		result = []domain.URL{}
	}
	return result, nil
}

func (m *URLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	for _, u := range m.urls {
		if u.OriginalURL == originalURL && u.DeletedAt == nil {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrURLNotFound
}

func (m *URLRepository) Update(ctx context.Context, url *domain.URL) (*domain.URL, error) {
	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	for _, u := range m.urls {
		if u.ShortCode == url.ShortCode && u.DeletedAt == nil {
			u.OriginalURL = url.OriginalURL
			u.UpdatedAt = time.Now().UTC()
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrURLNotFound
}

func (m *URLRepository) SoftDelete(ctx context.Context, code domain.ShortCode) (*domain.URL, error) {
	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	for _, u := range m.urls {
		if u.ShortCode == code && u.DeletedAt == nil {
			now := time.Now().UTC()
			u.DeletedAt = &now
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrURLNotFound
}

func (m *URLRepository) IncrementRedirectCount(ctx context.Context, code domain.ShortCode) (*domain.URL, error) {
	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	for _, u := range m.urls {
		if u.ShortCode == code && u.DeletedAt == nil {
			u.RedirectCount++
			u.UpdatedAt = time.Now().UTC()
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrURLNotFound
}
