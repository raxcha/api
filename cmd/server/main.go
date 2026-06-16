package main

import (
	"api"
	"log"
	"os"
)

func main() {
	cfg := api.Config{
		Addr:          envOr("API_ADDR", "127.0.0.1:8787"),
		Root:          envOr("API_ROOT", "/home/asdf/api/prsnl.spc"),
		Token:         os.Getenv("API_TOKEN"),
		FrontUser:     envOr("FRONT_USER", "prsnl"),
		FrontPassword: os.Getenv("FRONT_PASSWORD"),
	}

	if cfg.Token == "" {
		log.Fatal("API_TOKEN is required")
	}
	if cfg.FrontPassword == "" {
		log.Fatal("FRONT_PASSWORD is required")
	}

	if err := os.MkdirAll(cfg.Root, 0755); err != nil {
		log.Fatal(err)
	}

	log.Printf("api listening on %s, root %s", cfg.Addr, cfg.Root)
	log.Fatal(api.NewServer(cfg).ListenAndServe())
}

func envOr(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
