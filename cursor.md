# EchoRec — Real-time Music Recommendation Platform

## Project Goal

Build **EchoRec**, a production-style real-time music recommendation platform.

This is not a simple CRUD music app.
The goal is to build a backend-heavy system that demonstrates:

* real-time event ingestion
* user personalization
* recommendation serving
* feedback loop
* experimentation / A-B testing
* observability-ready architecture

The project should be useful for SWE job applications, especially backend, infrastructure, distributed systems, and recommendation-related roles.

## Core Idea

Users generate listening events such as play, skip, like, save, and replay.

Those events flow through the system, update each user's preference profile, and affect future recommendations.

The core loop is:

```text
recommend → user action → event ingestion → feature update → better recommendation
```

## Target Story for Resume

The final project should support this resume bullet:

```text
Built EchoRec, a real-time music recommendation platform that ingests listening events through Kafka-compatible streaming, updates user preference features in Redis, and serves personalized recommendations through a low-latency backend API.
```

Another possible bullet:

```text
Implemented a feedback loop where user actions such as skips, saves, and replays continuously update user preference features and influence future recommendations.
```

Another possible bullet:

```text
Designed an experimentation layer to compare recommendation strategies using skip rate, save rate, replay rate, and p95 recommendation latency.
```

## Tech Stack

Use this stack unless there is a strong reason not to:

```text
Backend API: Go
Event Streaming: Redpanda or Kafka
Cache / Feature Store: Redis
Database: PostgreSQL
Frontend Dashboard: Next.js + TypeScript
Local Dev: Docker Compose
Optional Later: Prometheus / Grafana
Optional Later: pgvector
```

Keep the implementation simple and focused.
Do not over-engineer.

## Important Engineering Principles

Follow these rules throughout the project:

1. Think before coding.
2. Keep the solution simple.
3. Modify only necessary parts.
4. Implement one phase at a time.
5. Define clear success criteria before each phase.
6. Verify correctness after each phase.
7. Prefer readable code over clever code.
8. Do not introduce unnecessary abstractions.
9. Do not add complex ML unless the MVP is already working.
10. Avoid making the project look like a generic CRUD app.

## MVP Scope

The MVP is complete when the following works:

1. A simulator generates listening events.
2. Events are sent to Redpanda or Kafka.
3. A consumer reads events and updates Redis user profiles.
4. PostgreSQL stores a small music catalog.
5. A Go recommendation API returns personalized recommendations for a user.
6. User behavior changes future recommendations.
7. A basic dashboard shows events, user features, recommendations, and experiment metrics.

## Non-Goals for MVP

Do not build these in the MVP:

* full authentication
* payment
* real Spotify API integration
* complex ML training pipeline
* Kubernetes
* microservice sprawl
* overly fancy UI
* production cloud deployment
* complicated admin panel

These can be added later only if the MVP is already complete.

## System Architecture

Expected architecture:

```text
Event Simulator
  → Redpanda / Kafka
  → Feature Consumer
  → Redis Feature Store
  → PostgreSQL Track Catalog
  → Go Recommendation API
  → Next.js Dashboard
```

## Main Components

### 1. Event Simulator

The simulator generates fake user listening events.

Example event:

```json
{
  "eventId": "evt_123",
  "userId": "user_1",
  "trackId": "track_42",
  "eventType": "skip",
  "genre": "indie",
  "artistId": "artist_7",
  "durationMs": 12000,
  "timestamp": "2026-07-05T12:00:00Z"
}
```

Supported event types:

```text
play
skip
like
save
replay
```

The simulator should be deterministic enough for testing, but random enough to make the dashboard interesting.

### 2. Streaming Layer

Use Redpanda or Kafka locally through Docker Compose.

Topic name:

```text
listening-events
```

The simulator publishes events to this topic.

The feature consumer consumes from this topic.

### 3. Feature Consumer

The consumer reads listening events and updates user preference features in Redis.

Example Redis features:

