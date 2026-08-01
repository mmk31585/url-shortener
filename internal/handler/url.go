package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/mmk31585/url-shortener/internal/domain"
	"github.com/mmk31585/url-shortener/internal/error"
	"github.com/mmk31585/url-shortener/internal/json"
	"github.com/mmk31585/url-shortener/internal/validator"
)

// CreateURL godoc
// @Summary      Create a new short URL
// @Description  Create a new shortened URL from a long URL
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        url  body      CreateURLPayload  true  "URL to shorten"
// @Success      201  {object}  URLResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      422  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/urls [post]
func (b *BaseHandler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var payload CreateURLPayload
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

// GetURL godoc
// @Summary      Get URL info by shortcode
// @Description  Retrieve details about a shortened URL
// @Tags         urls
// @Produce      json
// @Param        shortcode  path      string  true  "8-character shortcode"
// @Success      200        {object}  URLResponse
// @Failure      400        {object}  ErrorResponse
// @Failure      404        {object}  ErrorResponse
// @Router       /api/v1/urls/{shortcode} [get]
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

// Redirect godoc
// @Summary      Redirect to original URL
// @Description  Redirect to the original URL and increment redirect count
// @Tags         redirect
// @Param        shortcode  path      string  true  "8-character shortcode"
// @Success      302        "Redirect to original URL"
// @Failure      400        {object}  ErrorResponse
// @Failure      404        {object}  ErrorResponse
// @Failure      410        {object}  ErrorResponse
// @Router       /{shortcode} [get]
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
// UpdateURL godoc
// @Summary      Update URL destination
// @Description  Update the destination URL for an existing shortcode
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        shortcode  path      string             true  "8-character shortcode"
// @Param        url        body      UpdateURLPayload   true  "New destination URL"
// @Success      200        {object}  URLResponse
// @Failure      400        {object}  ErrorResponse
// @Failure      404        {object}  ErrorResponse
// @Router       /api/v1/urls/{shortcode} [put]
func (b *BaseHandler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("shortcode")
	if err := validator.ValidateShortCode(code); err != nil {
		b.writeError(w, r, err)
		return
	}
	var payload UpdateURLPayload
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
// DeleteURL godoc
// @Summary      Delete a short URL
// @Description  Soft-delete a shortened URL (returns 410 on future redirect)
// @Tags         urls
// @Param        shortcode  path      string  true  "8-character shortcode"
// @Success      204        "URL deleted"
// @Failure      400        {object}  ErrorResponse
// @Failure      404        {object}  ErrorResponse
// @Router       /api/v1/urls/{shortcode} [delete]
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
// GetStats godoc
// @Summary      Get redirect statistics
// @Description  Get redirect count and metadata for a shortened URL
// @Tags         urls
// @Produce      json
// @Param        shortcode  path      string           true  "8-character shortcode"
// @Success      200        {object}  StatsResponse
// @Failure      400        {object}  ErrorResponse
// @Failure      404        {object}  ErrorResponse
// @Router       /api/v1/urls/{shortcode}/stats [get]
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

// Health godoc
// @Summary      Health check
// @Description  Check server and database connectivity
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  ErrorResponse
// @Router       /health [get]
func (b *BaseHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := b.db.PingContext(ctx); err != nil {
		b.writeError(w, r, errs.NewServiceUnavailableError("database is unavailable"))
		return
	}
	data := HealthResponse{
		Status:    "ok",
		Database:  "connected",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if err := json.WriteJSON(w, http.StatusOK, data); err != nil {
		b.logger.Error("failed to write the json response", "error", err)
	}
}
