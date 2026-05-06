package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

var DB *pgxpool.Pool

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DB_URL")
	}
	if dsn == "" {
		slog.Error("DATABASE_URL or DB_URL not set")
		os.Exit(1)
	}

	var err error
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("db connect failed", "error", err)
		os.Exit(1)
	}

	if err := DB.Ping(context.Background()); err != nil {
		slog.Error("db ping failed", "error", err)
		os.Exit(1)
	}

	if _, err := DB.Exec(context.Background(), schemaSQL); err != nil {
		slog.Error("schema apply failed", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to postgres")
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
