package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

func EnqueueJob(c *gin.Context) {
	var req struct {
		Payload  string `json:"payload"`
		Duration int    `json:"duration"`
		Type     string `json:"type"`
		Priority string `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	job := &Job{
		ID:         generateID(),
		Payload:    req.Payload,
		Duration:   req.Duration,
		Type:       req.Type,
		Priority:   req.Priority,
		Status:     StatusQueued,
		Retries:    0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	if err := SaveJob(job); err != nil {
		log.Printf("❌ Failed to save job: %v", err) // <-- Add this
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	queueByPriority(job)
	jobsTotal.WithLabelValues(string(StatusQueued)).Inc()
	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID})
}

func GetJobStatus(c *gin.Context) {
	id := c.Param("id")
	job, err := GetJobByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func ListJobs(c *gin.Context) {
	jobs, err := GetAllJobs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch jobs"})
		return
	}

	if jobs == nil {
		jobs = []*Job{} // ✅ Ensure empty array, not null
	}

	c.JSON(http.StatusOK, jobs)
}

func ListDeadJobs(c *gin.Context) {
	rows, err := DB.Query(context.Background(), `
		SELECT id, payload, type, duration, retries, max_retries, priority, created_at, failed_at
		FROM dead_jobs ORDER BY failed_at DESC`)
	if err != nil {
		c.JSON(500, gin.H{"error": "DB error"})
		return
	}
	defer rows.Close()

	var deadJobs []DeadJob
	for rows.Next() {
		var job DeadJob
		err := rows.Scan(
			&job.ID,
			&job.Payload,
			&job.Type,
			&job.Duration,
			&job.Retries,
			&job.MaxRetries,
			&job.Priority,
			&job.CreatedAt,
			&job.FailedAt,
		)
		if err != nil {
			log.Printf("Failed to scan dead job: %v", err)
			continue
		}
		deadJobs = append(deadJobs, job)
	}

	c.JSON(200, deadJobs)
}

func EnqueueJobToDB(job Job) (string, error) {
	id := uuid.New().String()

	_, err := DB.Exec(context.Background(), `
		INSERT INTO jobs (id, payload, type, priority, duration, retries, max_retries, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 0, 3, 'queued', now())
	`, id, job.Payload, job.Type, job.Priority, job.Duration)

	if err != nil {
		return "", err
	}
	return id, nil
}
