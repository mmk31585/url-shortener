package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mmk31585/url-shortener/internal/handler"
)

// captureHandler stores slog records for assertion.
type captureRecord struct {
	msg   string
	attrs map[string]slog.Value
}

type captureHandler struct {
	records []captureRecord
}

func newCaptureLogger() (*slog.Logger, *captureHandler) {
	c := &captureHandler{}
	return slog.New(c), c
}

func (c *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (c *captureHandler) Handle(_ context.Context, rec slog.Record) error {
	attrs := map[string]slog.Value{}
	rec.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value
		return true
	})
	c.records = append(c.records, captureRecord{msg: rec.Message, attrs: attrs})
	return nil
}

func (c *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return c }
func (c *captureHandler) WithGroup(string) slog.Handler            { return c }

func (c *captureHandler) first() captureRecord {
	if len(c.records) == 0 {
		return captureRecord{attrs: map[string]slog.Value{}}
	}
	return c.records[0]
}

func (c *captureHandler) findByMessage(msg string) (captureRecord, bool) {
	for _, rec := range c.records {
		if rec.msg == msg {
			return rec, true
		}
	}
	return captureRecord{}, false
}

// --- RequestID ---

func TestRequestID_AddsHeader(t *testing.T) {
	var gotID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = r.Header.Get(requestIDHeader)
		w.WriteHeader(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	RequestID(next).ServeHTTP(w, r)

	if w.Header().Get(requestIDHeader) == "" {
		t.Error("X-Request-ID header missing from response")
	}
	if gotID == "" {
		t.Error("request ID missing from request header in downstream handler")
	}
	if gotID != w.Header().Get(requestIDHeader) {
		t.Errorf("downstream header %q != response header %q", gotID, w.Header().Get(requestIDHeader))
	}
}

func TestRequestID_GeneratedIsUUID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	RequestID(next).ServeHTTP(w, r)

	id := w.Header().Get(requestIDHeader)
	if len(id) != 36 {
		t.Errorf("expected UUID v4 length 36, got %q (len %d)", id, len(id))
	}
}

func TestRequestID_PropagatedToContext(t *testing.T) {
	var fromContext string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromContext = handler.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	RequestID(next).ServeHTTP(w, r)

	if fromContext == "" {
		t.Fatal("request ID missing from context")
	}
	if fromContext != w.Header().Get(requestIDHeader) {
		t.Errorf("context %q != header %q", fromContext, w.Header().Get(requestIDHeader))
	}
}

func TestRequestID_ReusesClientHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(requestIDHeader, "client-provided-id")

	RequestID(next).ServeHTTP(w, r)

	if got := w.Header().Get(requestIDHeader); got != "client-provided-id" {
		t.Errorf("expected client header to pass through, got %q", got)
	}
}

// --- Logging ---

func TestLogging_LogsRequestMetadata(t *testing.T) {
	logger, capture := newCaptureLogger()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"ok":true}`))
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	r.Header.Set(requestIDHeader, "req-123")

	Logging(logger)(next).ServeHTTP(w, r)

	rec := capture.first()
	if rec.msg != "request" {
		t.Errorf("log message: got %q, want request", rec.msg)
	}
	if got := rec.attrs["method"].String(); got != http.MethodPost {
		t.Errorf("method: got %q", got)
	}
	if got := rec.attrs["path"].String(); got != "/api/v1/urls" {
		t.Errorf("path: got %q", got)
	}
	if got := rec.attrs["status"].Int64(); got != http.StatusCreated {
		t.Errorf("status: got %d", got)
	}
	if got := rec.attrs["request_id"].String(); got != "req-123" {
		t.Errorf("request_id: got %q", got)
	}
	if rec.attrs["duration"].Duration() == 0 {
		t.Error("duration must be non-zero")
	}
}

func TestLogging_RecordsImplicit200(t *testing.T) {
	logger, capture := newCaptureLogger()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	Logging(logger)(next).ServeHTTP(w, r)

	rec := capture.first()
	if got := rec.attrs["status"].Int64(); got != http.StatusOK {
		t.Errorf("status: got %d, want %d", got, http.StatusOK)
	}
}

func TestLogging_StatusWriterUnwrapsToResponseController(t *testing.T) {
	logger, _ := newCaptureLogger()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		if err := rc.Flush(); err != nil {
			t.Errorf("Flush through statusWriter: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	Logging(logger)(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
}

// --- Recovery ---

func TestRecovery_Returns500(t *testing.T) {
	logger, capture := newCaptureLogger()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/abc11111", nil)
	r.Header.Set(requestIDHeader, "panic-req")

	Recovery(logger)(next).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
	body := w.Body.String()
	if !strings.Contains(body, "INTERNAL_ERROR") {
		t.Errorf("expected consistent error envelope in body, got %q", body)
	}
	if !strings.Contains(body, "panic-req") {
		t.Errorf("expected request_id in error envelope, got %q", body)
	}
	panicRec, ok := capture.findByMessage("panic recovered")
	if !ok {
		t.Fatal("expected 'panic recovered' log record")
	}
	if !strings.Contains(panicRec.attrs["stack"].String(), "panic") {
		t.Error("expected stack trace in log")
	}
}

func TestRecovery_DoesNotInterfereWithNormalRequests(t *testing.T) {
	logger, _ := newCaptureLogger()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	Recovery(logger)(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
}
