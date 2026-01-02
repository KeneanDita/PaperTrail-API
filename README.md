# PaperTrail API

A lightweight Go starter for the PaperTrail API with modular folders, Postgres migrations, JWT auth, and Supabase storage hooks.

## Quick start

1. Copy `.env.example` to `.env` and fill values.
2. Run Postgres locally and create the database.
3. Run migrations automatically via `go run ./cmd/server` (they execute at boot).

.env.example

```.
APP_PORT=8080
ENV=development
DATABASE_URL=
JWT_SECRET=super-secret-jwt-key
SUPABASE_URL=
SUPABASE_KEY=
SUPABASE_BUCKET=

```

## Project layout

- `cmd/server`: entrypoint wiring config, DB, and HTTP server.
- `internal/config`: env loading and validation.
- `internal/database`: Postgres connection and naive migration runner.
- `internal/middleware`: logger, JWT auth, and role guard helpers.
- `internal/modules/*`: feature modules (users, papers, reviews, comments) with handlers/services/repos.
- `internal/storage`: Supabase storage client for PDF uploads.
- `internal/routes`: router composition and wiring.
- `migrations`: SQL schema files applied on startup.

## Running

```bash
go run ./cmd/server
```

By default the API listens on `:8080` and exposes `GET /health` plus authenticated `/api/*` routes.

## Notes

- JWT verification expects HMAC with `JWT_SECRET`; adapt to Supabase JWT rules as needed.
- Supabase uploads are a minimal example; swap with the official client or signed URLs for production.
- Migration runner is deliberately simple and non-idempotent beyond file order—use a real tool like `golang-migrate` in production.
