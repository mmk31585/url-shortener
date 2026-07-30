package domain

import "errors"

var (
	ErrURLNotFound        = errors.New("url not found")
	ErrURLAlreadyDeleted  = errors.New("url already deleted")
	ErrShortCodeCollision = errors.New("short code is exist")
	ErrInvalidShortCode   = errors.New("short code is invalid")
	ErrInvalidURL         = errors.New("url is invalid")
)
