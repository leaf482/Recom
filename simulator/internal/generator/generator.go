package generator

import (
	"fmt"
	"math/rand"
	"time"

	"echorec/events"
	"echorec/simulator/internal/catalog"
)

type Generator struct {
	rng       *rand.Rand
	tracks    []catalog.Track
	userCount int
	sequence  uint64
}

func New(tracks []catalog.Track, userCount int, seed int64) *Generator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Generator{
		rng:       rand.New(rand.NewSource(seed)),
		tracks:    tracks,
		userCount: userCount,
	}
}

func (g *Generator) Next() events.ListeningEvent {
	g.sequence++

	track := g.tracks[g.rng.Intn(len(g.tracks))]
	userID := fmt.Sprintf("user_%d", 1+g.rng.Intn(g.userCount))
	eventType := weightedEventType(g.rng)
	durationMs := durationForEventType(g.rng, eventType)
	eventID := fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), g.sequence)

	return events.NewListeningEvent(
		eventID,
		userID,
		track.ID,
		eventType,
		track.Genre,
		track.ArtistID,
		durationMs,
		time.Now().UTC(),
	)
}

func weightedEventType(rng *rand.Rand) string {
	roll := rng.Intn(100)
	switch {
	case roll < 50:
		return events.EventTypePlay
	case roll < 75:
		return events.EventTypeSkip
	case roll < 85:
		return events.EventTypeLike
	case roll < 95:
		return events.EventTypeSave
	default:
		return events.EventTypeReplay
	}
}

func durationForEventType(rng *rand.Rand, eventType string) int {
	switch eventType {
	case events.EventTypeSkip:
		return randomInt(rng, 5_000, 30_000)
	case events.EventTypeLike, events.EventTypeSave:
		return randomInt(rng, 30_000, 180_000)
	case events.EventTypeReplay:
		return randomInt(rng, 120_000, 300_000)
	default:
		return randomInt(rng, 60_000, 240_000)
	}
}

func randomInt(rng *rand.Rand, min, max int) int {
	if max <= min {
		return min
	}
	return min + rng.Intn(max-min+1)
}
