package main

// @title           URL Shortener API
// @version         1.0
// @description     A high-performance URL shortener service with statistics tracking
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @schemes http https

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmk31585/url-shortener/internal/config"
	"github.com/mmk31585/url-shortener/internal/handler"
	"github.com/mmk31585/url-shortener/internal/logger"
	"github.com/mmk31585/url-shortener/internal/middleware"
	"github.com/mmk31585/url-shortener/internal/repository/postgres"
	"github.com/mmk31585/url-shortener/internal/router"
	"github.com/mmk31585/url-shortener/internal/service"
	"github.com/mmk31585/url-shortener/internal/shortener"
	"github.com/mmk31585/url-shortener/internal/storage"
	"github.com/mmk31585/url-shortener/internal/validator"
)

func main() {
	log.Print("Starting Server...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.New(cfg.AppEnv)
	logger.Info("starting server", "address", cfg.ServerAddress, "env", cfg.AppEnv)

	validator.SetMaxURLLength(cfg.MaxURL)

	store, err := storage.New(cfg.DB)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	repo := postgres.NewPostgresURLRepository(store.DB())
	shortener := shortener.NewRandomShortener()
	srv := service.NewURLService(repo, shortener)

	h := handler.NewBaseHandler(srv, logger, store.DB())
	mux := router.NewRouter(h)

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           middleware.Recovery(logger)(middleware.RequestID(middleware.Logging(logger)(mux))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}
