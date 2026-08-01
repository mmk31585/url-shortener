package errs

import "net/http"

type AppError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

const (
	CodeValidationError    = "VALIDATION_ERROR"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeGone               = "GONE"
	CodeUnprocessable      = "UNPROCESSABLE_ENTITY"
	CodeInternal           = "INTERNAL_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

func newAppError(statusCode int, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

func NewValidationError(message string) *AppError {
	return newAppError(http.StatusBadRequest, CodeValidationError, message)
}

func NewNotFoundError(message string) *AppError {
	return newAppError(http.StatusNotFound, CodeNotFound, message)
}

func NewConflictError(message string) *AppError {
	return newAppError(http.StatusConflict, CodeConflict, message)
}

func NewGoneError(message string) *AppError {
	return newAppError(http.StatusGone, CodeGone, message)
}

func NewUnprocessableError(message string) *AppError {
	return newAppError(http.StatusUnprocessableEntity, CodeUnprocessable, message)
}

func NewInternalError(message string) *AppError {
	return newAppError(http.StatusInternalServerError, CodeInternal, message)
}

func NewServiceUnavailableError(message string) *AppError {
	return newAppError(http.StatusServiceUnavailable, CodeServiceUnavailable, message)
}
