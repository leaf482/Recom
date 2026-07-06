package recommendation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"echorec/api/internal/db"
	"echorec/api/internal/experiments"
	"echorec/api/internal/redisstore"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	defaultLimit   = 10
	minCatalogYear = 2015
	maxCatalogYear = 2026
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

type RecommendationsResult struct {
	Strategy        string
	Recommendations []Recommendation
}

type Service struct {
	db      *pgxpool.Pool
	redis   *redis.Client
	metrics *experiments.MetricsRecorder
}

func NewService(pool *pgxpool.Pool, redisClient *redis.Client) *Service {
	return &Service{
		db:      pool,
		redis:   redisClient,
		metrics: experiments.NewMetricsRecorder(redisClient),
	}
}

func (s *Service) GetUserFeatures(ctx context.Context, userID string) (redisstore.UserFeatures, error) {
	return redisstore.LoadUserFeatures(ctx, s.redis, userID)
}

func (s *Service) GetRecommendations(ctx context.Context, userID string, limit int) (RecommendationsResult, error) {
	startedAt := time.Now()

	if limit <= 0 {
		limit = defaultLimit
	}

	strategy := experiments.AssignStrategy(userID)
	weights := experiments.WeightsForStrategy(strategy)

	features, err := redisstore.LoadUserFeatures(ctx, s.redis, userID)
	if err != nil {
		return RecommendationsResult{}, err
	}

	tracks, err := db.ListAllTracks(ctx, s.db)
	if err != nil {
		return RecommendationsResult{}, err
	}

	candidates := filterCandidates(tracks, features.RecentTracks, limit)
	coldStart := isColdStart(features)

	scored := make([]Recommendation, 0, len(candidates))
	for _, track := range candidates {
		score, reason := scoreTrack(track, features, coldStart, weights, strategy)
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

	latencyMs := float64(time.Since(startedAt).Microseconds()) / 1000.0
	if err := s.metrics.RecordRecommendation(ctx, experiments.DefaultExperimentID, strategy, len(scored), latencyMs); err != nil {
		return RecommendationsResult{}, fmt.Errorf("record experiment metrics: %w", err)
	}

	return RecommendationsResult{
		Strategy:        strategy,
		Recommendations: scored,
	}, nil
}

func (s *Service) GetExperimentMetrics(ctx context.Context, experimentID string) ([]experiments.StrategyMetrics, error) {
	return s.metrics.GetMetrics(ctx, experimentID)
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

func scoreTrack(track db.Track, features redisstore.UserFeatures, coldStart bool, weights experiments.Weights, strategy string) (float64, string) {
	popularity := float64(track.Popularity) / 100.0
	freshness := freshnessScore(track.ReleaseYear)

	if coldStart {
		discoveryWeight := weights.Popularity + weights.Freshness
		score := popularity*weights.Popularity + freshness*weights.Freshness
		if discoveryWeight > 0 {
			score /= discoveryWeight
		}
		return score, "Popular recent track for new listeners"
	}

	genreAffinity := positiveNormalized(features.GenreScores, track.Genre)
	artistAffinity := positiveNormalized(features.ArtistScores, track.ArtistID)

	score := genreAffinity*weights.Genre +
		artistAffinity*weights.Artist +
		popularity*weights.Popularity +
		freshness*weights.Freshness

	reason := buildReason(track, genreAffinity, artistAffinity, popularity, freshness, strategy)
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

func buildReason(track db.Track, genreAffinity, artistAffinity, popularity, freshness float64, strategy string) string {
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

	if strategy == experiments.StrategyExploration {
		candidates[2].value += 0.15
		candidates[3].value += 0.15
	}
	if strategy == experiments.StrategyGenreAffinity {
		candidates[0].value += 0.15
	}
	if strategy == experiments.StrategyArtistAffinity {
		candidates[1].value += 0.15
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
