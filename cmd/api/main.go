package main

import (
	"log"
	"os"

	"github.com/mmk31585/url-shortener/internal/config"
	"github.com/mmk31585/url-shortener/internal/handler"
	"github.com/mmk31585/url-shortener/internal/logger"
	"github.com/mmk31585/url-shortener/internal/repository/postgres"
	"github.com/mmk31585/url-shortener/internal/service"
	"github.com/mmk31585/url-shortener/internal/shortener"
	"github.com/mmk31585/url-shortener/internal/storage"
)

func main() {
	log.Print("Starting Server...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.New(cfg.AppEnv)
	logger.Info("starting server", "address", cfg.ServerAddress, "env", cfg.AppEnv)

	store, err := storage.New(cfg.DB)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	repo := postgres.NewPostgresURLRepository(store.DB())
	shortener := shortener.NewRandomShortener()
	srv := service.NewURLService(repo, shortener)
	handler.NewBaseHandler(srv, logger, store.DB())
}
