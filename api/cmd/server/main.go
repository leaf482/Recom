package main

import (
	"context"
	"log"
	"net/http"

	"echorec/api/internal/config"
	"echorec/api/internal/db"
	httpapi "echorec/api/internal/http"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	server := &httpapi.Server{DB: pool}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.Health)
	mux.HandleFunc("GET /tracks", server.ListTracks)

	log.Printf("starting API on %s", cfg.APIAddr)
	if err := http.ListenAndServe(cfg.APIAddr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
