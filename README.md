Chirpy is a small Go HTTP server that backs a micro‑blogging API. It exposes endpoints to register users, authenticate with JWTs and refresh tokens, create and manage chirps, and handle a webhook that upgrades accounts. Static assets are also served under `/app`.

## Prerequisites
- Go 1.21+ installed locally
- PostgreSQL running and reachable; `psql` available for applying migrations

## Setup
1) Clone the repo and install deps:
```
go mod download
```
2) Create a PostgreSQL database (example name `chirpy`) and apply the schema in order:
```
psql -d chirpy -f sql/schema/001_users.sql
psql -d chirpy -f sql/schema/002_chirps.sql
psql -d chirpy -f sql/schema/003_alter_users.sql
psql -d chirpy -f sql/schema/004_refresh_tokens.sql
psql -d chirpy -f sql/schema/005_alter_users.sql
```
3) Provide environment variables (a `.env` file works locally):
```
DB_URL=postgres://user:pass@localhost:5432/chirpy?sslmode=disable
PRIVATE_KEY=replace-with-jwt-secret
POLKA_KEY=replace-with-webhook-key
# optional
PLATFORM=dev   # enables POST /admin/reset when set to dev
```
4) Start the server:
```
go run ./...
```
The API listens on `:8080`.

## API Overview
- `GET /api/healthz` — readiness probe.
- `POST /api/users` — sign up with `email`, `password`.
- `POST /api/login` — authenticate and receive JWT plus refresh token (`expires_in_seconds` optional, defaults to 60s).
- `POST /api/refresh` — exchange a refresh token (Authorization: `Bearer <refresh_token>`) for a new JWT.
- `POST /api/revoke` — revoke the presented refresh token.
- `PUT /api/users` — update `email` and `password` for the authenticated user (Authorization: `Bearer <jwt>`).
- `GET /api/chirps` — list chirps; supports `author_id=<uuid>` filter and `sort=asc|desc` (default desc).
- `GET /api/chirps/{chirp_id}` — fetch a single chirp.
- `POST /api/chirps` — create a chirp (Authorization: `Bearer <jwt>`); body limited to 140 chars.
- `DELETE /api/chirps/{chirpID}` — delete a chirp you own (Authorization: `Bearer <jwt>`).
- `GET /admin/metrics` — simple page showing file‑server hit count.
- `POST /admin/reset` — clears users table and resets metrics (only when `PLATFORM=dev`).
- `POST /api/polka/webhooks` — webhook secured via `Authorization: ApiKey <POLKA_KEY>`; when `event` is `user.upgraded`, marks the user as `is_chirpy_red=true`.
- Static assets served at `/app/` with `/app/assets` for files like `assets/logo.png`.

## Project Layout
- `main.go` — HTTP server setup and routing.
- `middleware.go`, `handlers.go` — request handlers and middleware.
- `internal/auth` — password hashing, JWT helpers, refresh token generator, header parsing.
- `internal/database` — sqlc‑generated data access layer built from `sql/queries`.
- `sql/schema` — migration files applied with `psql` (or your migration tool of choice).
- `assets/`, `index.html` — static frontend served from `/app`.

## Testing
```
go test ./...
```
Current tests cover utility functions; API endpoints are easiest to exercise with curl, HTTPie, or a client like Insomnia.
