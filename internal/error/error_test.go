package errs

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := NewValidationError("bad input")
	if err.Error() != "bad input" {
		t.Errorf("Error(): got %q, want %q", err.Error(), "bad input")
	}
}

func TestAppError_ImplementsError(t *testing.T) {
	var _ error = NewValidationError("x")
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("msg")
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode: got %d", err.StatusCode)
	}
	if err.Code != CodeValidationError {
		t.Errorf("Code: got %q", err.Code)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("msg")
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode: got %d", err.StatusCode)
	}
	if err.Code != CodeNotFound {
		t.Errorf("Code: got %q", err.Code)
	}
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("msg")
	if err.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode: got %d", err.StatusCode)
	}
	if err.Code != CodeConflict {
		t.Errorf("Code: got %q", err.Code)
	}
}

func TestNewGoneError(t *testing.T) {
	err := NewGoneError("msg")
	if err.StatusCode != http.StatusGone {
		t.Errorf("StatusCode: got %d", err.StatusCode)
	}
	if err.Code != CodeGone {
		t.Errorf("Code: got %q", err.Code)
	}
}

func TestNewUnprocessableError(t *testing.T) {
	err := NewUnprocessableError("msg")
	if err.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode: got %d", err.StatusCode)
	}
	if err.Code != CodeUnprocessable {
		t.Errorf("Code: got %q", err.Code)
	}
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("msg")
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode: got %d", err.StatusCode)
	}
	if err.Code != CodeInternal {
		t.Errorf("Code: got %q", err.Code)
	}
}

func TestNewServiceUnavailableError(t *testing.T) {
	err := NewServiceUnavailableError("msg")
	if err.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode: got %d", err.StatusCode)
	}
	if err.Code != CodeServiceUnavailable {
		t.Errorf("Code: got %q", err.Code)
	}
}

func TestAppError_ErrorsAs(t *testing.T) {
	var target *AppError
	wrapped := errors.Join(NewNotFoundError("nope"), errors.New("other"))
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should match *AppError through wrapped error")
	}
	if target.Code != CodeNotFound {
		t.Errorf("Code: got %q", target.Code)
	}
}
