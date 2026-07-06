package recommendation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"echorec/api/internal/db"
	"echorec/api/internal/redisstore"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	defaultLimit     = 10
	minCatalogYear   = 2015
	maxCatalogYear   = 2026
	genreWeight      = 0.45
	artistWeight     = 0.25
	popularityWeight = 0.20
	freshnessWeight  = 0.10
)

type Recommendation struct {
	TrackID    string  `json:"trackId"`
	Title      string  `json:"title"`
	ArtistID   string  `json:"artistId"`
	ArtistName string  `json:"artistName"`
	Genre      string  `json:"genre"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
}

type Service struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewService(pool *pgxpool.Pool, redisClient *redis.Client) *Service {
	return &Service{db: pool, redis: redisClient}
}

func (s *Service) GetUserFeatures(ctx context.Context, userID string) (redisstore.UserFeatures, error) {
	return redisstore.LoadUserFeatures(ctx, s.redis, userID)
}

func (s *Service) GetRecommendations(ctx context.Context, userID string, limit int) ([]Recommendation, error) {
	if limit <= 0 {
		limit = defaultLimit
	}

	features, err := redisstore.LoadUserFeatures(ctx, s.redis, userID)
	if err != nil {
		return nil, err
	}

	tracks, err := db.ListAllTracks(ctx, s.db)
	if err != nil {
		return nil, err
	}

	candidates := filterCandidates(tracks, features.RecentTracks, limit)
	coldStart := isColdStart(features)

	scored := make([]Recommendation, 0, len(candidates))
	for _, track := range candidates {
		score, reason := scoreTrack(track, features, coldStart)
		scored = append(scored, Recommendation{
			TrackID:    track.ID,
			Title:      track.Title,
			ArtistID:   track.ArtistID,
			ArtistName: track.ArtistName,
			Genre:      track.Genre,
			Score:      roundScore(score),
			Reason:     reason,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	return scored, nil
}

func filterCandidates(tracks []db.Track, recentTracks []string, limit int) []db.Track {
	if len(recentTracks) == 0 || len(tracks)-len(recentTracks) < limit {
		return tracks
	}

	recent := make(map[string]struct{}, len(recentTracks))
	for _, trackID := range recentTracks {
		recent[trackID] = struct{}{}
	}

	filtered := make([]db.Track, 0, len(tracks))
	for _, track := range tracks {
		if _, seen := recent[track.ID]; seen {
			continue
		}
		filtered = append(filtered, track)
	}

	if len(filtered) < limit {
		return tracks
	}

	return filtered
}

func isColdStart(features redisstore.UserFeatures) bool {
	return len(features.GenreScores) == 0 && len(features.ArtistScores) == 0
}

func scoreTrack(track db.Track, features redisstore.UserFeatures, coldStart bool) (float64, string) {
	popularity := float64(track.Popularity) / 100.0
	freshness := freshnessScore(track.ReleaseYear)

	if coldStart {
		score := popularity*popularityWeight + freshness*freshnessWeight
		normalized := score / (popularityWeight + freshnessWeight)
		return normalized, "Popular recent track for new listeners"
	}

	genreAffinity := positiveNormalized(features.GenreScores, track.Genre)
	artistAffinity := positiveNormalized(features.ArtistScores, track.ArtistID)

	score := genreAffinity*genreWeight +
		artistAffinity*artistWeight +
		popularity*popularityWeight +
		freshness*freshnessWeight

	reason := buildReason(track, genreAffinity, artistAffinity, popularity, freshness)
	return score, reason
}

func positiveNormalized(scores map[string]int64, key string) float64 {
	if len(scores) == 0 {
		return 0
	}

	maxPositive := int64(1)
	for _, value := range scores {
		if value > maxPositive {
			maxPositive = value
		}
	}

	raw := scores[key]
	if raw <= 0 {
		return 0
	}

	return float64(raw) / float64(maxPositive)
}

func freshnessScore(releaseYear int) float64 {
	span := float64(maxCatalogYear - minCatalogYear)
	if span <= 0 {
		return 0
	}

	score := float64(releaseYear-minCatalogYear) / span
	return math.Max(0, math.Min(1, score))
}

func buildReason(track db.Track, genreAffinity, artistAffinity, popularity, freshness float64) string {
	type reasonCandidate struct {
		text  string
		value float64
	}

	candidates := []reasonCandidate{
		{
			text:  fmt.Sprintf("Strong match with user's %s preference", track.Genre),
			value: genreAffinity,
		},
		{
			text:  fmt.Sprintf("Matches user's preferred artist %s", track.ArtistName),
			value: artistAffinity,
		},
		{
			text:  "Popular track with strong listener appeal",
			value: popularity,
		},
		{
			text:  "Recent release with high freshness",
			value: freshness,
		},
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].value > candidates[j].value
	})

	parts := make([]string, 0, 2)
	for _, candidate := range candidates {
		if candidate.value <= 0.2 {
			continue
		}
		parts = append(parts, candidate.text)
		if len(parts) == 2 {
			break
		}
	}

	if len(parts) == 0 {
		return "Balanced match across catalog signals"
	}

	return strings.Join(parts, " and ")
}

func roundScore(score float64) float64 {
	return math.Round(score*100) / 100
}
