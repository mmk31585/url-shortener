package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	playground "github.com/go-playground/validator/v10"
	"github.com/mmk31585/url-shortener/internal/domain"
	"github.com/mmk31585/url-shortener/internal/error"
	"github.com/mmk31585/url-shortener/internal/json"
	"github.com/mmk31585/url-shortener/internal/validator"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

func mapAppError(err error) *errs.AppError {
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	var validationErrors playground.ValidationErrors
	if errors.As(err, &validationErrors) {
		return errs.NewValidationError(json.FormatValidationError(err))
	}

	switch {
	case errors.Is(err, json.ErrEmptyBody),
		errors.Is(err, json.ErrMalformedJSON),
		errors.Is(err, json.ErrUnknownField),
		errors.Is(err, validator.ErrEmptyURL),
		errors.Is(err, validator.ErrInvalidURL),
		errors.Is(err, validator.ErrInvalidURLScheme),
		errors.Is(err, validator.ErrInvalidURLHost),
		errors.Is(err, validator.ErrInvalidShortCodeLen),
		errors.Is(err, validator.ErrInvalidShortCodeChars),
		errors.Is(err, validator.ErrInvalidShortCode),
		errors.Is(err, domain.ErrInvalidShortCode),
		errors.Is(err, domain.ErrInvalidURL):
		return errs.NewValidationError(err.Error())
	case errors.Is(err, validator.ErrURLTooLong):
		return errs.NewUnprocessableError(err.Error())
	case errors.Is(err, domain.ErrURLNotFound):
		return errs.NewNotFoundError("url not found")
	case errors.Is(err, domain.ErrURLAlreadyDeleted):
		return errs.NewGoneError("url has been deleted")
	case errors.Is(err, domain.ErrShortCodeCollision):
		return errs.NewConflictError("short code already exists")
	default:
		return errs.NewInternalError("something went wrong")
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	appErr := mapAppError(err)
	requestID := RequestIDFromContext(r.Context())
	if requestID == "" {
		requestID = r.Header.Get("X-Request-ID")
	}
	logger.Error("request failed",
		"error", err.Error(),
		"code", appErr.Code,
		"status", appErr.StatusCode,
		"method", r.Method,
		"path", r.URL.Path,
		"request_id", requestID,
	)
	envelope := ErrorResponse{
		Error: ErrorDetail{
			Code:      appErr.Code,
			Message:   appErr.Message,
			RequestID: requestID,
		},
	}
	if writeErr := json.WriteJSON(w, appErr.StatusCode, envelope); writeErr != nil {
		logger.Error("failed to write error response", "error", writeErr)
	}
}

func (b *BaseHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	WriteError(w, r, b.logger, err)
}
