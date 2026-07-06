package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"echorec/api/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DB *pgxpool.Pool
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (s *Server) ListTracks(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			http.Error(w, "limit must be a positive integer", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	tracks, err := db.ListTracks(r.Context(), s.DB, limit)
	if err != nil {
		log.Printf("list tracks: %v", err)
		http.Error(w, "failed to load tracks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"count":  len(tracks),
		"tracks": tracks,
	}); err != nil {
		log.Printf("encode tracks response: %v", err)
	}
}
