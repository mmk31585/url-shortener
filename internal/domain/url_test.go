package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestURL_DefaultValues(t *testing.T) {
	u := URL{}
	if u.ID != 0 {
		t.Errorf("expected ID to be 0, got %d", u.ID)
	}
	if u.ShortCode != "" {
		t.Errorf("expected ShortCode to be empty, got %q", u.ShortCode)
	}
	if u.OriginalURL != "" {
		t.Errorf("expected OriginalURL to be empty, got %q", u.OriginalURL)
	}
	if u.RedirectCount != 0 {
		t.Errorf("expected RedirectCount to be 0, got %d", u.RedirectCount)
	}
	if !u.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be zero, got %v", u.CreatedAt)
	}
	if !u.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be zero, got %v", u.UpdatedAt)
	}
	if u.DeletedAt != nil {
		t.Errorf("expected DeletedAt to be nil, got %v", u.DeletedAt)
	}
}

func TestURL_DeletedAtIsNilByDefault(t *testing.T) {
	u := URL{}
	if u.DeletedAt != nil {
		t.Error("DeletedAt should be nil (not deleted) by default")
	}
}

func TestURL_JSONMarshalUnmarshal(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	u := URL{
		ID:            1,
		ShortCode:     "abc123de",
		OriginalURL:   "https://example.com/long/path",
		RedirectCount: 42,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("failed to marshal URL: %v", err)
	}

	var decoded URL
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal URL: %v", err)
	}

	if decoded.ID != u.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, u.ID)
	}
	if decoded.ShortCode != u.ShortCode {
		t.Errorf("ShortCode: got %q, want %q", decoded.ShortCode, u.ShortCode)
	}
	if decoded.OriginalURL != u.OriginalURL {
		t.Errorf("OriginalURL: got %q, want %q", decoded.OriginalURL, u.OriginalURL)
	}
	if decoded.RedirectCount != u.RedirectCount {
		t.Errorf("RedirectCount: got %d, want %d", decoded.RedirectCount, u.RedirectCount)
	}
	if !decoded.CreatedAt.Equal(u.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", decoded.CreatedAt, u.CreatedAt)
	}
	if !decoded.UpdatedAt.Equal(u.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", decoded.UpdatedAt, u.UpdatedAt)
	}
}

func TestURL_JSONMarshalUnmarshalWithDeletedAt(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	u := URL{
		ID:            2,
		ShortCode:     "xyz789ab",
		OriginalURL:   "https://example.com/other",
		RedirectCount: 5,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     &now,
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("failed to marshal URL with DeletedAt: %v", err)
	}

	var decoded URL
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal URL with DeletedAt: %v", err)
	}

	if decoded.DeletedAt == nil {
		t.Fatal("expected DeletedAt to be non-nil after unmarshal")
	}
	if !decoded.DeletedAt.Equal(*u.DeletedAt) {
		t.Errorf("DeletedAt: got %v, want %v", decoded.DeletedAt, u.DeletedAt)
	}
}

func TestURL_JSONTagsAreSnakeCase(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	u := URL{
		ID:            1,
		ShortCode:     "test",
		OriginalURL:   "https://x.com",
		RedirectCount: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     &now,
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"id", "short_code", "original_url", "redirect_count", "created_at", "updated_at", "deleted_at"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q not found", key)
		}
	}
}

func TestURL_DeletedAtOmittedWhenNil(t *testing.T) {
	u := URL{ID: 1, ShortCode: "test", OriginalURL: "https://x.com"}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if containsJSONKey(data, "deleted_at") {
		t.Error("expected deleted_at to be omitted when nil")
	}
}

func TestShortCode_IsString(t *testing.T) {
	var sc ShortCode = "abc123de"
	if string(sc) != "abc123de" {
		t.Errorf("ShortCode string conversion: got %q, want %q", string(sc), "abc123de")
	}
}

func TestShortCode_EmptyIsValid(t *testing.T) {
	var sc ShortCode
	if string(sc) != "" {
		t.Errorf("expected empty ShortCode, got %q", string(sc))
	}
}

func containsJSONKey(data []byte, key string) bool {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return false
	}
	_, ok := raw[key]
	return ok
}