```text
user:{userId}:genre_score
user:{userId}:artist_score
user:{userId}:recent_tracks
user:{userId}:skip_count
user:{userId}:like_count
user:{userId}:save_count
user:{userId}:replay_count
```

Suggested scoring rules:

```text
play   → small positive signal
skip   → negative signal
like   → strong positive signal
save   → strong positive signal
replay → very strong positive signal
```

Keep the scoring logic simple and explainable.

### 4. PostgreSQL Music Catalog

PostgreSQL stores track metadata.

Minimum table:

```text
tracks
- id
- title
- artist_id
- artist_name
- genre
- popularity
- release_year
- created_at
```

Seed the database with fake tracks.

The catalog should include enough data to make recommendations meaningful.

Minimum target:

```text
100 tracks
10 genres
20 artists
```

### 5. Recommendation API

Build a Go API that exposes recommendation endpoints.

Required endpoints:

```text
GET /health
GET /users/{userId}/features
GET /users/{userId}/recommendations
POST /events
GET /experiments/{experimentId}/metrics
```

`POST /events` should allow manual event ingestion for testing.
It can either publish to Kafka/Redpanda or directly reuse shared event logic if streaming is not available in early development.

Recommendation response example:

```json
{
  "userId": "user_1",
  "strategy": "genre_affinity",
  "recommendations": [
    {
      "trackId": "track_42",
      "title": "Night Drive",
      "artistName": "The Signals",
      "genre": "indie",
      "score": 0.91,
      "reason": "Strong match with user's indie preference and recent replay behavior"
    }
  ]
}
```

### 6. Recommendation Strategy

Start with explainable rule-based recommendation.

Do not build a complex ML model for MVP.

Candidate retrieval:

```text
1. Load user genre and artist preferences from Redis.
2. Fetch candidate tracks from PostgreSQL.
3. Exclude recently played tracks.
4. Score tracks using user preferences and track popularity.
5. Return top N recommendations.
```

Suggested scoring formula:

```text
score =
  genreAffinity * 0.45
+ artistAffinity * 0.25
+ popularity * 0.15
+ freshness * 0.10
- skipPenalty * 0.05
```

The exact formula can be adjusted, but it must stay understandable.

Every recommendation should include a short `reason`.

### 7. Experimentation Layer

Implement a simple A/B testing layer.

Each user should be assigned to one recommendation strategy.

Example groups:

```text
A: genre_affinity
B: artist_affinity
C: exploration
```

Assignment should be stable by user ID.

For example:

```text
hash(userId) % 3
```

Track basic metrics per strategy:

```text
impressions
plays
skips
likes
saves
replays
skip_rate
save_rate
replay_rate
average_latency_ms
p95_latency_ms
```

The experimentation layer does not need to be perfect.
It should be good enough to explain how different recommendation strategies can be compared.

### 8. Dashboard

Build a simple Next.js dashboard.

Dashboard should show:

1. live or recent listening events
2. selected user's feature profile
3. recommendation results
4. experiment group assignment
5. experiment metrics
6. basic system status

Keep the UI clean and practical.

The dashboard should help explain the system during interviews.

## Suggested Repository Structure

Use a simple monorepo structure:

```text
echorec/
  cursor.md
  README.md
  DESIGN.md
  docker-compose.yml

  api/
    cmd/
    internal/
      config/
      db/
      redis/
      events/
      recommendation/
      experiments/
      http/

  consumer/
    cmd/
    internal/
      events/
      features/
      redis/

  simulator/
    cmd/
    internal/

  dashboard/
    app/
    components/
    lib/

  db/
    migrations/
    seed/

  scripts/
```

Do not create unnecessary folders unless needed.

## Development Phases

### Phase 1 — Project Skeleton

Goal:

Set up the repository structure, Docker Compose, and basic services.

Implement:

* Go API skeleton
* Docker Compose with PostgreSQL, Redis, and Redpanda/Kafka
* health endpoint
* README start command

Success criteria:

```text
docker compose up works
GET /health returns OK
PostgreSQL starts
Redis starts
Redpanda/Kafka starts
```

Do not move to Phase 2 until this works.

---

