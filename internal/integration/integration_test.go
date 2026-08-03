package integration_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mmk31585/url-shortener/internal/handler"
	"github.com/mmk31585/url-shortener/internal/middleware"
	"github.com/mmk31585/url-shortener/internal/repository/mock"
	"github.com/mmk31585/url-shortener/internal/router"
	"github.com/mmk31585/url-shortener/internal/service"
	"github.com/mmk31585/url-shortener/internal/shortener"
)

type fakePinger struct{ err error }

func (f fakePinger) PingContext(ctx context.Context) error { return f.err }

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	repo := mock.NewURLRepository()
	srv := service.NewURLService(repo, shortener.NewRandomShortener())
	h := handler.NewBaseHandler(srv, logger, fakePinger{})
	mux := router.NewRouter(h)

	stack := middleware.Recovery(logger)(middleware.RequestID(middleware.Logging(logger)(mux)))
	ts := httptest.NewServer(stack)
	t.Cleanup(ts.Close)
	return ts
}

type httpClient struct {
	t      *testing.T
	client *http.Client
	host   string
}

func newHTTPClient(t *testing.T, ts *httptest.Server) *httpClient {
	t.Helper()
	return &httpClient{
		t:    t,
		host: ts.URL,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: 5 * time.Second,
		},
	}
}

func (c *httpClient) request(method, path string, body string) *http.Response {
	c.t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, c.host+path, r)
	if err != nil {
		c.t.Fatalf("NewRequest: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("request %s %s: %v", method, path, err)
	}
	return resp
}

func (c *httpClient) createURL(url string) (int, map[string]any) {
	c.t.Helper()
	resp := c.request(http.MethodPost, "/api/v1/urls", `{"url":"`+url+`"}`)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(body, &result)
	return resp.StatusCode, result
}

func (c *httpClient) getURL(code string) (int, map[string]any) {
	c.t.Helper()
	resp := c.request(http.MethodGet, "/api/v1/urls/"+code, "")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(body, &result)
	return resp.StatusCode, result
}

func (c *httpClient) redirect(code string) (int, string, string, string) {
	c.t.Helper()
	resp := c.request(http.MethodGet, "/"+code, "")
	defer resp.Body.Close()
	return resp.StatusCode,
		resp.Header.Get("Location"),
		resp.Header.Get("Cache-Control"),
		resp.Header.Get("X-Request-ID")
}

func (c *httpClient) updateURL(code, url string) (int, map[string]any) {
	c.t.Helper()
	resp := c.request(http.MethodPut, "/api/v1/urls/"+code, `{"url":"`+url+`"}`)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(body, &result)
	return resp.StatusCode, result
}

func (c *httpClient) deleteURL(code string) int {
	c.t.Helper()
	resp := c.request(http.MethodDelete, "/api/v1/urls/"+code, "")
	defer resp.Body.Close()
	return resp.StatusCode
}

func (c *httpClient) stats(code string) (int, map[string]any) {
	c.t.Helper()
	resp := c.request(http.MethodGet, "/api/v1/urls/"+code+"/stats", "")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(body, &result)
	return resp.StatusCode, result
}

func (c *httpClient) health() (int, map[string]any) {
	c.t.Helper()
	resp := c.request(http.MethodGet, "/health", "")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(body, &result)
	return resp.StatusCode, result
}

func (c *httpClient) errorEnvelope(resp *http.Response) map[string]string {
	c.t.Helper()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var envelope map[string]map[string]string
	json.Unmarshal(body, &envelope)
	return envelope["error"]
}

func assertStatus(t *testing.T, got, want int, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", msg, got, want)
	}
}

func assertString(t *testing.T, got, want, field string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

func assertNotEmpty(t *testing.T, val, field string) {
	t.Helper()
	if val == "" {
		t.Errorf("%s must not be empty", field)
	}
}

func TestCurl_CreateURL_Success(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	status, body := client.createURL("https://example.com/long/path")
	assertStatus(t, status, http.StatusCreated, "create status")

	if body["short_code"] == nil || body["short_code"] == "" {
		t.Fatal("short_code missing")
	}
	if body["original_url"] != "https://example.com/long/path" {
		t.Errorf("original_url: got %v", body["original_url"])
	}
	if body["redirect_count"] != float64(0) {
		t.Errorf("redirect_count: got %v", body["redirect_count"])
	}
	if body["created_at"] == nil {
		t.Error("created_at missing")
	}
	if body["updated_at"] == nil {
		t.Error("updated_at missing")
	}
}

func TestCurl_CreateURL_MissingBody(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodPost, "/api/v1/urls", "")
	defer resp.Body.Close()
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "missing body")
	assertNotEmpty(t, resp.Header.Get("X-Request-ID"), "X-Request-ID")
}

