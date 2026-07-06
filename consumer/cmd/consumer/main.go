package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"echorec/consumer/internal/config"
	"echorec/consumer/internal/features"
	"echorec/events"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaGroupID,
	})
	defer reader.Close()

	updater := features.NewUpdater(redisClient)
	log.Printf(
		"starting consumer topic=%s group=%s brokers=%v redis=%s",
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		cfg.KafkaBrokers,
		cfg.RedisAddr,
	)

	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("consumer stopped")
				return
			}
			log.Printf("fetch message failed: %v", err)
			continue
		}

		var event events.ListeningEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("invalid event payload: %v", err)
			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Printf("commit invalid message failed: %v", err)
			}
			continue
		}

		log.Printf(
			"consumed event eventId=%s userId=%s trackId=%s eventType=%s genre=%s artistId=%s",
			event.EventID,
			event.UserID,
			event.TrackID,
			event.EventType,
			event.Genre,
			event.ArtistID,
		)

		result, err := updater.ApplyEvent(ctx, event)
		if err != nil {
			log.Printf("feature update failed userId=%s eventId=%s: %v", event.UserID, event.EventID, err)
			continue
		}

		log.Printf(
			"updated features userId=%s genre=%s genreScore=%d artistId=%s artistScore=%d eventType=%s eventCount=%d recentTracks=%d",
			event.UserID,
			event.Genre,
			result.GenreScore,
			event.ArtistID,
			result.ArtistScore,
			event.EventType,
			result.EventCount,
			result.RecentTracks,
		)

		if err := reader.CommitMessages(ctx, message); err != nil {
			log.Printf("commit message failed: %v", err)
		}
	}
}
