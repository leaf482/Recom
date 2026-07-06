package events

import "time"

const (
	EventTypePlay   = "play"
	EventTypeSkip   = "skip"
	EventTypeLike   = "like"
	EventTypeSave   = "save"
	EventTypeReplay = "replay"
)

var EventTypes = []string{
	EventTypePlay,
	EventTypeSkip,
	EventTypeLike,
	EventTypeSave,
	EventTypeReplay,
}

type ListeningEvent struct {
	EventID    string `json:"eventId"`
	UserID     string `json:"userId"`
	TrackID    string `json:"trackId"`
	EventType  string `json:"eventType"`
	Genre      string `json:"genre"`
	ArtistID   string `json:"artistId"`
	DurationMs int    `json:"durationMs"`
	Timestamp  string `json:"timestamp"`
}

func NewListeningEvent(eventID, userID, trackID, eventType, genre, artistID string, durationMs int, timestamp time.Time) ListeningEvent {
	return ListeningEvent{
		EventID:    eventID,
		UserID:     userID,
		TrackID:    trackID,
		EventType:  eventType,
		Genre:      genre,
		ArtistID:   artistID,
		DurationMs: durationMs,
		Timestamp:  timestamp.UTC().Format(time.RFC3339),
	}
}
