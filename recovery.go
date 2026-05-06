package main

import (
	"context"
	"log/slog"
)

func RecoverPendingJobs() {
	ctx := context.Background()

	if _, err := DB.Exec(ctx, `UPDATE jobs SET status='queued', updated_at=now() WHERE status='processing'`); err != nil {
		slog.Warn("recovery: reset processing failed", "error", err)
		return
	}

	rows, err := DB.Query(ctx, `
		SELECT id, payload, type, duration, status, retries, max_retries, priority, created_at, updated_at
		FROM jobs WHERE status='queued' ORDER BY created_at ASC
	`)
	if err != nil {
		slog.Warn("recovery: load pending failed", "error", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var job Job
		if err := rows.Scan(
			&job.ID, &job.Payload, &job.Type, &job.Duration,
			&job.Status, &job.Retries, &job.MaxRetries,
			&job.Priority, &job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			slog.Warn("recovery: scan failed", "error", err)
			continue
		}
		queueByPriority(&job)
		count++
	}

	if count > 0 {
		slog.Info("recovery complete", "count", count)
	}
}
