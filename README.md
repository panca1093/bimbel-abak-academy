# Abak Academy — App

Monorepo for the Abak Academy platform: exam engine + ecommerce + video courses.

## What's built

| Phase | Scope | Status |
|---|---|---|
| 1 — Foundation & Auth | OTP registration, JWT auth, RBAC middleware, refresh tokens | ✅ |
| 2 — Store & E-Commerce | Product catalog, cart, Midtrans Snap checkout, transactional-outbox worker, promos, refunds, revenue | ✅ |
| 3 — Student Frontend | Dashboard, catalog, cart, orders, course viewer (YouTube), profile + photo upload, i18n (ID/EN) | ✅ |
| Admin — MD3 shell | Blue/teal MD3 palette, dark mode, AdminPageHeader, role-based nav, super admin system pages | ✅ |
| Admin — Super Admin | User management, audit log, system config (AES-256-GCM encrypted Midtrans keys), store dashboard | ✅ |
| 4 — Exam Engine | Question bank, tryouts, schedules, auto-grading | 🔲 |

## Layout

```
app/
├── backend/            Go module — two binaries
│   ├── cmd/api/            REST API server (:8080)
│   ├── cmd/worker/         transactional-outbox worker
│   ├── internal/
│   │   ├── adapter/        external integrations (Midtrans, MinIO)
│   │   ├── handler/        HTTP layer (echo)
│   │   ├── infra/          postgres + redis wiring
│   │   ├── model/          shared types
│   │   ├── repository/     pgxpool data access
│   │   ├── server/         echo instance, middleware, /api/v1 routes
│   │   ├── service/        business logic
│   │   └── worker/         outbox poll loop
│   └── db/
│       └── migrations/     0001–0013 golang-migrate .sql files
├── web/                Next.js 15 (App Router, TS, Tailwind v4)
│   ├── app/(auth)/         login, register, OTP
│   ├── app/(student)/      dashboard, catalog, cart, orders, courses, profile
│   └── app/(admin)/        admin shell + all domain pages
├── deploy/             docker-compose.yml (full stack) + Dockerfiles
└── Makefile
```

Layering: `handler → service → repository / adapter`.

## Stack

| Layer | Choice |
|---|---|
| Router | echo v4 |
| DB | PostgreSQL via pgx/v5 (raw, no ORM) |
| Migrations | golang-migrate (13 migrations) |
| Cache / idempotency | Redis via go-redis/v9 |
| Logging | stdlib slog (JSON) |
| Payment | Midtrans Snap |
| Object storage | MinIO (S3-compatible) |
| Frontend | Next.js 15 + React 19 + Tailwind v4 |
| State | Zustand (auth, cart, UI/theme/lang) |
| Data fetching | TanStack Query v5 |
| i18n | Custom DICT hook (ID/EN, no external lib) |

## Quickstart (Docker)

```bash
cd deploy
docker compose up -d          # postgres + redis + minio + api + worker + web
```

- API: `http://localhost:8080/api/v1/health`
- App: `http://localhost:3000`
- Admin: `http://localhost:3000/admin` (login with super_admin account)
- MinIO console: `http://localhost:9001`

## Local development (without Docker)

**Prerequisites:** Go 1.23+ · Node 20+ · `golang-migrate` CLI · Docker (for infra only)

```bash
# 1. Start infra only
cd deploy && docker compose up -d postgres redis minio

# 2. Backend
cd backend
export GOROOT=/opt/homebrew/Cellar/go/1.26.3/libexec   # if on this machine
make migrate-up
make api      # :8080
make worker   # separate terminal

# 3. Frontend
cd web && npm install && npm run dev   # :3000
```

## Tests

```bash
# Backend (includes integration tests — requires postgres + redis running)
cd backend && go test ./... -race

# Frontend
cd web && npx vitest run
```
