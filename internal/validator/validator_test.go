package validator

import (
	"strings"
	"testing"
)

func TestValidateURL_Empty(t *testing.T) {
	err := ValidateURL("")
	if err != ErrEmptyURL {
		t.Errorf("expected ErrEmptyURL, got %v", err)
	}
}

func TestValidateURL_Whitespace(t *testing.T) {
	err := ValidateURL("   ")
	if err != ErrEmptyURL {
		t.Errorf("expected ErrEmptyURL, got %v", err)
	}
}

func TestValidateURL_TooLong(t *testing.T) {
	longURL := "https://example.com/" + strings.Repeat("a", MaxURLLength)
	err := ValidateURL(longURL)
	if err != ErrURLTooLong {
		t.Errorf("expected ErrURLTooLong, got %v", err)
	}
}

func TestValidateURL_InvalidParse(t *testing.T) {
	err := ValidateURL("://not a valid url")
	if err != ErrInvalidURL {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestValidateURL_InvalidScheme(t *testing.T) {
	err := ValidateURL("ftp://example.com")
	if err != ErrInvalidURLScheme {
		t.Errorf("expected ErrInvalidURLScheme, got %v", err)
	}
}

func TestValidateURL_NoHost(t *testing.T) {
	err := ValidateURL("http://")
	if err != ErrInvalidURLHost {
		t.Errorf("expected ErrInvalidURLHost, got %v", err)
	}
}

func TestValidateURL_ValidHTTP(t *testing.T) {
	err := ValidateURL("http://example.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateURL_ValidHTTPS(t *testing.T) {
	err := ValidateURL("https://example.com/path?query=value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateURL_BoundaryLengths(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{"exactly 2048", "https://e.com/" + strings.Repeat("a", MaxURLLength-14), nil},
		{"exceeds 2048", "https://e.com/" + strings.Repeat("a", MaxURLLength-13), ErrURLTooLong},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateURL(tc.url)
			if err != tc.wantErr {
				t.Errorf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestSetMaxURLLength(t *testing.T) {
	original := maxURLLength
	defer func() { maxURLLength = original }()

	SetMaxURLLength(100)
	if maxURLLength != 100 {
		t.Fatalf("maxURLLength: got %d, want 100", maxURLLength)
	}

	base := "https://e.com/"
	if err := ValidateURL(base + strings.Repeat("a", 100-len(base))); err != nil {
		t.Errorf("URL of length 100 should be valid, got %v", err)
	}
	if err := ValidateURL(base + strings.Repeat("a", 101-len(base))); err != ErrURLTooLong {
		t.Errorf("URL longer than 100 should return ErrURLTooLong, got %v", err)
	}

	SetMaxURLLength(0)
	if maxURLLength != 100 {
		t.Errorf("invalid length must be ignored, got %d", maxURLLength)
	}
}

func TestValidateShortCode_InvalidLength(t *testing.T) {
	tests := []string{"", "a", "abc", "abc123", "abc1234", "abc123456", "abc12345678"}
	for _, code := range tests {
		err := ValidateShortCode(code)
		if err != ErrInvalidShortCodeLen {
			t.Errorf("code %q: expected ErrInvalidShortCodeLen, got %v", code, err)
		}
	}
}

func TestValidateShortCode_InvalidChars(t *testing.T) {
	invalidCodes := []string{
		"abc1234!", "abc1234@", "abc1234#", "abc1234$", "abc1234%",
		"abc1234^", "abc1234&", "abc1234*", "abc1234(", "abc1234)",
		"abc1234-", "abc1234_", "abc1234=", "abc1234+", "abc1234[",
		"abc1234]", "abc1234{", "abc1234}", "abc1234|", "abc1234\\",
		"abc1234;", "abc1234:", "abc1234'", "abc1234\"", "abc1234,",
		"abc1234.", "abc1234<", "abc1234>", "abc1234/", "abc1234?",
		"abc1234 ", "abc1234\t", "abc1234\n",
	}
	for _, code := range invalidCodes {
		err := ValidateShortCode(code)
		if err != ErrInvalidShortCodeChars {
			t.Errorf("code %q: expected ErrInvalidShortCodeChars, got %v", code, err)
		}
	}
}

func TestValidateShortCode_ValidBase62(t *testing.T) {
	validCodes := []string{
		"abcdefgh", "ABCDEFGH", "01234567",
		"abc123DE", "aBcDeFgH", "zYxWvUtS",
	}
	for _, code := range validCodes {
		err := ValidateShortCode(code)
		if err != nil {
			t.Errorf("code %q: unexpected error: %v", code, err)
		}
	}
}

func TestValidateID_Invalid(t *testing.T) {
	invalidIDs := []int64{0, -1, -100}
	for _, id := range invalidIDs {
		err := ValidateID(id)
		if err != ErrInvalidID {
			t.Errorf("id %d: expected ErrInvalidID, got %v", id, err)
		}
	}
}

func TestValidateID_Valid(t *testing.T) {
	validIDs := []int64{1, 2, 100, 999999}
	for _, id := range validIDs {
		err := ValidateID(id)
		if err != nil {
			t.Errorf("id %d: unexpected error: %v", id, err)
		}
	}
}