func TestCurl_CreateURL_EmptyJSON(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodPost, "/api/v1/urls", "{}")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "empty json")
	assertString(t, errBody["code"], "VALIDATION_ERROR", "error.code")
}

func TestCurl_CreateURL_InvalidJSON(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodPost, "/api/v1/urls", `{"url":`)
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "invalid json")
	assertString(t, errBody["code"], "VALIDATION_ERROR", "error.code")
}

func TestCurl_CreateURL_UnknownField(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodPost, "/api/v1/urls", `{"url":"https://x.com","extra":1}`)
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "unknown field")
	assertString(t, errBody["code"], "VALIDATION_ERROR", "error.code")
}

func TestCurl_CreateURL_InvalidURL(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodPost, "/api/v1/urls", `{"url":"not-a-url"}`)
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "invalid url")
	assertString(t, errBody["code"], "VALIDATION_ERROR", "error.code")
}

func TestCurl_CreateURL_URLTooLong(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	longURL := "https://example.com/" + strings.Repeat("a", 3000)
	status, _ := client.createURL(longURL)
	assertStatus(t, status, http.StatusUnprocessableEntity, "url too long")
}

func TestCurl_GetURL_Success(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/get")
	code := created["short_code"].(string)

	status, body := client.getURL(code)
	assertStatus(t, status, http.StatusOK, "get status")
	assertString(t, body["short_code"].(string), code, "short_code")
	assertString(t, body["original_url"].(string), "https://example.com/get", "original_url")
}

func TestCurl_GetURL_NotFound(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/api/v1/urls/xxxxxxxx", "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "not found")
	assertString(t, errBody["code"], "NOT_FOUND", "error.code")
}

func TestCurl_GetURL_InvalidCode(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/api/v1/urls/short", "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "invalid code")
	assertString(t, errBody["code"], "VALIDATION_ERROR", "error.code")
}

func TestCurl_Redirect_Success(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/dest")
	code := created["short_code"].(string)

	status, loc, cacheCtrl, requestID := client.redirect(code)
	assertStatus(t, status, http.StatusFound, "redirect status")
	assertString(t, loc, "https://example.com/dest", "Location")
	assertString(t, cacheCtrl, "no-store", "Cache-Control")
	assertNotEmpty(t, requestID, "X-Request-ID")
}

func TestCurl_Redirect_NotFound(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/xxxxxxxx", "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "redirect not found")
	assertString(t, errBody["code"], "NOT_FOUND", "error.code")
}

func TestCurl_Redirect_Deleted(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/will-delete")
	code := created["short_code"].(string)

	status := client.deleteURL(code)
	assertStatus(t, status, http.StatusNoContent, "delete status")

	// Both mock and postgres repos filter deleted rows -> ErrURLNotFound (404)
	resp := client.request(http.MethodGet, "/"+code, "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "redirect deleted")
	assertString(t, errBody["code"], "NOT_FOUND", "error.code")
}

func TestCurl_Redirect_InvalidCode(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/short", "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "invalid code")
	assertString(t, errBody["code"], "VALIDATION_ERROR", "error.code")
}

func TestCurl_UpdateURL_Success(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/old")
	code := created["short_code"].(string)

	status, body := client.updateURL(code, "https://example.com/new")
	assertStatus(t, status, http.StatusOK, "update status")
	assertString(t, body["original_url"].(string), "https://example.com/new", "original_url")
}

func TestCurl_UpdateURL_NotFound(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodPut, "/api/v1/urls/xxxxxxxx", `{"url":"https://x.com"}`)
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "not found")
	assertString(t, errBody["code"], "NOT_FOUND", "error.code")
}

func TestCurl_UpdateURL_EmptyBody(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/has-body")
	code := created["short_code"].(string)

	resp := client.request(http.MethodPut, "/api/v1/urls/"+code, "{}")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusBadRequest, "empty body")
	assertString(t, errBody["code"], "VALIDATION_ERROR", "error.code")
}

func TestCurl_DeleteURL_Success(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/del")
	code := created["short_code"].(string)

	status := client.deleteURL(code)
	assertStatus(t, status, http.StatusNoContent, "delete status")
}

