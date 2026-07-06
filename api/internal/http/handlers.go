package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"echorec/api/internal/db"
	"echorec/api/internal/recommendation"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DB              *pgxpool.Pool
	Recommendations *recommendation.Service
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

	writeJSON(w, map[string]any{
		"count":  len(tracks),
		"tracks": tracks,
	})
}

func (s *Server) GetUserFeatures(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	features, err := s.Recommendations.GetUserFeatures(r.Context(), userID)
	if err != nil {
		log.Printf("get user features: %v", err)
		http.Error(w, "failed to load user features", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"userId":       userID,
		"genreScores":  features.GenreScores,
		"artistScores": features.ArtistScores,
		"recentTracks": features.RecentTracks,
		"eventCounts":  features.EventCounts,
	})
}

func (s *Server) GetUserRecommendations(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			http.Error(w, "limit must be a positive integer", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	recommendations, err := s.Recommendations.GetRecommendations(r.Context(), userID, limit)
	if err != nil {
		log.Printf("get recommendations: %v", err)
		http.Error(w, "failed to load recommendations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"userId":          userID,
		"strategy":        recommendations.Strategy,
		"recommendations": recommendations.Recommendations,
	})
}

func (s *Server) GetExperimentMetrics(w http.ResponseWriter, r *http.Request) {
	experimentID := r.PathValue("experimentId")
	if experimentID == "" {
		http.Error(w, "experimentId is required", http.StatusBadRequest)
		return
	}

	metrics, err := s.Recommendations.GetExperimentMetrics(r.Context(), experimentID)
	if err != nil {
		log.Printf("get experiment metrics: %v", err)
		http.Error(w, "failed to load experiment metrics", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"experimentId": experimentID,
		"strategies":   metrics,
	})
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode json response: %v", err)
	}
}
