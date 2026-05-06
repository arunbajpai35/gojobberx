# GoJobberX

A Postgres-backed job queue in Go — priority dispatch, exponential-backoff retries, dead-letter queue, crash recovery, graceful shutdown, Prometheus metrics, React dashboard. Built end-to-end as a learning project to internalize the design trade-offs that production queues like [River](https://riverqueue.com) and [Asynq](https://github.com/hibiken/asynq) make.

> Not intended for production. See [What this is NOT](#what-this-is-not).

## Architecture

```
                    ┌────────────────────────────────────────────┐
                    │                  backend                   │
   POST /job        │                                            │
  ─────────────────▶│  EnqueueJob ──▶ SaveJob ──▶ queueByPriority│
                    │                              │             │
                    │                              ▼             │
                    │   ┌──────────┐   ┌──────────┐  ┌─────────┐ │
                    │   │ highQueue│   │medQueue  │  │lowQueue │ │
                    │   └────┬─────┘   └────┬─────┘  └────┬────┘ │
                    │        └────────────┐ │ ┌───────────┘      │
                    │            priority │ │ │                  │
                    │             dispatcher                     │
                    │                  │                         │
                    │                  ▼                         │
                    │             jobQueue                       │
                    │                  │                         │
                    │      ┌───────────┼───────────┐             │
                    │      ▼           ▼           ▼             │
                    │   worker 0    worker 1    worker 2         │
                    │      │           │           │             │
                    │      └───┬───────┴───────────┘             │
                    │          │                                 │
                    │   success│  fail+retries<max  fail+retries=max
                    │          ▼          │              │       │
                    │     completed   exp backoff       DLQ      │
                    │                                            │
                    └────────────────────┬───────────────────────┘
                                         ▼
                                  PostgreSQL
                              (jobs, dead_jobs)
```

## Features

- **Priority queues** — high → medium → low, drained in order each tick
- **Worker pool** — N goroutines consume from a shared `jobQueue`
- **Retries with exponential backoff + jitter** — `2^attempt seconds + 0–500ms`
- **Dead-letter queue** — jobs that exceed `max_retries` move to a `dead_jobs` table
- **Crash recovery** — on boot, `processing` rows are reset to `queued` and re-enqueued
- **Graceful shutdown** — SIGINT/SIGTERM trigger drain; in-flight jobs finish before exit
- **Input validation** — bad payloads rejected at the HTTP boundary
- **Structured logging** — `slog` with per-job context (`job_id`, `worker`, `type`)
- **Prometheus metrics** — `jobs_total{status}` exposed at `/metrics`
- **React dashboard** — enqueue form, status counts, jobs/DLQ tabs, color-coded badges

## Quickstart

```bash
docker-compose down -v && docker-compose up -d --build
```

| Service     | URL                          |
| ----------- | ---------------------------- |
| Dashboard   | http://localhost             |
| API         | http://localhost:8080        |
| Postgres    | localhost:5432 (`postgres`/`password`) |
| Metrics     | http://localhost:8080/metrics |

Stop with `docker-compose down`. Wipe state with `docker-compose down -v`.

## API

| Method | Path           | Description                          |
| ------ | -------------- | ------------------------------------ |
| POST   | `/job`         | Enqueue a job                        |
| GET    | `/jobs`        | List all jobs (newest first)         |
| GET    | `/job/:id`     | Get one job by id                    |
| GET    | `/dead-jobs`   | List dead-lettered jobs              |
| GET    | `/metrics`     | Prometheus exposition                |
| GET    | `/health`      | Health check                         |

### Enqueue

```bash
curl -X POST http://localhost:8080/job \
  -H 'Content-Type: application/json' \
  -d '{"payload":"hello@example.com","type":"send_email","duration":2,"priority":"high"}'
```

| Field      | Required | Constraint                                   |
| ---------- | -------- | -------------------------------------------- |
| `payload`  | yes      | non-empty string                             |
| `type`     | yes      | `send_email` \| `generate_invoice`           |
| `duration` | no       | int, 0–600 (simulated work seconds)          |
| `priority` | no       | `high` \| `medium` \| `low` (default medium) |

To exercise the DLQ flow, send a payload starting with `fail` (e.g. `fail-demo`). It always fails, exhausts retries, and lands in `/dead-jobs` after ~15s.

## Design choices and trade-offs

A few decisions worth calling out — these are where the project differs from a serious queue and where the interesting learning lives:

- **Postgres over Redis.** Strong durability, transactional inserts, easy to reason about. Cost: throughput is much lower than a Redis-backed queue. River makes the same call; Asynq does not.
- **Channels for in-memory dispatch.** Simple, idiomatic Go. Cost: state is lost on crash for jobs sitting in the channel, which is why the boot-time recovery loop exists.
- **In-memory retry timer (`time.AfterFunc`).** Easy to write. Cost: a crash between scheduling and firing loses the retry; recovery only kicks in next boot when the row is re-read. A production queue persists `next_run_at` and polls.
- **Strict priority drain (high → medium → low).** Easy to implement and demo. Cost: sustained high-priority load starves medium/low. Real queues use weighted draining or aged-priority promotion.
- **Single-node only.** No leader election, no distributed locking. Cost: cannot horizontally scale workers. River uses Postgres advisory locks for this; we don't.

## What this is NOT

- Not a replacement for River, Asynq, Sidekiq, Celery, or SQS. Use those.
- Not multi-node. The worker pool is in-process.
- Not a scheduler. There's no `run_at` / cron / delay-by support.
- Not authenticated. Anyone with network access to the API can enqueue.

## Stack

- **Backend:** Go 1.23, Gin, pgx/v5, Prometheus client, `log/slog`
- **Frontend:** React 19, Vite 6, TailwindCSS 3
- **Database:** PostgreSQL 14
- **Infra:** Docker Compose, Nginx (frontend reverse proxy), multi-stage build

## Local dev

```bash
# backend tests
go test ./...

# backend only (needs postgres reachable)
DB_URL=postgres://postgres:password@localhost:5432/gojobberx?sslmode=disable go run .

# frontend dev server
cd frontend && npm install && npm run dev
```

## Repo layout

```
.
├── main.go            # http server, signal handling, shutdown sequence
├── handlers.go        # gin handlers, request validation
├── worker.go          # worker pool, priority dispatcher, retry loop
├── recovery.go        # boot-time requeue of pending/processing jobs
├── job.go             # Job/DeadJob structs, db crud, queueByPriority
├── db.go              # pgx pool init, dlq insert
├── metrics.go         # prometheus collectors
├── schema.sql         # jobs + dead_jobs tables (auto-loaded by postgres)
├── db/migrations/     # versioned migrations (not auto-applied at runtime)
├── frontend/          # react + vite dashboard
└── docker-compose.yml # postgres + backend + frontend
```