func TestCurl_DeleteURL_NotFound(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodDelete, "/api/v1/urls/xxxxxxxx", "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "not found")
	assertString(t, errBody["code"], "NOT_FOUND", "error.code")
}

func TestCurl_Stats_Success(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/stats")
	code := created["short_code"].(string)

	status, body := client.stats(code)
	assertStatus(t, status, http.StatusOK, "stats status")
	assertString(t, body["short_code"].(string), code, "short_code")
	if body["redirect_count"] != float64(0) {
		t.Errorf("redirect_count: got %v, want 0", body["redirect_count"])
	}
}

func TestCurl_Stats_NotFound(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/api/v1/urls/xxxxxxxx/stats", "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "not found")
	assertString(t, errBody["code"], "NOT_FOUND", "error.code")
}

func TestCurl_Stats_AfterRedirects(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	_, created := client.createURL("https://example.com/track")
	code := created["short_code"].(string)

	for i := 0; i < 3; i++ {
		status, _, _, _ := client.redirect(code)
		assertStatus(t, status, http.StatusFound, "redirect status")
	}

	status, body := client.stats(code)
	assertStatus(t, status, http.StatusOK, "stats status")
	if body["redirect_count"] != float64(3) {
		t.Errorf("redirect_count: got %v, want 3", body["redirect_count"])
	}
}

func TestCurl_Health_Success(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	status, body := client.health()
	assertStatus(t, status, http.StatusOK, "health status")
	assertString(t, body["status"].(string), "ok", "status")
	assertString(t, body["database"].(string), "connected", "database")
	assertNotEmpty(t, body["timestamp"].(string), "timestamp")
}

func TestCurl_Health_DBUnavailable(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := mock.NewURLRepository()
	srv := service.NewURLService(repo, shortener.NewRandomShortener())
	h := handler.NewBaseHandler(srv, logger, fakePinger{err: context.DeadlineExceeded})
	mux := router.NewRouter(h)
	stack := middleware.Recovery(logger)(middleware.RequestID(middleware.Logging(logger)(mux)))
	ts := httptest.NewServer(stack)
	t.Cleanup(ts.Close)

	client := newHTTPClient(t, ts)

	status, body := client.health()
	assertStatus(t, status, http.StatusServiceUnavailable, "health unavailable")
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error envelope")
	}
	assertString(t, errObj["code"].(string), "SERVICE_UNAVAILABLE", "error.code")
}

func TestCurl_ErrorEnvelopeFormat(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create missing body", http.MethodPost, "/api/v1/urls", ""},
		{"create empty json", http.MethodPost, "/api/v1/urls", "{}"},
		{"get not found", http.MethodGet, "/api/v1/urls/xxxxxxxx", ""},
		{"get invalid code", http.MethodGet, "/api/v1/urls/short", ""},
		{"update not found", http.MethodPut, "/api/v1/urls/xxxxxxxx", `{"url":"https://x.com"}`},
		{"delete not found", http.MethodDelete, "/api/v1/urls/xxxxxxxx", ""},
		{"stats not found", http.MethodGet, "/api/v1/urls/xxxxxxxx/stats", ""},
		{"redirect not found", http.MethodGet, "/xxxxxxxx", ""},
		{"redirect invalid code", http.MethodGet, "/short", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := client.request(tc.method, tc.path, tc.body)
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var envelope map[string]map[string]string
			if err := json.Unmarshal(body, &envelope); err != nil {
				t.Fatalf("failed to decode envelope: %v (body: %s)", err, body)
			}

			errObj, ok := envelope["error"]
			if !ok {
				t.Fatalf("missing error key in envelope: %s", body)
			}
			assertNotEmpty(t, errObj["code"], "error.code")
			assertNotEmpty(t, errObj["message"], "error.message")
			assertNotEmpty(t, errObj["request_id"], "error.request_id")

			headerID := resp.Header.Get("X-Request-ID")
			assertString(t, errObj["request_id"], headerID, "request_id match")
		})
	}
}

func TestCurl_RequestID_Propagated(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	// Client sends X-Request-ID
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/health", nil)
	req.Header.Set("X-Request-ID", "my-custom-id-123")
	resp, err := client.client.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	headerID := resp.Header.Get("X-Request-ID")
	assertString(t, headerID, "my-custom-id-123", "echoed X-Request-ID")
}

