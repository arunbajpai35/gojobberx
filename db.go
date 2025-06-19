package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("❌ DB_URL not set in environment")
	}

	var err error
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")
}

func SaveToDeadLetterQueue(job *Job) error {
	_, err := DB.Exec(context.Background(), `
		INSERT INTO dead_jobs (id, payload, type, duration, retries, max_retries, priority, created_at, failed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`,
		job.ID, job.Payload, job.Type, job.Duration, job.Retries, job.MaxRetries, job.Priority, job.CreatedAt,
	)
	return err
}
