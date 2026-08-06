package router

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mmk31585/url-shortener/internal/domain"
	"github.com/mmk31585/url-shortener/internal/handler"
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

type fakePinger struct{}

func (fakePinger) PingContext(ctx context.Context) error { return nil }

func newTestMux() *http.ServeMux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewBaseHandler(&fakeService{}, logger, fakePinger{})
	return NewRouter(h)
}

func TestRouter_BindsAllRoutes(t *testing.T) {
	mux := newTestMux()
	cases := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{"create url", http.MethodPost, "/api/v1/urls", `{"url":"https://example.com"}`, http.StatusConflict},
		{"get url", http.MethodGet, "/api/v1/urls/abc11111", "", http.StatusNotFound},
		{"update url", http.MethodPut, "/api/v1/urls/abc11111", `{"url":"https://example.com"}`, http.StatusNotFound},
		{"delete url", http.MethodDelete, "/api/v1/urls/abc11111", "", http.StatusNotFound},
		{"get stats", http.MethodGet, "/api/v1/urls/abc11111/stats", "", http.StatusNotFound},
		{"redirect", http.MethodGet, "/abc11111", "", http.StatusNotFound},
		{"health", http.MethodGet, "/health", "", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			r := httptest.NewRequest(tc.method, tc.path, body)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			if w.Code != tc.want {
				t.Errorf("status: got %d, want %d", w.Code, tc.want)
			}
		})
	}
}

func TestRouter_PathValueReachesHandler(t *testing.T) {
	var gotCode domain.ShortCode
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &fakeService{
		redirectFn: func(ctx context.Context, code domain.ShortCode) (string, error) {
			gotCode = code
			return "https://example.com/dest", nil
		},
	}
	h := handler.NewBaseHandler(svc, logger, fakePinger{})
	mux := NewRouter(h)

	r := httptest.NewRequest(http.MethodGet, "/AbC12345", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusFound)
	}
	if gotCode != "AbC12345" {
		t.Errorf("handler received shortcode %q, want AbC12345", gotCode)
	}
}

func TestRouter_WrongMethodReturns405(t *testing.T) {
	mux := newTestMux()

	r := httptest.NewRequest(http.MethodGet, "/api/v1/urls", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestRouter_UnknownRouteReturns404(t *testing.T) {
	mux := newTestMux()

	r := httptest.NewRequest(http.MethodGet, "/does/not/exist", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRouter_SwaggerUI(t *testing.T) {
	mux := newTestMux()

	r := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("swagger index.html: got %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "swagger") {
		t.Error("expected swagger HTML content")
	}
}

func TestRouter_SwaggerNoSlashRedirects(t *testing.T) {
	mux := newTestMux()

	r := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("GET /swagger: got %d, want %d", w.Code, http.StatusMovedPermanently)
	}
	if loc := w.Header().Get("Location"); loc != "/swagger/" {
		t.Errorf("Location: got %q, want /swagger/", loc)
	}
}

func TestRouter_SwaggerDocJSON(t *testing.T) {
	mux := newTestMux()

	r := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("swagger doc.json: got %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	for _, want := range []string{"swagger", "/api/v1/urls", "/{shortcode}", "/health", "CreateURLPayload", "URLResponse", "ErrorResponse"} {
		if !strings.Contains(body, want) {
			t.Errorf("swagger doc.json missing %q", want)
		}
	}
}