func TestCurl_RequestID_GeneratedWhenOmitted(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/health", "")
	defer resp.Body.Close()

	requestID := resp.Header.Get("X-Request-ID")
	assertNotEmpty(t, requestID, "X-Request-ID")
	if len(requestID) != 36 {
		t.Errorf("expected UUID v4 length 36, got %d: %q", len(requestID), requestID)
	}
}

func TestCurl_WrongMethod(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/v1/urls", http.StatusMethodNotAllowed},
		{http.MethodPost, "/xxxxxxxx", http.StatusMethodNotAllowed},
		{http.MethodPut, "/xxxxxxxx", http.StatusMethodNotAllowed},
		{http.MethodDelete, "/xxxxxxxx", http.StatusMethodNotAllowed},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{http.MethodPut, "/api/v1/urls/xxxxxxxx", http.StatusBadRequest},
		{http.MethodDelete, "/api/v1/urls/xxxxxxxx", http.StatusNotFound},
		{http.MethodPost, "/api/v1/urls/xxxxxxxx/stats", http.StatusMethodNotAllowed},
		{http.MethodPut, "/api/v1/urls/xxxxxxxx/stats", http.StatusMethodNotAllowed},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			resp := client.request(tc.method, tc.path, "")
			defer resp.Body.Close()
			assertStatus(t, resp.StatusCode, tc.want, "wrong method")
		})
	}
}

func TestCurl_UnknownRoute(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/does/not/exist", "")
	defer resp.Body.Close()
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "unknown route")
}

func TestCurl_FullWorkflow(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	// Create
	status, created := client.createURL("https://example.com/workflow")
	assertStatus(t, status, http.StatusCreated, "create")
	code := created["short_code"].(string)

	// Get
	status, got := client.getURL(code)
	assertStatus(t, status, http.StatusOK, "get")
	assertString(t, got["original_url"].(string), "https://example.com/workflow", "original_url")

	// Redirect
	status, loc, _, _ := client.redirect(code)
	assertStatus(t, status, http.StatusFound, "redirect")
	assertString(t, loc, "https://example.com/workflow", "Location")

	// Stats after 1 redirect
	status, stats := client.stats(code)
	assertStatus(t, status, http.StatusOK, "stats")
	if stats["redirect_count"] != float64(1) {
		t.Errorf("redirect_count: got %v, want 1", stats["redirect_count"])
	}

	// Update
	status, updated := client.updateURL(code, "https://example.com/updated")
	assertStatus(t, status, http.StatusOK, "update")
	assertString(t, updated["original_url"].(string), "https://example.com/updated", "updated url")

	// Get after update
	status, got = client.getURL(code)
	assertStatus(t, status, http.StatusOK, "get after update")
	assertString(t, got["original_url"].(string), "https://example.com/updated", "updated url")

	// Redirect to new destination
	status, loc, _, _ = client.redirect(code)
	assertStatus(t, status, http.StatusFound, "redirect after update")
	assertString(t, loc, "https://example.com/updated", "Location after update")

	// Stats after 2 redirects
	status, stats = client.stats(code)
	assertStatus(t, status, http.StatusOK, "stats after 2")
	if stats["redirect_count"] != float64(2) {
		t.Errorf("redirect_count: got %v, want 2", stats["redirect_count"])
	}

	// Delete
	status = client.deleteURL(code)
	assertStatus(t, status, http.StatusNoContent, "delete")

	// Get after delete -> 404
	status, _ = client.getURL(code)
	assertStatus(t, status, http.StatusNotFound, "get after delete")

	// Redirect after delete -> 404 (repos filter deleted rows)
	resp := client.request(http.MethodGet, "/"+code, "")
	errBody := client.errorEnvelope(resp)
	assertStatus(t, resp.StatusCode, http.StatusNotFound, "redirect after delete")
	assertString(t, errBody["code"], "NOT_FOUND", "error.code after delete")

	// Stats after delete -> 404
	status, _ = client.stats(code)
	assertStatus(t, status, http.StatusNotFound, "stats after delete")
}

func TestCurl_HealthResponseFormat(t *testing.T) {
	ts := newTestServer(t)
	client := newHTTPClient(t, ts)

	resp := client.request(http.MethodGet, "/health", "")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	assertStatus(t, resp.StatusCode, http.StatusOK, "health status")
	assertString(t, result["status"], "ok", "status")
	assertString(t, result["database"], "connected", "database")
	assertNotEmpty(t, result["timestamp"], "timestamp")
	assertNotEmpty(t, resp.Header.Get("X-Request-ID"), "X-Request-ID")
}
