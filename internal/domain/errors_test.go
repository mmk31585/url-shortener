package domain

import (
	"errors"
	"testing"
)

func TestErrURLNotFound(t *testing.T) {
	if !errors.Is(ErrURLNotFound, ErrURLNotFound) {
		t.Error("ErrURLNotFound should match itself")
	}
	if errors.Is(ErrURLNotFound, ErrURLAlreadyDeleted) {
		t.Error("ErrURLNotFound should not match ErrURLAlreadyDeleted")
	}
}

func TestErrURLAlreadyDeleted(t *testing.T) {
	if !errors.Is(ErrURLAlreadyDeleted, ErrURLAlreadyDeleted) {
		t.Error("ErrURLAlreadyDeleted should match itself")
	}
}

func TestErrShortCodeCollision(t *testing.T) {
	if !errors.Is(ErrShortCodeCollision, ErrShortCodeCollision) {
		t.Error("ErrShortCodeCollision should match itself")
	}
	if errors.Is(ErrShortCodeCollision, ErrURLNotFound) {
		t.Error("ErrShortCodeCollision should not match ErrURLNotFound")
	}
}

func TestErrInvalidShortCode(t *testing.T) {
	if !errors.Is(ErrInvalidShortCode, ErrInvalidShortCode) {
		t.Error("ErrInvalidShortCode should match itself")
	}
}

func TestErrInvalidURL(t *testing.T) {
	if !errors.Is(ErrInvalidURL, ErrInvalidURL) {
		t.Error("ErrInvalidURL should match itself")
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	errs := []error{
		ErrURLNotFound,
		ErrURLAlreadyDeleted,
		ErrShortCodeCollision,
		ErrInvalidShortCode,
		ErrInvalidURL,
	}

	for i, a := range errs {
		for j, b := range errs {
			if i == j {
				continue
			}
			if errors.Is(a, b) {
				t.Errorf("sentinel errors %v and %v should be distinct", a, b)
			}
		}
	}
}

func TestSentinelErrorsWrap(t *testing.T) {
	wrapped := struct {
		error
	}{errors.New("this should not match")}
	if errors.Is(wrapped, ErrURLNotFound) {
		t.Error("wrapped non-sentinel should not match ErrURLNotFound")
	}
}

func TestSentinelErrorMessages(t *testing.T) {
	tests := []struct {
		err     error
		wantMsg string
	}{
		{ErrURLNotFound, "url not found"},
		{ErrURLAlreadyDeleted, "url already deleted"},
		{ErrShortCodeCollision, "short code is exist"},
		{ErrInvalidShortCode, "short code is invalid"},
		{ErrInvalidURL, "url is invalid"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.wantMsg {
			t.Errorf("got message %q, want %q", tt.err.Error(), tt.wantMsg)
		}
	}
}
