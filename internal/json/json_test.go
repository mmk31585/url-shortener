package json

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testPayload struct {
	URL  string `json:"url" validate:"required"`
	Name string `json:"name,omitempty"`
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	payload := map[string]string{"status": "ok"}

	if err := WriteJSON(w, 200, payload); err != nil {
		t.Fatalf("WriteJSON() returned error: %v", err)
	}
	if w.Code != 200 {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: got %q", ct)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("body: got %q", w.Body.String())
	}
}

func TestWriteJSON_StatusHeaderWritten(t *testing.T) {
	w := httptest.NewRecorder()
	if err := WriteJSON(w, http.StatusTeapot, testPayload{URL: "x"}); err != nil {
		t.Fatalf("WriteJSON() returned error: %v", err)
	}
	if w.Code != http.StatusTeapot {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusTeapot)
	}
}

func TestReadJSON_EmptyBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(""))

	var payload testPayload
	err := ReadJSON(w, r, &payload)
	if !errors.Is(err, ErrEmptyBody) {
		t.Errorf("expected ErrEmptyBody, got %v", err)
	}
}

func TestReadJSON_MalformedSyntax(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"url":`))

	var payload testPayload
	err := ReadJSON(w, r, &payload)
	if !errors.Is(err, ErrMalformedJSON) {
		t.Errorf("expected ErrMalformedJSON, got %v", err)
	}
}

func TestReadJSON_UnknownField(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"url":"https://x.com","bogus":1}`))

	var payload testPayload
	err := ReadJSON(w, r, &payload)
	if !errors.Is(err, ErrUnknownField) {
		t.Errorf("expected ErrUnknownField, got %v", err)
	}
}

func TestReadJSON_TypeMismatch(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"url":123}`))

	var payload testPayload
	err := ReadJSON(w, r, &payload)
	if !errors.Is(err, ErrMalformedJSON) {
		t.Errorf("expected ErrMalformedJSON, got %v", err)
	}
}

func TestReadJSON_TrailingValue(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"url":"https://x.com"} {"url":"https://y.com"}`))

	var payload testPayload
	err := ReadJSON(w, r, &payload)
	if !errors.Is(err, ErrMalformedJSON) {
		t.Errorf("expected ErrMalformedJSON for trailing value, got %v", err)
	}
}

func TestReadJSON_Valid(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"url":"https://x.com"}`))

	var payload testPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		t.Fatalf("ReadJSON() returned error: %v", err)
	}
	if payload.URL != "https://x.com" {
		t.Errorf("url: got %q", payload.URL)
	}
}

func TestReadAndValidate_MissingRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))

	var payload testPayload
	err := ReadAndValidate(w, r, &payload)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestFormatValidationError_Nil(t *testing.T) {
	if got := FormatValidationError(nil); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFormatValidationError_NonValidation(t *testing.T) {
	if got := FormatValidationError(errors.New("plain")); got != "plain" {
		t.Errorf("got %q, want plain", got)
	}
}

func TestFormatValidationError_Validation(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))

	var payload testPayload
	err := ReadAndValidate(w, r, &payload)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	formatted := FormatValidationError(err)
	if !strings.Contains(formatted, "url") || !strings.Contains(formatted, "required") {
		t.Errorf("formatted message %q should mention url is required", formatted)
	}
}
