package validator

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrEmptyURL          = errors.New("url is empty")
	ErrInvalidURL        = errors.New("invalid url")
	ErrInvalidURLScheme  = errors.New("url scheme must be http or https")
	ErrInvalidURLHost    = errors.New("url must have a non-empty host")
	ErrURLTooLong        = errors.New("url exceeds maximum length of 2048 characters")
	ErrInvalidShortCode  = errors.New("invalid shortcode")
	ErrInvalidShortCodeLen = errors.New("shortcode must be exactly 8 characters")
	ErrInvalidShortCodeChars = errors.New("shortcode must contain only base62 characters (a-z, A-Z, 0-9)")
	ErrInvalidID         = errors.New("id must be positive")
)

const (
	MaxURLLength      = 2048
	ShortCodeLength   = 8
	Base62Alphabet    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func ValidateURL(rawURL string) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return ErrEmptyURL
	}

	if len(rawURL) > MaxURLLength {
		return ErrURLTooLong
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURLScheme
	}

	if parsed.Host == "" {
		return ErrInvalidURLHost
	}

	return nil
}

func ValidateShortCode(code string) error {
	if len(code) != ShortCodeLength {
		return ErrInvalidShortCodeLen
	}

	valid := make(map[byte]bool, len(Base62Alphabet))
	for i := 0; i < len(Base62Alphabet); i++ {
		valid[Base62Alphabet[i]] = true
	}

	for i := 0; i < len(code); i++ {
		if !valid[code[i]] {
			return ErrInvalidShortCodeChars
		}
	}

	return nil
}

func ValidateID(id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return nil
}