# EchoRec

Real-time music recommendation platform (work in progress).

## Phase 1 — Local Development

Start all services (PostgreSQL, Redis, Redpanda, API, simulator):

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
| PostgreSQL | 5432  |
| Redis      | 6379  |
| Redpanda   | 19092 |

See [cursor.md](./cursor.md) for the full project plan and development phases.
