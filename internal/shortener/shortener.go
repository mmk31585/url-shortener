package shortener

import (
	"crypto/rand"

	"github.com/mmk31585/url-shortener/internal/domain"
)

const ShortCodeLength = 8

const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Shortener interface {
	Generate() (domain.ShortCode, error)
}

type RandomShortener struct {
	length   int
	alphabet string
}

func NewRandomShortener(length ...int) *RandomShortener {
	l := ShortCodeLength
	if len(length) > 0 && length[0] > 0 {
		l = length[0]
	}
	return &RandomShortener{
		length:   l,
		alphabet: base62Alphabet,
	}
}
func (r *RandomShortener) Generate() (domain.ShortCode, error) {
	b := make([]byte, r.length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	code := make([]byte, r.length)
	for i, v := range b {
		code[i] = r.alphabet[int(v)%len(r.alphabet)]
	}
	return domain.ShortCode(code), nil
}
