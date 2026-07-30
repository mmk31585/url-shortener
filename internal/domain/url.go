package domain

import "time"

type ShortCode string

type URL struct {
	ID            int64      `json:"id"`
	ShortCode     ShortCode  `json:"short_code"`
	OriginalURL   string     `json:"original_url"`
	RedirectCount int64      `json:"redirect_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}
