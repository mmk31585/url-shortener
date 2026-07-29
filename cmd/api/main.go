package main

import (
	"log"

	"github.com/mmk31585/url-shortener/internal/config"
)

func main() {
	log.Print("Starting Server...")
	_, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
}
