package redisstore

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type UserFeatures struct {
	UserID       string
	GenreScores  map[string]int64
	ArtistScores map[string]int64
	RecentTracks []string
	EventCounts  map[string]int64
}

func LoadUserFeatures(ctx context.Context, client *redis.Client, userID string) (UserFeatures, error) {
	genreKey := fmt.Sprintf("user:%s:genre_score", userID)
	artistKey := fmt.Sprintf("user:%s:artist_score", userID)
	recentKey := fmt.Sprintf("user:%s:recent_tracks", userID)
	countsKey := fmt.Sprintf("user:%s:event_counts", userID)

	genreScores, err := client.HGetAll(ctx, genreKey).Result()
	if err != nil {
		return UserFeatures{}, fmt.Errorf("load genre scores: %w", err)
	}

	artistScores, err := client.HGetAll(ctx, artistKey).Result()
	if err != nil {
		return UserFeatures{}, fmt.Errorf("load artist scores: %w", err)
	}

	recentTracks, err := client.LRange(ctx, recentKey, 0, -1).Result()
	if err != nil {
		return UserFeatures{}, fmt.Errorf("load recent tracks: %w", err)
	}

	eventCounts, err := client.HGetAll(ctx, countsKey).Result()
	if err != nil {
		return UserFeatures{}, fmt.Errorf("load event counts: %w", err)
	}

	return UserFeatures{
		UserID:       userID,
		GenreScores:  parseIntMap(genreScores),
		ArtistScores: parseIntMap(artistScores),
		RecentTracks: recentTracks,
		EventCounts:  parseIntMap(eventCounts),
	}, nil
}

func parseIntMap(values map[string]string) map[string]int64 {
	result := make(map[string]int64, len(values))
	for key, value := range values {
		var parsed int64
		if _, err := fmt.Sscan(value, &parsed); err == nil {
			result[key] = parsed
		}
	}
	return result
}
