package handler

import (
	"net/http"
	"time"

	"github.com/mmk31585/url-shortener/internal/domain"
)

type CreateURLPayload struct {
	URL string `json:"url" validate:"required"`
}

type UpdateURLPayload struct {
	URL string `json:"url" validate:"required"`
}

type URLResponse struct {
	ID            int64     `json:"id"`
	ShortCode     string    `json:"short_code"`
	ShortURL      string    `json:"short_url"`
	OriginalURL   string    `json:"original_url"`
	RedirectCount int64     `json:"redirect_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StatsResponse struct {
	ShortCode     string    `json:"short_code"`
	RedirectCount int64     `json:"redirect_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

func newStatsResponse(url domain.URL) StatsResponse {
	return StatsResponse{
		ShortCode:     string(url.ShortCode),
		RedirectCount: url.RedirectCount,
		CreatedAt:     url.CreatedAt,
		UpdatedAt:     url.UpdatedAt,
	}
}

func newURLResponse(r *http.Request, url domain.URL) URLResponse {
	return URLResponse{
		ID:            url.ID,
		ShortCode:     string(url.ShortCode),
		ShortURL:      shortURL(r, string(url.ShortCode)),
		OriginalURL:   url.OriginalURL,
		RedirectCount: url.RedirectCount,
		CreatedAt:     url.CreatedAt,
		UpdatedAt:     url.UpdatedAt,
	}
}

func requestScheme(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "http" || proto == "https" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func requestHost(r *http.Request) string {
	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		return host
	}
	return r.Host
}

func shortURL(r *http.Request, code string) string {
	return requestScheme(r) + "://" + requestHost(r) + "/" + code
}
