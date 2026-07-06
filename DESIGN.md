# EchoRec — Design Document

This document explains the technical decisions behind EchoRec: why each component exists, how data flows through the system, and what would change at production scale.

## Overview

EchoRec implements a classic recommendation feedback loop:

```text
recommend → user action → event ingestion → feature update → better recommendation
```

The MVP uses a simulator instead of real clients, but the data path mirrors production recommendation systems: events are decoupled from serving, features are updated asynchronously, and the API reads precomputed signals at request time.

## Why Streaming Is Used

Listening events (`play`, `skip`, `like`, `save`, `replay`) are produced continuously and consumed by a separate feature-updater service.

**Redpanda** (Kafka-compatible) provides:

- **Decoupling** — the simulator and API do not call the consumer directly
- **Buffering** — bursts of events are absorbed without blocking serving
- **Replay** — events can be reprocessed if feature logic changes
- **Multiple consumers later** — metrics, audit, or ML pipelines can subscribe to the same topic

Topic: `listening-events`

In production, producers would be mobile/web clients or backend services recording real user behavior. The consumer group (`feature-consumer`) ensures each event is processed once per group.

## Why Redis Is the Feature Store

User preference signals need **fast reads** on every recommendation request. Redis fits because:

- Sub-millisecond reads for genre/artist hashes and recent-track lists
- Simple increment operations (`HINCRBY`) for event-driven score updates
- Low operational overhead for an MVP

**Redis keys per user:**

| Key | Type | Purpose |
|-----|------|---------|
| `user:{userId}:genre_score` | Hash | Genre → accumulated score |
| `user:{userId}:artist_score` | Hash | Artist → accumulated score |
| `user:{userId}:recent_tracks` | List | Last 20 track IDs (newest first) |
| `user:{userId}:event_counts` | Hash | Event type → count |

Redis is a **feature store**, not the system of record. If Redis data were lost, features could be rebuilt by replaying the event stream (not implemented in the MVP).

## Why PostgreSQL Holds the Track Catalog

Track metadata is relational, mostly static, and queried as a set of candidates:

- `id`, `title`, `artist_id`, `artist_name`, `genre`, `popularity`, `release_year`

PostgreSQL provides durable storage, seed data for demos, and straightforward SQL for listing candidates. The MVP seeds **100 tracks**, **10 genres**, and **20 artists**.

At MVP scale, the API loads all tracks per recommendation request. This is acceptable for 100 rows but would not scale unchanged to millions of tracks.

## How User Features Are Updated

The **feature consumer** reads `ListeningEvent` JSON from Kafka and applies explainable scoring rules:

| Event | Genre / artist delta |
|-------|----------------------|
| play | +1 |
| skip | -2 |
| like | +4 |
| save | +5 |
| replay | +6 |

For each event, the consumer atomically (Redis pipeline):

1. `HINCRBY` genre and artist scores
2. `HINCRBY` event type count
3. `LPUSH` track ID to recent list, `LTRIM` to 20 entries

The simulator loads real `trackId`, `genre`, and `artistId` from PostgreSQL so events align with the catalog.

## How Recommendation Scoring Works

The Go API combines Redis features with PostgreSQL candidates.

**Steps:**

1. Assign experiment strategy (stable per `userId`)
2. Load Redis features
3. Load all tracks from PostgreSQL
4. Exclude recently played tracks when enough candidates remain
5. Score each track, sort descending, return top N

**Per-track score** (non-cold-start):

```text
score =
  genreAffinity  × Wg
+ artistAffinity × Wa
+ popularity     × Wp
+ freshness      × Wf
```

Where:

- `genreAffinity` / `artistAffinity` = user's positive score for that genre/artist, divided by the user's max positive score in that dimension (negative scores contribute 0)
- `popularity` = `tracks.popularity / 100`
- `freshness` = normalized release year in range 2015–2026

**Strategy weights (Wg, Wa, Wp, Wf):**

| Strategy | Genre | Artist | Popularity | Freshness |
|----------|-------|--------|------------|-----------|
| `genre_affinity` | 0.55 | 0.20 | 0.15 | 0.10 |
| `artist_affinity` | 0.25 | 0.45 | 0.20 | 0.10 |
| `exploration` | 0.20 | 0.10 | 0.35 | 0.35 |

