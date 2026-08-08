package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mmk31585/url-shortener/internal/domain"
)

var (
	QueryTimeoutDuration = time.Second * 5
)

type PostgresURLRepository struct {
	db *sql.DB
}

func NewPostgresURLRepository(db *sql.DB) *PostgresURLRepository {
	return &PostgresURLRepository{db: db}
}

func scanURL(s interface{ Scan(dest ...any) error }) (domain.URL, error) {
	var u domain.URL
	err := s.Scan(
		&u.ID, &u.ShortCode, &u.OriginalURL,
		&u.RedirectCount, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	return u, err
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrURLNotFound
	}
	return err
}

func (r *PostgresURLRepository) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, QueryTimeoutDuration)
}

func (r *PostgresURLRepository) Create(ctx context.Context, url *domain.URL) (domain.URL, error) {
	query := `
	INSERT INTO urls (short_code, original_url, redirect_count, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at`

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	now := time.Now()
	newURL, err := scanURL(r.db.QueryRowContext(ctx, query,
		url.ShortCode, url.OriginalURL, url.RedirectCount, now, now))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.URL{}, fmt.Errorf("repository: %w", domain.ErrShortCodeCollision)
		}
		return domain.URL{}, fmt.Errorf("repository: failed to create url: %w", err)
	}
	return newURL, nil
}

func (r *PostgresURLRepository) GetByShortCode(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	query := `
	SELECT id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at
	FROM urls
	WHERE short_code = $1 AND deleted_at IS NULL`

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	url, err := scanURL(r.db.QueryRowContext(ctx, query, code))
	if err != nil {
		return domain.URL{}, fmt.Errorf("repository: failed to get url by shortcode: %w", notFound(err))
	}
	return url, nil
}

func (r *PostgresURLRepository) GetAll(ctx context.Context) ([]domain.URL, error) {
	query := `
	SELECT id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at
	FROM urls
	WHERE deleted_at IS NULL
	ORDER BY created_at DESC`

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get all urls: %w", err)
	}
	defer rows.Close()

	var urls []domain.URL
	for rows.Next() {
		u, err := scanURL(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: failed to scan url: %w", err)
		}
		urls = append(urls, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows iteration error: %w", err)
	}
	if urls == nil {
		urls = []domain.URL{}
	}
	return urls, nil
}

func (r *PostgresURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (domain.URL, error) {
	query := `
	SELECT id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at
	FROM urls
	WHERE original_url = $1 AND deleted_at IS NULL`

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	url, err := scanURL(r.db.QueryRowContext(ctx, query, originalURL))
	if err != nil {
		return domain.URL{}, fmt.Errorf("repository: failed to get url by original url: %w", notFound(err))
	}
	return url, nil
}

func (r *PostgresURLRepository) Update(ctx context.Context, url *domain.URL) (domain.URL, error) {
	query := `
	UPDATE urls
	SET original_url = $1, updated_at = NOW()
	WHERE short_code = $2 AND deleted_at IS NULL
	RETURNING id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at`

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	newURL, err := scanURL(r.db.QueryRowContext(ctx, query, url.OriginalURL, url.ShortCode))
	if err != nil {
		return domain.URL{}, fmt.Errorf("repository: failed to update url: %w", notFound(err))
	}
	return newURL, nil
}

func (r *PostgresURLRepository) SoftDelete(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	checkQuery := `SELECT id FROM urls WHERE short_code = $1`
	updateQuery := `
	UPDATE urls
	SET deleted_at = NOW()
	WHERE short_code = $1 AND deleted_at IS NULL
	RETURNING id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at`

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var id int64
	err := r.db.QueryRowContext(ctx, checkQuery, code).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.URL{}, fmt.Errorf("repository: failed to soft delete url: %w", domain.ErrURLNotFound)
		}
		return domain.URL{}, fmt.Errorf("repository: failed to soft delete url: %w", err)
	}

	url, err := scanURL(r.db.QueryRowContext(ctx, updateQuery, code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.URL{}, fmt.Errorf("repository: failed to soft delete url: %w", domain.ErrURLAlreadyDeleted)
		}
		return domain.URL{}, fmt.Errorf("repository: failed to soft delete url: %w", err)
	}
	return url, nil
}

func (r *PostgresURLRepository) IncrementRedirectCount(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	checkQuery := `SELECT id FROM urls WHERE short_code = $1 AND deleted_at IS NULL`
	updateQuery := `
	UPDATE urls
	SET redirect_count = redirect_count + 1, updated_at = NOW()
	WHERE short_code = $1 AND deleted_at IS NULL
	RETURNING id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at`

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var id int64
	err := r.db.QueryRowContext(ctx, checkQuery, code).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.URL{}, fmt.Errorf("repository: failed to increment redirect count: %w", domain.ErrURLNotFound)
		}
		return domain.URL{}, fmt.Errorf("repository: failed to increment redirect count: %w", err)
	}

	url, err := scanURL(r.db.QueryRowContext(ctx, updateQuery, code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.URL{}, fmt.Errorf("repository: failed to increment redirect count: %w", domain.ErrURLAlreadyDeleted)
		}
		return domain.URL{}, fmt.Errorf("repository: failed to increment redirect count: %w", err)
	}
	return url, nil
}
