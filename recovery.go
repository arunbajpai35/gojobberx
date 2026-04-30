package main

import (
	"context"
	"log"
)

func RecoverPendingJobs() {
	ctx := context.Background()

	_, err := DB.Exec(ctx, `UPDATE jobs SET status='queued', updated_at=now() WHERE status='processing'`)
	if err != nil {
		log.Printf("⚠️ failed to reset processing jobs: %v", err)
		return
	}

	rows, err := DB.Query(ctx, `
		SELECT id, payload, type, duration, status, retries, max_retries, priority, created_at, updated_at
		FROM jobs WHERE status='queued' ORDER BY created_at ASC
	`)
	if err != nil {
		log.Printf("⚠️ failed to load pending jobs: %v", err)
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
			log.Printf("⚠️ failed to scan pending job: %v", err)
			continue
		}
		queueByPriority(&job)
		count++
	}

	if count > 0 {
		log.Printf("♻️ recovered %d pending jobs", count)
	}
}
