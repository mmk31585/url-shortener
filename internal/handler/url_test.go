package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mmk31585/url-shortener/internal/domain"
	"github.com/mmk31585/url-shortener/internal/error"
	"github.com/mmk31585/url-shortener/internal/service"
)

type fakeService struct {
	createURLFn func(ctx context.Context, url string) (domain.URL, error)
	getURLFn    func(ctx context.Context, code domain.ShortCode) (domain.URL, error)
	listURLsFn  func(ctx context.Context) ([]domain.URL, error)
	updateURLFn func(ctx context.Context, originalURL string, code domain.ShortCode) (domain.URL, error)
	deleteURLFn func(ctx context.Context, code domain.ShortCode) (domain.URL, error)
	redirectFn  func(ctx context.Context, code domain.ShortCode) (string, error)
}

func (f *fakeService) CreateURL(ctx context.Context, url string) (domain.URL, error) {
	if f.createURLFn != nil {
		return f.createURLFn(ctx, url)
	}
	return domain.URL{}, domain.ErrShortCodeCollision
}

func (f *fakeService) GetURL(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	if f.getURLFn != nil {
		return f.getURLFn(ctx, code)
	}
	return domain.URL{}, domain.ErrURLNotFound
}

func (f *fakeService) ListURLs(ctx context.Context) ([]domain.URL, error) {
	if f.listURLsFn != nil {
		return f.listURLsFn(ctx)
	}
	return []domain.URL{}, nil
}

func (f *fakeService) UpdateURL(ctx context.Context, originalURL string, code domain.ShortCode) (domain.URL, error) {
	if f.updateURLFn != nil {
		return f.updateURLFn(ctx, originalURL, code)
	}
	return domain.URL{}, domain.ErrURLNotFound
}

func (f *fakeService) DeleteURL(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
	if f.deleteURLFn != nil {
		return f.deleteURLFn(ctx, code)
	}
	return domain.URL{}, domain.ErrURLNotFound
}

func (f *fakeService) Redirect(ctx context.Context, code domain.ShortCode) (string, error) {
	if f.redirectFn != nil {
		return f.redirectFn(ctx, code)
	}
	return "", domain.ErrURLNotFound
}

type fakePinger struct {
	err error
}

func (f fakePinger) PingContext(ctx context.Context) error {
	return f.err
}

const testShortCode = "abc11111"

var testURL = domain.URL{
	ID:            1,
	ShortCode:     testShortCode,
	OriginalURL:   "https://example.com",
	RedirectCount: 0,
}

var _ service.URLService = (*fakeService)(nil)

func newTestBaseHandler(svc service.URLService, pinger Pinger) *BaseHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewBaseHandler(svc, logger, pinger)
}

func newRequest(method, path string, body io.Reader, withShortCode bool) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), method, path, body)
	if withShortCode {
		r.SetPathValue("shortcode", testShortCode)
	}
	return r
}

func decodeErrorEnvelope(t *testing.T, body io.Reader) map[string]map[string]string {
	t.Helper()
	var envelope map[string]map[string]string
	if err := json.NewDecoder(body).Decode(&envelope); err != nil {
		t.Fatalf("failed to decode error envelope: %v", err)
	}
	return envelope
}

func assertErrorEnvelope(t *testing.T, body io.Reader, wantStatus int, wantCode, wantRequestID string) {
	t.Helper()
	envelope := decodeErrorEnvelope(t, body)
	errBody, ok := envelope["error"]
	if !ok {
		t.Fatal("missing error envelope key")
	}
	if errBody["code"] != wantCode {
		t.Errorf("error.code: got %q, want %q", errBody["code"], wantCode)
	}
	if errBody["message"] == "" {
		t.Error("error.message must not be empty")
	}
	if wantRequestID != "" && errBody["request_id"] != wantRequestID {
		t.Errorf("error.request_id: got %q, want %q", errBody["request_id"], wantRequestID)
	}
	if errBody["request_id"] != wantRequestID {
		t.Errorf("error.request_id: got %q, want %q", errBody["request_id"], wantRequestID)
	}
}

// --- CreateURL ---

