# Location Processing System

A system for tracking and querying user locations, built with Go microservices.

## Overview

The system consists of two microservices:

| Service | Port | Responsibility |
|---|---|---|
| **loc-management** | 8080 | Update user location, search users by radius |
| **loc-history** | 8081 (HTTP) / 50051 (gRPC) | Store location history, calculate travel distance |

When a user's location is updated, `loc-management` persists it to its own database and forwards the event to `loc-history` via gRPC/Protobuf. `loc-history` stores all historical events, enabling distance queries over any time range.

## API Reference

### Service 1 – loc-management

#### Update user location
```
PUT /api/v1/users/:username/location
Content-Type: application/json

{
  "latitude": 35.12314,
  "longitude": 27.64532
}
```
Response `200 OK`:
```json
{"message": "location updated"}
```

#### Search users by location and radius
```
GET /api/v1/users?lat=35.12314&lon=27.64532&radius=100&page=1&size=10
```
Response `200 OK`:
```json
{
  "users": [
    {"username": "alice", "latitude": 35.12, "longitude": 27.64}
  ],
  "total": 1,
  "page": 1,
  "size": 10
}
```

- `radius` is in **kilometres**
- `page` and `size` are optional (defaults: 1 and 10)

### Service 2 – loc-history

#### Get distance traveled
```
GET /api/v1/users/:username/distance?from=2021-09-01T00:00:00Z&to=2021-09-02T00:00:00Z
```
Response `200 OK`:
```json
{
  "username": "alice",
  "distance_km": 445.87,
  "from": "2021-09-01T00:00:00Z",
  "to": "2021-09-02T00:00:00Z"
}
```

- Dates use **ISO 8601 / RFC3339** format. The `+` sign must be URL-encoded as `%2B` in query strings.
- `to` is optional; when omitted it defaults to `from + 24h`.

### Validation rules

| Field | Rule |
|---|---|
| `username` | 4–16 characters, `[a-zA-Z0-9]` only |
| `latitude` | −90 … 90 |
| `longitude` | −180 … 180 |
| `radius` | positive number, km |
| dates | ISO 8601 (RFC3339) |

## Local Setup

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- `protoc` + plugins (only needed to regenerate proto files)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (only for `make migrate-up` / `make migrate-down` when not using Compose)

### Run with Docker Compose (recommended)

```bash
docker-compose up --build
```

This starts both PostgreSQL databases, applies schema migrations, then both
services. New files such as `000002_*.up.sql` are applied on the next `up`
without wiping data. Use `docker-compose down -v` only when you want a full
reset (deletes all rows).

The first time after switching to golang-migrate, run `docker-compose down -v`
once so volumes created by the old init scripts are replaced.

### Run locally (without Docker)

1. Start two PostgreSQL instances (or adjust the URLs):

```bash
# DB for loc-management
docker run -d --name pgmgmt -e POSTGRES_DB=loc_management \
  -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password \
  -p 5432:5432 postgres:16-alpine

# DB for loc-history
docker run -d --name pghist -e POSTGRES_DB=loc_history \
  -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password \
  -p 5433:5432 postgres:16-alpine
```

2. Apply schema migrations (requires [golang-migrate](https://github.com/golang-migrate/migrate)):

```bash
# brew install golang-migrate
make migrate-up
```

Roll back the latest version on both databases with `make migrate-down`.

3. Start `loc-history` first (it provides the gRPC endpoint):

```bash
make run-history
```

4. Start `loc-management`:

```bash
make run-management
```

### Adding a later schema change (`ALTER`)

Do **not** edit `000001_init.up.sql`. Add a new numbered pair, for example:

`migrations/loc-management/000002_add_accuracy.up.sql`

```sql
ALTER TABLE users_location
    ADD COLUMN accuracy DOUBLE PRECISION;
```

`migrations/loc-management/000002_add_accuracy.down.sql`

```sql
ALTER TABLE users_location
    DROP COLUMN IF EXISTS accuracy;
```

Then apply (existing rows are kept):

```bash
docker-compose up --build
```

or, with local Postgres:

```bash
make migrate-up
```

Do not run `docker-compose down -v` unless you want to wipe the databases.

### Environment variables

**loc-management:**

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:password@localhost:5432/loc_management?sslmode=disable` | PostgreSQL connection string |
| `HISTORY_SERVICE_ADDR` | `localhost:50051` | gRPC address of loc-history |
| `PORT` | `8080` | HTTP listen port |

**loc-history:**

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:password@localhost:5433/loc_history?sslmode=disable` | PostgreSQL connection string |
| `GRPC_PORT` | `50051` | gRPC listen port |
| `PORT` | `8081` | HTTP listen port |

## Testing

```bash
# Unit and functional tests
make test

# Integration tests (requires running services)
LOC_MGMT_URL=http://localhost:8080 make test-integration
```

Integration tests are gated behind the `integration` build tag so they are
never run accidentally in CI without the required infrastructure.

## Protobuf / gRPC

The service contract lives in `pkg/proto/location.proto`. To regenerate the
Go bindings after modifying it:

```bash
# Install tools (once)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Regenerate
make proto
```

## Observability

Both services expose Prometheus metrics at `/metrics` and a liveness endpoint at `/health`.
