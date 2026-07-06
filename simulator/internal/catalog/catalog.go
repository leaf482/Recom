package catalog

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Track struct {
	ID       string
	Genre    string
	ArtistID string
}

func LoadTracks(ctx context.Context, pool *pgxpool.Pool) ([]Track, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, genre, artist_id
		FROM tracks
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query tracks: %w", err)
	}
	defer rows.Close()

	tracks := make([]Track, 0, 100)
	for rows.Next() {
		var track Track
		if err := rows.Scan(&track.ID, &track.Genre, &track.ArtistID); err != nil {
			return nil, fmt.Errorf("scan track: %w", err)
		}
		tracks = append(tracks, track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tracks: %w", err)
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found in catalog")
	}

	return tracks, nil
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