func TestCreateURL_Success(t *testing.T) {
	svc := &fakeService{
		createURLFn: func(ctx context.Context, url string) (domain.URL, error) {
			if url != "https://example.com" {
				t.Errorf("CreateURL url: got %q", url)
			}
			return testURL, nil
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{"url":"https://example.com"}`), false)
	w := httptest.NewRecorder()

	h.CreateURL(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusCreated)
	}
	var resp URLResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ShortCode != testShortCode {
		t.Errorf("short_code: got %q, want %q", resp.ShortCode, testShortCode)
	}
	if resp.OriginalURL != testURL.OriginalURL {
		t.Errorf("original_url: got %q, want %q", resp.OriginalURL, testURL.OriginalURL)
	}
	if !strings.HasSuffix(resp.ShortURL, "/"+testShortCode) {
		t.Errorf("short_url: got %q", resp.ShortURL)
	}
}

func TestCreateURL_EmptyBody(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(``), false)
	w := httptest.NewRecorder()

	h.CreateURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

func TestCreateURL_MissingRequiredField(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{}`), false)
	w := httptest.NewRecorder()

	h.CreateURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

func TestCreateURL_InvalidURLFromService(t *testing.T) {
	svc := &fakeService{
		createURLFn: func(ctx context.Context, url string) (domain.URL, error) {
			return domain.URL{}, errors.New("invalid url")
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{"url":"https://example.com"}`), false)
	w := httptest.NewRecorder()

	h.CreateURL(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d (internal errors must be masked)", w.Code, http.StatusInternalServerError)
	}
	assertErrorEnvelope(t, w.Body, http.StatusInternalServerError, errs.CodeInternal, "")
}

func TestCreateURL_Conflict(t *testing.T) {
	svc := &fakeService{
		createURLFn: func(ctx context.Context, url string) (domain.URL, error) {
			return domain.URL{}, domain.ErrShortCodeCollision
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{"url":"https://example.com"}`), false)
	w := httptest.NewRecorder()

	h.CreateURL(w, r)

	if w.Code != http.StatusConflict {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusConflict)
	}
	assertErrorEnvelope(t, w.Body, http.StatusConflict, errs.CodeConflict, "")
}

