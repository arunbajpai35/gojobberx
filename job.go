package main

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// JobStatus represents the state of a job
type JobStatus string

const (
	StatusQueued    JobStatus = "queued"
	StatusRunning   JobStatus = "processing"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

// Job represents a task to be processed
type Job struct {
	ID         string    `json:"id"`
	Payload    string    `json:"payload"`
	Type       string    `json:"type"`     // "send_email", "generate_invoice"
	Duration   int       `json:"duration"` // Simulated work time (in seconds)
	Status     JobStatus `json:"status"`
	Retries    int       `json:"retries"`
	MaxRetries int       `json:"max_retries"`
	Priority   string    `json:"priority"` // "high", "medium", "low"
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DeadJob struct {
	ID         uuid.UUID `json:"id"`
	Payload    string    `json:"payload"`
	Type       string    `json:"type"`
	Duration   int       `json:"duration"`
	Retries    int       `json:"retries"`
	MaxRetries int       `json:"max_retries"`
	Priority   string    `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
	FailedAt   time.Time `json:"failed_at"`
}

// generateID creates a unique job ID
func generateID() string {
	return uuid.New().String()
}

// SaveJob inserts a new job into PostgreSQL
func SaveJob(job *Job) error {
	_, err := DB.Exec(context.Background(), `
		INSERT INTO jobs (id, payload, type, duration, status, retries, max_retries, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, job.ID, job.Payload, job.Type, job.Duration, job.Status, job.Retries, job.MaxRetries, job.Priority, job.CreatedAt, job.CreatedAt)
	return err
}

// UpdateJobStatus updates the status and retries of a job
func UpdateJobStatus(job *Job) error {
	_, err := DB.Exec(context.Background(), `
		UPDATE jobs SET status=$1, retries=$2, updated_at=now()
		WHERE id=$3
	`, job.Status, job.Retries, job.ID)
	return err
}

// GetJobByID fetches a single job by ID from PostgreSQL
func GetJobByID(id string) (*Job, error) {
	var job Job
	err := DB.QueryRow(context.Background(), `
		SELECT id, payload, type, duration, status, retries, max_retries, priority, created_at, updated_at
		FROM jobs WHERE id=$1
	`, id).Scan(
		&job.ID, &job.Payload, &job.Type, &job.Duration,
		&job.Status, &job.Retries, &job.MaxRetries,
		&job.Priority, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetAllJobs fetches all jobs from the DB
func GetAllJobs() ([]*Job, error) {
	rows, err := DB.Query(context.Background(), `
		SELECT id, payload, type, duration, status, retries, max_retries, priority, created_at, updated_at
		FROM jobs ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		var job Job
		err := rows.Scan(
			&job.ID, &job.Payload, &job.Type, &job.Duration,
			&job.Status, &job.Retries, &job.MaxRetries,
			&job.Priority, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// queueByPriority puts the job in the appropriate queue
func queueByPriority(job *Job) {
	switch job.Priority {
	case "high":
		highQueue <- job
	case "low":
		lowQueue <- job
	default:
		mediumQueue <- job
	}
}
