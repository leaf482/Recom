# EchoRec

Real-time music recommendation platform (work in progress).

## Phase 1 — Local Development

Start all services (PostgreSQL, Redis, Redpanda, API, simulator, consumer):

```bash
docker compose up --build
```

Verify the API health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"OK"}
```

List tracks from the catalog:

```bash
curl http://localhost:8080/tracks
```

Optional limit:

```bash
curl "http://localhost:8080/tracks?limit=10"
```

If PostgreSQL was already initialized before Phase 2, recreate the database volume so migrations and seed run:

```bash
docker compose down -v
docker compose up --build
```

Watch simulator logs:

```bash
docker compose logs -f simulator
```

Inspect events on the `listening-events` topic:

```bash
docker compose exec redpanda rpk topic consume listening-events -n 5
```

Watch consumer logs:

```bash
docker compose logs -f consumer
```

Inspect Redis user features (replace `user_1` as needed):

```bash
docker compose exec redis redis-cli HGETALL user:user_1:genre_score
docker compose exec redis redis-cli HGETALL user:user_1:artist_score
docker compose exec redis redis-cli LRANGE user:user_1:recent_tracks 0 -1
docker compose exec redis redis-cli HGETALL user:user_1:event_counts
```

User features and recommendations:

```bash
curl http://localhost:8080/users/user_1/features
curl http://localhost:8080/users/user_1/recommendations
curl "http://localhost:8080/users/user_99/recommendations?limit=5"
```

After the simulator and consumer run for a minute, compare recommendations for different users:

```bash
curl http://localhost:8080/users/user_1/recommendations
curl http://localhost:8080/users/user_2/recommendations
```

Example event payload:

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

## Services

| Service    | Port  |
|------------|-------|
| API        | 8080  |
| Simulator  | —     |
| Consumer   | —     |
| PostgreSQL | 5432  |
| Redis      | 6379  |
| Redpanda   | 19092 |

See [cursor.md](./cursor.md) for the full project plan and development phases.