### Phase 2 — Music Catalog

Goal:

Add PostgreSQL track catalog.

Implement:

* tracks table
* migration
* seed script
* DB connection from Go API
* endpoint to list tracks if helpful

Success criteria:

```text
tracks table exists
database is seeded with at least 100 tracks
API can query tracks from PostgreSQL
```

---

### Phase 3 — Event Model and Simulator

Goal:

Generate realistic listening events.

Implement:

* event struct/schema
* simulator service
* random event generation
* publish events to streaming topic

Success criteria:

```text
simulator generates events
events are visible in Redpanda/Kafka topic
event payloads are valid
```

---

### Phase 4 — Feature Consumer

Goal:

Consume listening events and update Redis user features.

Implement:

* Kafka/Redpanda consumer
* event parsing
* Redis feature update logic
* simple scoring rules

Success criteria:

```text
consumer reads listening-events topic
Redis user features change after events
skip/like/save/replay affect scores differently
```

---

### Phase 5 — Recommendation API

Goal:

Serve personalized recommendations.

Implement:

* GET /users/{userId}/features
* GET /users/{userId}/recommendations
* candidate retrieval from PostgreSQL
* feature lookup from Redis
* scoring function
* explanation reason per recommendation

Success criteria:

```text
recommendations differ by user
new events change future recommendations
recently played tracks can be excluded
response includes score and reason
```

---

### Phase 6 — Experimentation

Goal:

Add stable A/B strategy assignment and metrics.

Implement:

* stable user-to-strategy assignment
* multiple simple strategies
* basic metrics collection
* GET /experiments/{experimentId}/metrics

Success criteria:

```text
same user always gets same strategy
different users can receive different strategies
metrics are updated from events or recommendation responses
```

---

### Phase 7 — Dashboard

Goal:

Build a simple UI to explain the system.

Implement:

* user selector
* feature profile panel
* recommendation panel
* recent events panel
* experiment metrics panel
* system status panel

Success criteria:

```text
dashboard can call API
dashboard shows user features
dashboard shows recommendations
dashboard shows experiment metrics
```

---

### Phase 8 — Polish for Job Applications

Goal:

Make the project portfolio-ready.

Implement:

* README with architecture diagram
* DESIGN.md with tradeoffs
* setup instructions
* screenshots
* benchmark notes if available
* resume bullets
* future improvements section

Success criteria:

```text
a recruiter or engineer can understand the project in under 2 minutes
a technical interviewer can ask system design questions from the README
the project has clear backend/distributed-systems value
```

## Testing Requirements

Add tests where they provide value.

Prioritize tests for:

* event parsing
* feature scoring
* recommendation scoring
* experiment assignment
* API handlers where simple

Do not block MVP progress by trying to test everything.

## Code Style

Use clear, readable code.

Prefer names like:

```text
userFeatures
trackCatalog
recommendationScore
experimentGroup
recentTracks
```

Avoid overly abstract names.

Avoid unnecessary design patterns.

Do not write code that looks artificially complex.

## README Requirements

The final README should include:

1. What the project does
2. Why it exists
3. Architecture diagram
4. Tech stack
5. How to run locally
6. Example API requests
7. Recommendation strategy explanation
8. Experimentation explanation
9. What tradeoffs were made
10. Future improvements
11. Resume bullets

## DESIGN.md Requirements

The design document should explain:

1. Why streaming is used
2. Why Redis is used as a feature store
3. Why PostgreSQL is used for the catalog
4. How recommendation scoring works
5. How feedback loop works
6. How A/B assignment works
7. What would change at production scale
8. Current limitations

## Important Constraints

Do not implement the entire project in one giant change.

Work phase by phase.

Before coding each phase:

1. briefly summarize what will be changed
2. define success criteria
3. implement only that phase
4. run checks
5. explain what was completed and what remains

## First Task

Start with **Phase 1 — Project Skeleton**.

Before writing code, inspect the current repository.

Then propose the minimal file structure and implementation plan for Phase 1.

After that, implement Phase 1 only.

Do not implement later phases yet.