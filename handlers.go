package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type enqueueRequest struct {
	Payload  string `json:"payload"`
	Duration int    `json:"duration"`
	Type     string `json:"type"`
	Priority string `json:"priority"`
}

func EnqueueJob(c *gin.Context) {
	var req enqueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if msg := validateEnqueue(&req); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if req.Priority == "" {
		req.Priority = "medium"
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
		slog.Error("save job failed", "job_id", job.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	queueByPriority(job)
	jobsTotal.WithLabelValues(string(StatusQueued)).Inc()
	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID})
}

var (
	validTypes      = map[string]bool{"send_email": true, "generate_invoice": true}
	validPriorities = map[string]bool{"": true, "high": true, "medium": true, "low": true}
)

const maxDurationSec = 600

func validateEnqueue(r *enqueueRequest) string {
	if r.Payload == "" {
		return "payload is required"
	}
	if !validTypes[r.Type] {
		return "type must be send_email or generate_invoice"
	}
	if r.Duration < 0 || r.Duration > maxDurationSec {
		return "duration must be between 0 and 600 seconds"
	}
	if !validPriorities[r.Priority] {
		return "priority must be high, medium, or low"
	}
	return ""
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
		jobs = []*Job{}
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
			slog.Warn("scan dead job failed", "error", err)
			continue
		}
		deadJobs = append(deadJobs, job)
	}

	c.JSON(200, deadJobs)
}

