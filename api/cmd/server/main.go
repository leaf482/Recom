package main

import (
	"context"
	"log"
	"net/http"

	"echorec/api/internal/config"
	"echorec/api/internal/db"
	httpapi "echorec/api/internal/http"
	"echorec/api/internal/recommendation"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	recommendationService := recommendation.NewService(pool, redisClient)
	server := &httpapi.Server{
		DB:              pool,
		Recommendations: recommendationService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.Health)
	mux.HandleFunc("GET /tracks", server.ListTracks)
	mux.HandleFunc("GET /users/{userId}/features", server.GetUserFeatures)
	mux.HandleFunc("GET /users/{userId}/recommendations", server.GetUserRecommendations)

	log.Printf("starting API on %s", cfg.APIAddr)
	if err := http.ListenAndServe(cfg.APIAddr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