func TestCreateURL_RequestIDFromContext(t *testing.T) {
	svc := &fakeService{
		createURLFn: func(ctx context.Context, url string) (domain.URL, error) {
			return domain.URL{}, errors.New("boom")
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{"url":"https://example.com"}`), false)
	w := httptest.NewRecorder()

	const rid = "test-request-id"
	ctx := WithRequestID(r.Context(), rid)
	h.CreateURL(w, r.WithContext(ctx))

	assertErrorEnvelope(t, w.Body, http.StatusInternalServerError, errs.CodeInternal, rid)
}

// --- GetURL ---

func TestGetURL_Success(t *testing.T) {
	svc := &fakeService{
		getURLFn: func(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
			return testURL, nil
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodGet, "/api/v1/urls/"+testShortCode, nil, true)
	w := httptest.NewRecorder()

	h.GetURL(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	var resp URLResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ShortCode != testShortCode {
		t.Errorf("short_code: got %q, want %q", resp.ShortCode, testShortCode)
	}
}

func TestGetURL_NotFound(t *testing.T) {
	svc := &fakeService{
		getURLFn: func(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
			return domain.URL{}, domain.ErrURLNotFound
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodGet, "/api/v1/urls/"+testShortCode, nil, true)
	w := httptest.NewRecorder()

	h.GetURL(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
	assertErrorEnvelope(t, w.Body, http.StatusNotFound, errs.CodeNotFound, "")
}

func TestGetURL_InvalidShortCode(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodGet, "/api/v1/urls/invalid", nil, false)
	r.SetPathValue("shortcode", "invalid")
	w := httptest.NewRecorder()

	h.GetURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

// --- UpdateURL ---

func TestUpdateURL_Success(t *testing.T) {
	svc := &fakeService{
		updateURLFn: func(ctx context.Context, originalURL string, code domain.ShortCode) (domain.URL, error) {
			updated := testURL
			updated.OriginalURL = originalURL
			return updated, nil
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodPut, "/api/v1/urls/"+testShortCode, strings.NewReader(`{"url":"https://new.example.com"}`), true)
	w := httptest.NewRecorder()

	h.UpdateURL(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	var resp URLResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.OriginalURL != "https://new.example.com" {
		t.Errorf("original_url: got %q", resp.OriginalURL)
	}
}

func TestUpdateURL_NotFound(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodPut, "/api/v1/urls/"+testShortCode, strings.NewReader(`{"url":"https://new.example.com"}`), true)
	w := httptest.NewRecorder()

	h.UpdateURL(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
	assertErrorEnvelope(t, w.Body, http.StatusNotFound, errs.CodeNotFound, "")
}

// --- DeleteURL ---

func TestDeleteURL_Success(t *testing.T) {
	now := time.Now().UTC()
	svc := &fakeService{
		deleteURLFn: func(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
			deleted := testURL
			deleted.DeletedAt = &now
			return deleted, nil
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodDelete, "/api/v1/urls/"+testShortCode, nil, true)
	w := httptest.NewRecorder()

	h.DeleteURL(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusNoContent)
	}
	if body := w.Body.String(); body != "" {
		t.Errorf("expected empty body, got %q", body)
	}
}

func TestDeleteURL_NotFound(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodDelete, "/api/v1/urls/"+testShortCode, nil, true)
	w := httptest.NewRecorder()

	h.DeleteURL(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
	assertErrorEnvelope(t, w.Body, http.StatusNotFound, errs.CodeNotFound, "")
}

// --- Redirect ---

func TestRedirect_Success(t *testing.T) {
	svc := &fakeService{
		redirectFn: func(ctx context.Context, code domain.ShortCode) (string, error) {
			return "https://example.com/dest", nil
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodGet, "/"+testShortCode, nil, true)
	w := httptest.NewRecorder()

	h.Redirect(w, r)

	if w.Code != http.StatusFound {
		t.Fatalf("status: got %d, want %d (must be 302)", w.Code, http.StatusFound)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com/dest" {
		t.Errorf("Location: got %q", loc)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control: got %q, want no-store", cc)
	}
	if body := w.Body.String(); body != "" {
		t.Errorf("expected empty body, got %q", body)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodGet, "/"+testShortCode, nil, true)
	w := httptest.NewRecorder()

	h.Redirect(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
	assertErrorEnvelope(t, w.Body, http.StatusNotFound, errs.CodeNotFound, "")
}

func TestRedirect_Deleted(t *testing.T) {
	svc := &fakeService{
		redirectFn: func(ctx context.Context, code domain.ShortCode) (string, error) {
			return "", domain.ErrURLAlreadyDeleted
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodGet, "/"+testShortCode, nil, true)
	w := httptest.NewRecorder()

	h.Redirect(w, r)

	if w.Code != http.StatusGone {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusGone)
	}
	assertErrorEnvelope(t, w.Body, http.StatusGone, errs.CodeGone, "")
}

// --- GetStats ---

func TestGetStats_Success(t *testing.T) {
	svc := &fakeService{
		getURLFn: func(ctx context.Context, code domain.ShortCode) (domain.URL, error) {
			u := testURL
			u.RedirectCount = 42
			return u, nil
		},
	}
	h := newTestBaseHandler(svc, fakePinger{})

	r := newRequest(http.MethodGet, "/api/v1/urls/"+testShortCode+"/stats", nil, true)
	w := httptest.NewRecorder()

	h.GetStats(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	var resp StatsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.RedirectCount != 42 {
		t.Errorf("redirect_count: got %d, want 42", resp.RedirectCount)
	}
	if resp.ShortCode != testShortCode {
		t.Errorf("short_code: got %q, want %q", resp.ShortCode, testShortCode)
	}
}

func TestGetStats_NotFound(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodGet, "/api/v1/urls/"+testShortCode+"/stats", nil, true)
	w := httptest.NewRecorder()

	h.GetStats(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
	assertErrorEnvelope(t, w.Body, http.StatusNotFound, errs.CodeNotFound, "")
}

// --- Health ---

func TestHealth_OK(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodGet, "/health", nil, false)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field: got %q, want ok", body["status"])
	}
	if body["database"] != "connected" {
		t.Errorf("database field: got %q, want connected", body["database"])
	}
	if body["timestamp"] == "" {
		t.Error("timestamp must not be empty")
	}
}

func TestHealth_DBUnavailable(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{err: errors.New("connection refused")})

	r := newRequest(http.MethodGet, "/health", nil, false)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	assertErrorEnvelope(t, w.Body, http.StatusServiceUnavailable, errs.CodeServiceUnavailable, "")
}

// --- invalid shortcode across endpoints ---

func TestUpdateURL_InvalidShortCode(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodPut, "/api/v1/urls/invalid", strings.NewReader(`{"url":"https://example.com"}`), false)
	r.SetPathValue("shortcode", "invalid")
	w := httptest.NewRecorder()

	h.UpdateURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

func TestUpdateURL_EmptyBody(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodPut, "/api/v1/urls/"+testShortCode, strings.NewReader(``), true)
	w := httptest.NewRecorder()

	h.UpdateURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

func TestDeleteURL_InvalidShortCode(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodDelete, "/api/v1/urls/invalid", nil, false)
	r.SetPathValue("shortcode", "invalid")
	w := httptest.NewRecorder()

	h.DeleteURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

func TestGetStats_InvalidShortCode(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodGet, "/api/v1/urls/invalid/stats", nil, false)
	r.SetPathValue("shortcode", "invalid")
	w := httptest.NewRecorder()

	h.GetStats(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

func TestRedirect_InvalidShortCode(t *testing.T) {
	h := newTestBaseHandler(&fakeService{}, fakePinger{})

	r := newRequest(http.MethodGet, "/invalid", nil, false)
	r.SetPathValue("shortcode", "invalid")
	w := httptest.NewRecorder()

	h.Redirect(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	assertErrorEnvelope(t, w.Body, http.StatusBadRequest, errs.CodeValidationError, "")
}

// --- shortURL scheme ---

func TestShortURL_HTTPS(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/"+testShortCode, nil)
	r.TLS = &tls.ConnectionState{}
	got := shortURL(r, testShortCode)
	if want := "https://example.com/" + testShortCode; got != want {
		t.Errorf("shortURL: got %q, want %q", got, want)
	}
}

func TestShortURL_HTTP(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/"+testShortCode, nil)
	got := shortURL(r, testShortCode)
	if want := "http://example.com/" + testShortCode; got != want {
		t.Errorf("shortURL: got %q, want %q", got, want)
	}
}

func TestShortURL_ForwardedHeaders(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/"+testShortCode, nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	r.Header.Set("X-Forwarded-Host", "short.example.com")

	got := shortURL(r, testShortCode)
	if want := "https://short.example.com/" + testShortCode; got != want {
		t.Errorf("shortURL: got %q, want %q", got, want)
	}
}
