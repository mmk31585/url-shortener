package handler

import (
	"net/http"
	"time"

	"github.com/mmk31585/url-shortener/internal/domain"
)

type createURLPayload struct {
	URL string `json:"url" validate:"required"`
}

type urlResponse struct {
	ID            int64     `json:"id"`
	ShortCode     string    `json:"short_code"`
	ShortURL      string    `json:"short_url"`
	OriginalURL   string    `json:"original_url"`
	RedirectCount int64     `json:"redirect_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type updateurlpayload struct {
	URL string `json:"url" validate:"required"`
}

type statsResponse struct {
	ShortCode     string    `json:"short_code"`
	RedirectCount int64     `json:"redirect_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func newStatsResponse(url domain.URL) statsResponse {
	return statsResponse{
		ShortCode:     string(url.ShortCode),
		RedirectCount: url.RedirectCount,
		CreatedAt:     url.CreatedAt,
		UpdatedAt:     url.UpdatedAt,
	}
}

func newURLResponse(r *http.Request, url domain.URL) urlResponse {
	return urlResponse{
		ID:            url.ID,
		ShortCode:     string(url.ShortCode),
		ShortURL:      shortURL(r, string(url.ShortCode)),
		OriginalURL:   url.OriginalURL,
		RedirectCount: url.RedirectCount,
		CreatedAt:     url.CreatedAt,
		UpdatedAt:     url.UpdatedAt,
	}
}

func shortURL(r *http.Request, code string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/r/" + code
}
