CREATE TABLE IF NOT EXISTS tracks (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    artist_id    TEXT NOT NULL,
    artist_name  TEXT NOT NULL,
    genre        TEXT NOT NULL,
    popularity   INTEGER NOT NULL CHECK (popularity >= 0 AND popularity <= 100),
    release_year INTEGER NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tracks_genre ON tracks (genre);
CREATE INDEX IF NOT EXISTS idx_tracks_artist_id ON tracks (artist_id);
