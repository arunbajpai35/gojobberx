🧩 GoJobberX

GoJobberX is a scalable, Dockerized job queue system written in Go, featuring PostgreSQL-backed persistence, a retry mechanism with exponential backoff, dead-letter queue support, and a live React dashboard. It is designed for backend developers looking to demonstrate advanced systems skills in real-world architectures.

🚀 Features

✅ Priority Queues: Jobs can be enqueued with high, medium, or low priorities.

✅ Worker Pool: Concurrent workers process jobs efficiently.

✅ Retries: Failed jobs are retried with exponential backoff.

✅ Dead Letter Queue: Persistently failed jobs are tracked.

✅ RESTful API: Simple endpoints to enqueue, monitor, and manage jobs.

✅ React Dashboard: Live frontend UI polls for job updates.

✅ Dockerized: All services managed via Docker Compose.

✅ Prometheus Metrics: Exposes a /metrics endpoint for observability.

⚖️ Tech Stack

Backend: Go + Gin

Database: PostgreSQL

Frontend: React (Vite + TailwindCSS)

DevOps: Docker, Docker Compose, Nginx

Monitoring: Prometheus-ready metrics

⚙️ Getting Started

Clone the repository

git clone https://github.com/arunbajpai35/gojobberx.git
cd gojobberx

Run the full stack

docker-compose down -v
docker-compose build --no-cache
docker-compose up -d

Access:

Backend API: http://localhost:8080

Frontend Dashboard: http://localhost

🔧 API Endpoints

Method

Endpoint

Description

POST

/job

Enqueue a new job

GET

/jobs

List all jobs

GET

/job/:id

Get job status by ID

GET

/dead-jobs

List dead-lettered jobs

GET

/metrics

Prometheus metrics

GET

/health

Health check endpoint

Example

curl -X POST http://localhost:8080/job \
 -H "Content-Type: application/json" \
 -d '{"payload":"send-email","type":"email","priority":"high"}'

💡 Architecture

Frontend (React)
    ↓
 Nginx (Proxy)
    ↓
Backend (Go + Gin) → PostgreSQL
                  → Worker Pool
                  → Dead Letter Queue

💥 Highlights for Recruiters

Designed and built concurrent job execution system with prioritization.

Used Go routines, channels, and database transactions for reliable job processing.

Dockerized full stack with PostgreSQL, Go backend, and React frontend.

Live dashboard for operational observability.

Integrated Prometheus metrics for monitoring system health and throughput.

🧳 Future Improvements

WebSocket live updates

Role-based dashboard access

Job retry scheduling via CRON

Integration with external task APIs

🙋 Contact

Made with ❤️ by Arun Bajpai.
Feel free to connect or contribute!
