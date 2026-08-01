package router

import (
	"net/http"

	"github.com/mmk31585/url-shortener/internal/handler"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/mmk31585/url-shortener/docs"
)

func NewRouter(h *handler.BaseHandler) *http.ServeMux {
	mux := http.NewServeMux()

	api := http.NewServeMux()
	api.HandleFunc("POST /urls", h.CreateURL)
	api.HandleFunc("GET /urls/{shortcode}", h.GetURL)
	api.HandleFunc("PUT /urls/{shortcode}", h.UpdateURL)
	api.HandleFunc("DELETE /urls/{shortcode}", h.DeleteURL)
	api.HandleFunc("GET /urls/{shortcode}/stats", h.GetStats)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	mux.HandleFunc("GET /{shortcode}", h.Redirect)
	mux.HandleFunc("GET /health", h.Health)

	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}
