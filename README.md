# PaperTrail API

PaperTrail is a small Go API (Postgres + migrations + JWT auth) and a Next.js UI for managing papers, reviews, and comments.

![Screenshot of running session](/1.png)

## Quick start

### 1) Configure environment

Create a `.env` in the repo root (see the example below).

```dotenv
APP_PORT=8080
ENV=development
DATABASE_URL=postgres://postgres:password@localhost:5432/papertrail?sslmode=disable
JWT_SECRET=super-secret-jwt-key
SUPABASE_URL=
SUPABASE_KEY=
SUPABASE_BUCKET=papers
```

### 2) Start the API

```bash
go run ./cmd/server
```

Default: `http://localhost:8080`

- `GET /health` (no auth)
- `POST /api/papers` (auth)
- `POST /api/papers/{id}/reviews` (auth)
- `POST /api/papers/{id}/comments` (auth)
- `GET /api/users` + `POST /api/users` are currently public for easy bootstrapping

### 3) Start the UI

```bash
npm --prefix ui install
npm --prefix ui run dev
```

Default: `http://localhost:3000`

The UI talks to the API at:

- `NEXT_PUBLIC_API_BASE_URL` (optional), otherwise
- a user-configurable localStorage value, otherwise
- defaults to `http://localhost:8080`

## Public pages / endpoints

To support a lightweight public view, the API exposes unauthenticated read-only paper endpoints:

- `GET /api/papers`
- `GET /api/papers/{id}`

And the UI includes:

- `GET /public/papers` (lists papers without a JWT)
- `GET /public/papers/{id}` (paper details without a JWT)

## Auth notes

- Protected endpoints require `Authorization: Bearer <JWT>`.
- JWTs are validated as HS256 using `JWT_SECRET`.
- The middleware reads `sub` and `role` claims (role gates are used by some endpoints).

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

API:

```bash
go run ./cmd/server
```

UI:

```bash
npm --prefix ui run dev
```

### ~ DB Schema

![Screenshot of DB Schema](/Screenshot.png)

## Notes

- JWT verification expects HMAC with `JWT_SECRET`; adapt to Supabase JWT rules as needed.
- Supabase uploads are a minimal example; swap with the official client or signed URLs for production.
- Migration runner is deliberately simple and non-idempotent beyond file order—use a real tool like `golang-migrate` in production.
