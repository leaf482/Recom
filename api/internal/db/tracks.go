package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Track struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ArtistID    string    `json:"artistId"`
	ArtistName  string    `json:"artistName"`
	Genre       string    `json:"genre"`
	Popularity  int       `json:"popularity"`
	ReleaseYear int       `json:"releaseYear"`
	CreatedAt   time.Time `json:"createdAt"`
}

func ListTracks(ctx context.Context, pool *pgxpool.Pool, limit int) ([]Track, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, title, artist_id, artist_name, genre, popularity, release_year, created_at
		FROM tracks
		ORDER BY id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query tracks: %w", err)
	}
	defer rows.Close()

	tracks := make([]Track, 0, limit)
	for rows.Next() {
		var track Track
		if err := rows.Scan(
			&track.ID,
			&track.Title,
			&track.ArtistID,
			&track.ArtistName,
			&track.Genre,
			&track.Popularity,
			&track.ReleaseYear,
			&track.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan track: %w", err)
		}
		tracks = append(tracks, track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tracks: %w", err)
	}

	return tracks, nil
}
