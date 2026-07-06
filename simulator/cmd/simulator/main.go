package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"echorec/events"
	"echorec/simulator/internal/catalog"
	"echorec/simulator/internal/config"
	"echorec/simulator/internal/generator"

	"github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := catalog.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	tracks, err := catalog.LoadTracks(ctx, pool)
	if err != nil {
		log.Fatalf("load track catalog failed: %v", err)
	}
	log.Printf("loaded %d tracks from PostgreSQL", len(tracks))

	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	gen := generator.New(tracks, cfg.UserCount, 0)
	log.Printf(
		"starting simulator topic=%s brokers=%v interval=%s users=%d",
		cfg.KafkaTopic,
		cfg.KafkaBrokers,
		cfg.Interval,
		cfg.UserCount,
	)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("simulator stopped")
			return
		case <-ticker.C:
			event := gen.Next()
			if err := publishEvent(ctx, writer, event); err != nil {
				log.Printf("publish failed: %v", err)
				continue
			}
			log.Printf(
				"published event eventId=%s userId=%s trackId=%s eventType=%s genre=%s artistId=%s durationMs=%d",
				event.EventID,
				event.UserID,
				event.TrackID,
				event.EventType,
				event.Genre,
				event.ArtistID,
				event.DurationMs,
			)
		}
	}
}

func publishEvent(ctx context.Context, writer *kafka.Writer, event events.ListeningEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.UserID),
		Value: payload,
	})
}
