# Akademi Bimbel — App

Monorepo for the Abak Academy platform (exam engine + ecommerce + video courses).
Source of truth for requirements lives one level up in `../requirements/`
(`product-requirements.md`, `technical-requirements.md`, `schema.dbml`,
`c4-diagrams.puml`).

## Layout

```
app/
├── backend/        Go module — two binaries from one codebase
│   ├── cmd/api/        REST API server (:8080)
│   ├── cmd/worker/     transactional-outbox worker
│   ├── internal/
│   │   ├── config/         env -> typed Config
│   │   ├── server/         echo instance, middleware, /api/v1 routes
│   │   ├── handler/        HTTP layer
│   │   ├── service/        business logic
│   │   ├── repository/     pgxpool data access
│   │   ├── platform/       postgres, redis, pluggable payment/logistics/notif/storage
│   │   └── worker/         outbox poll loop
│   └── db/
│       └── migrations/     golang-migrate .sql files
├── web/            Next.js (App Router, TS, Tailwind v4) — student UI + /admin
├── deploy/         docker-compose (postgres/redis/minio) + nginx prod reference
└── Makefile
```

Layering: `handler -> service -> repository / platform`. `GET /api/v1/health`
is wired through every layer (pings Postgres + Redis) to prove the skeleton.

## Stack

| Layer | Choice |
|---|---|
| Router | echo v4 |
| DB | PostgreSQL via pgx/v5 (raw, no ORM/codegen) |
| Migrations | golang-migrate |
| Cache | Redis via go-redis/v9 |
| Logging | stdlib slog (JSON) |
| Frontend | Next.js 15 + React 19 + Tailwind v4 |
| Object storage | S3-compatible (MinIO locally) |

## Prerequisites

- Go 1.23+ (this machine: `GOROOT=/opt/homebrew/Cellar/go/1.26.3/libexec`, baked into the Makefile)
- Docker (for the infra compose)
- Node 20+ and npm (for `web/`)
- `golang-migrate` CLI — `brew install golang-migrate`

## Quickstart

```bash
make up           # start postgres + redis + minio
make tidy         # resolve Go deps (first run)
make migrate-up   # create the outbox table
make api          # :8080  -> GET /api/v1/health
make worker       # outbox poll loop (separate terminal)

cd web && npm install && npm run dev   # :3000 (or: make web after install)
```

Verify:
- `curl localhost:8080/api/v1/health` → `{"status":"ok","postgres":"ok","redis":"ok"}`
- `http://localhost:3000` renders the student shell and shows live API health
- `http://localhost:3000/admin` renders the admin shell

## Scope

Runnable skeleton only — **zero business logic**. The six domain handlers from
the TRD component diagram (auth, exam, course, store, admin, payment webhook)
mount onto `/api/v1` as feature work begins. Design reference for screens:
`../abak-mockup-app` and `../design-app-abak/source`.
