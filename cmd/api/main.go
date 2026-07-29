package main

import (
	"log"

	"github.com/mmk31585/url-shortener/internal/config"
	"github.com/mmk31585/url-shortener/internal/logger"
)

func main() {
	log.Print("Starting Server...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger.New(cfg.AppEnv)
}