Each recommendation includes a `reason` derived from the strongest contributing signals.

## How Cold-Start Recommendations Work

A user is in **cold start** when both `genre_score` and `artist_score` hashes are empty (e.g. `user_99` before any events).

In that case:

- Genre and artist affinity are not used
- Score uses only popularity and freshness, renormalized to 0–1
- Reason: `"Popular recent track for new listeners"`

This ensures the API always returns recommendations, even with no listening history.

## How A/B Strategy Assignment Works

Assignment is **deterministic** per user:

```text
strategy = Strategies[FNV-1a(userId) % 3]
```

Strategies:

1. `genre_affinity`
2. `artist_affinity`
3. `exploration`

The same `userId` always maps to the same strategy. No per-request randomness.

Strategy affects scoring weights and slightly biases explanation text — not the assignment of users to Redis features (all users share the same feature update logic).

## How Metrics Are Tracked

On each `GET /users/{userId}/recommendations` call, the API records experiment metrics in Redis for the assigned strategy:

| Metric | How it is updated |
|--------|-------------------|
| `recommendation_requests` | `HINCRBY` +1 |
| `impressions` | `HINCRBY` by number of tracks returned |
| Latency samples | `LPUSH` request duration (ms), keep last 100 |

Keys:

```text
experiment:default:strategy:{strategy}:metrics
experiment:default:strategy:{strategy}:latencies
```

`GET /experiments/default/metrics` aggregates all three strategies, computing average and p95 latency from the stored samples.

**Not tracked in MVP:** skip rate, save rate, replay rate after recommendation (would require linking served recommendations to subsequent events).

## What Would Change at Production Scale

| Area | MVP | Production direction |
|------|-----|----------------------|
| Catalog query | Load all tracks | Indexed retrieval, pagination, pre-filtered candidates |
| Feature store | Redis hashes | Redis + periodic snapshot to warehouse; optional stream replay |
| Streaming | Single topic, one consumer | Partitioned topics, multiple consumer groups, dead-letter queues |
| Serving | Synchronous scoring in API | Precomputed candidate pools, caching, dedicated ranking service |
| Experiments | 3 strategies, Redis counters | Experiment platform, statistical testing, outcome-linked metrics |
| Observability | Logs + dashboard | Prometheus, Grafana, distributed tracing |
| Security | None | AuthN/AuthZ, rate limiting, tenant isolation |
| Deployment | Docker Compose | Kubernetes, autoscaling, managed Kafka/Redis/Postgres |

## Current Limitations and Tradeoffs

**Intentional simplifications:**

- **No ML** — explainable rules are easier to debug and discuss in interviews
- **Full catalog scan** — simple code path; acceptable only at small scale
- **In-process experiment metrics** — Redis is convenient but not a full experiment analytics platform
- **Simulator-only input** — validates the pipeline without external integrations
- **No event schema registry** — shared Go struct is enough for the monorepo MVP
- **Dashboard proxies API** — avoids CORS changes; Next.js rewrites `/backend/*` to the Go API

**Honest scope boundaries:**

- Not production-hardened (no auth, no HA, no backup strategy documented)
- Not benchmarked — latency numbers in metrics reflect your local Docker run, not published SLAs
- Not integrated with real music APIs

## Shared Event Schema

Events are defined in `events/listening_event.go` and shared by the simulator and consumer:

```json
{
  "eventId": "evt_123",
  "userId": "user_1",
  "trackId": "track_042",
  "eventType": "skip",
  "genre": "indie",
  "artistId": "artist_7",
  "durationMs": 12000,
  "timestamp": "2026-07-05T12:00:00Z"
}
```

Keeping one schema avoids drift between producer and consumer.

## Service Boundaries

```text
simulator/   Produces events (reads catalog for realistic track metadata)
consumer/    Consumes events → updates Redis
api/         Serves features, recommendations, experiment metrics
dashboard/   Visualizes API data for demos
events/      Shared event types
db/          SQL migrations and seed data
```

This separation makes it straightforward to explain each component in a system design interview.
