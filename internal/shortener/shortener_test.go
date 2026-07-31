package shortener

import (
	"testing"

	"github.com/mmk31585/url-shortener/internal/domain"
)

func TestGenerate_Returns8Chars(t *testing.T) {
	s := NewRandomShortener()
	for i := 0; i < 1000; i++ {
		code, err := s.Generate()
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}
		if len(code) != ShortCodeLength {
			t.Errorf("iteration %d: len(code)=%d, want %d", i, len(code), ShortCodeLength)
		}
	}
}

func TestGenerate_ValidBase62Only(t *testing.T) {
	s := NewRandomShortener()
	valid := map[byte]bool{}
	for _, c := range base62Alphabet {
		valid[byte(c)] = true
	}

	for i := 0; i < 1000; i++ {
		code, err := s.Generate()
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}
		for j := 0; j < len(code); j++ {
			if !valid[code[j]] {
				t.Errorf("iteration %d: code[%d]=%q not in base62 alphabet", i, j, code[j])
			}
		}
	}
}

func TestGenerate_NoCollisionsIn10K(t *testing.T) {
	s := NewRandomShortener()
	seen := make(map[domain.ShortCode]bool, 10000)

	for i := 0; i < 10000; i++ {
		code, err := s.Generate()
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}
		if seen[code] {
			t.Errorf("collision at iteration %d: %q", i, code)
		}
		seen[code] = true
	}
}

func TestGenerate_UniqueAcrossCalls(t *testing.T) {
	s := NewRandomShortener()
	seen := make(map[domain.ShortCode]bool, 1000)

	for i := 0; i < 1000; i++ {
		code, err := s.Generate()
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}
		if seen[code] {
			t.Errorf("duplicate code at iteration %d: %q", i, code)
		}
		seen[code] = true
	}
}

func TestNewRandomShortener_Defaults(t *testing.T) {
	s := NewRandomShortener()
	if s.length != ShortCodeLength {
		t.Errorf("length: got %d, want %d", s.length, ShortCodeLength)
	}
	if s.alphabet != base62Alphabet {
		t.Errorf("alphabet: got %q, want %q", s.alphabet, base62Alphabet)
	}
}
