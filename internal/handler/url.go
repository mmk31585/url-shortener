package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/mmk31585/url-shortener/internal/domain"
	"github.com/mmk31585/url-shortener/internal/json"
	"github.com/mmk31585/url-shortener/internal/validator"
)

func (b *BaseHandler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var payload createURLPayload
	if err := json.ReadAndValidate(w, r, &payload); err != nil {
		b.writeError(w, r, err)
		return
	}

	created, err := b.service.CreateURL(r.Context(), payload.URL)
	if err != nil {
		b.writeError(w, r, err)
		return
	}

	if err := json.WriteJSON(w, http.StatusCreated, newURLResponse(r, created)); err != nil {
		b.logger.Error("failed to write response", "error", err)
	}
}

func (b *BaseHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("shortcode")
	if err := validator.ValidateShortCode(code); err != nil {
		b.writeError(w, r, err)
		return
	}
	found, err := b.service.GetURL(r.Context(), domain.ShortCode(code))
	if err != nil {
		b.writeError(w, r, err)
		return
	}
	if err := json.WriteJSON(w, http.StatusOK, newURLResponse(r, found)); err != nil {
		b.logger.Error("failed to write response", "error", err)
	}
}

func (b *BaseHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("shortcode")
	if err := validator.ValidateShortCode(code); err != nil {
		b.writeError(w, r, err)
		return
	}
	originalURL, err := b.service.Redirect(r.Context(), domain.ShortCode(code))
	if err != nil {
		b.writeError(w, r, err)
		return
	}
	w.Header().Set("Location", originalURL)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusFound)
}
func (b *BaseHandler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("shortcode")
	if err := validator.ValidateShortCode(code); err != nil {
		b.writeError(w, r, err)
		return
	}
	var payload updateurlpayload
	if err := json.ReadAndValidate(w, r, &payload); err != nil {
		b.writeError(w, r, err)
		return
	}

	updated, err := b.service.UpdateURL(r.Context(), payload.URL, domain.ShortCode(code))
	if err != nil {
		b.writeError(w, r, err)
		return
	}

	if err := json.WriteJSON(w, http.StatusOK, newURLResponse(r, updated)); err != nil {
		b.logger.Error("failed to write response", "error", err)
	}
}
func (b *BaseHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("shortcode")
	if err := validator.ValidateShortCode(code); err != nil {
		b.writeError(w, r, err)
		return
	}
	_, err := b.service.DeleteURL(r.Context(), domain.ShortCode(code))
	if err != nil {
		b.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (b *BaseHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("shortcode")
	if err := validator.ValidateShortCode(code); err != nil {
		b.writeError(w, r, err)
		return
	}
	found, err := b.service.GetURL(r.Context(), domain.ShortCode(code))
	if err != nil {
		b.writeError(w, r, err)
		return
	}
	if err := json.WriteJSON(w, http.StatusOK, newStatsResponse(found)); err != nil {
		b.logger.Error("failed to write response", "error", err)
	}
}

func (b *BaseHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := b.db.PingContext(ctx); err != nil {
		b.writeError(w, r, err)
		return
	}
	data := map[string]string{
		"status":    "ok",
		"database":  "connected",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if err := json.WriteJSON(w, http.StatusOK, data); err != nil {
		b.logger.Error("failed to write the json response", "error", err)
	}
}
