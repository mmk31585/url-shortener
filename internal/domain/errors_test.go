package domain

import (
	"errors"
	"fmt"
	"testing"
)

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

	wrappedSentinel := fmt.Errorf("context: %w", ErrURLNotFound)
	if !errors.Is(wrappedSentinel, ErrURLNotFound) {
		t.Error("wrapped sentinel should match ErrURLNotFound")
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
