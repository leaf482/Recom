package features

import (
	"context"
	"fmt"

	"echorec/events"

	"github.com/redis/go-redis/v9"
)

const recentTrackLimit = 20

type Updater struct {
	redis *redis.Client
}

func NewUpdater(client *redis.Client) *Updater {
	return &Updater{redis: client}
}

type UpdateResult struct {
	GenreScore   int64
	ArtistScore  int64
	EventCount   int64
	RecentTracks int64
}

func (u *Updater) ApplyEvent(ctx context.Context, event events.ListeningEvent) (UpdateResult, error) {
	delta, err := ScoreDelta(event.EventType)
	if err != nil {
		return UpdateResult{}, err
	}

	genreKey := fmt.Sprintf("user:%s:genre_score", event.UserID)
	artistKey := fmt.Sprintf("user:%s:artist_score", event.UserID)
	recentKey := fmt.Sprintf("user:%s:recent_tracks", event.UserID)
	countsKey := fmt.Sprintf("user:%s:event_counts", event.UserID)

	pipe := u.redis.Pipeline()
	genreCmd := pipe.HIncrBy(ctx, genreKey, event.Genre, int64(delta))
	artistCmd := pipe.HIncrBy(ctx, artistKey, event.ArtistID, int64(delta))
	countCmd := pipe.HIncrBy(ctx, countsKey, event.EventType, 1)
	pipe.LPush(ctx, recentKey, event.TrackID)
	pipe.LTrim(ctx, recentKey, 0, recentTrackLimit-1)

	if _, err := pipe.Exec(ctx); err != nil {
		return UpdateResult{}, fmt.Errorf("update redis features: %w", err)
	}

	recentTracks, err := u.redis.LLen(ctx, recentKey).Result()
	if err != nil {
		return UpdateResult{}, fmt.Errorf("read recent tracks length: %w", err)
	}

	return UpdateResult{
		GenreScore:   genreCmd.Val(),
		ArtistScore:  artistCmd.Val(),
		EventCount:   countCmd.Val(),
		RecentTracks: recentTracks,
	}, nil
}
