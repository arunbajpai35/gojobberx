package main

import (
	"log"
	"math"
	"math/rand"
	"time"
)

// Central job queue where priority dispatcher sends jobs
var jobQueue = make(chan *Job, 100)

// Priority queues
var (
	highQueue   = make(chan *Job, 100)
	mediumQueue = make(chan *Job, 100)
	lowQueue    = make(chan *Job, 100)
)

// StartWorkerPool spins up `n` worker goroutines
func StartWorkerPool(n int) {
	for i := 0; i < n; i++ {
		go worker(i)
	}
	log.Printf("👷 Started %d workers\n", n)
}

// worker pulls jobs from jobQueue and processes them
func worker(id int) {
	for job := range jobQueue {
		processJob(id, job)
	}
}

// processJob executes and retries the job
func processJob(workerID int, job *Job) {
	log.Printf("👷 Worker %d picked job %s [%s]", workerID, job.ID, job.Type)

	job.Status = StatusRunning
	UpdateJobStatus(job)

	success := executeJob(job)

	if success {
		job.Status = StatusCompleted
		UpdateJobStatus(job)
		log.Printf("✅ Job %s completed", job.ID)
		jobsTotal.WithLabelValues(string(StatusCompleted)).Inc()
		return
	}

	// Job failed
	job.Retries++
	if job.Retries > job.MaxRetries {
		job.Status = StatusFailed
		UpdateJobStatus(job)
		log.Printf("❌ Job %s failed after %d retries", job.ID, job.Retries)
		jobsTotal.WithLabelValues(string(StatusFailed)).Inc()

		// Save to dead letter queue
		err := SaveToDeadLetterQueue(job)
		if err != nil {
			log.Printf("⚠️ Failed to insert into dead_jobs: %v", err)
		} else {
			log.Printf("☠️ Job %s moved to DLQ", job.ID)
		}
		return
	}

	// Retry with exponential backoff + jitter
	backoff := exponentialBackoff(job.Retries)
	log.Printf("🔁 Retrying job %s after %v (attempt %d)", job.ID, backoff, job.Retries)

	time.AfterFunc(backoff, func() {
		queueByPriority(job)
	})
}

// executeJob simulates execution based on job type
func executeJob(job *Job) bool {
	switch job.Type {
	case "send_email":
		log.Printf("📧 Sending email: %s", job.Payload)
	case "generate_invoice":
		log.Printf("🧾 Generating invoice: %s", job.Payload)
	default:
		log.Printf("⚠️ Unknown job type: %s", job.Type)
		return false
	}

	// Simulate work (e.g. API call, file generation, etc.)
	time.Sleep(time.Duration(job.Duration) * time.Second)

	// Simulate a failure 30% of the time (for retry testing)
	if job.Retries == 0 && time.Now().UnixNano()%10 < 3 {
		return false
	}

	return true
}

// priorityDispatcher checks high > medium > low queues in order
func priorityDispatcher() {
	for {
		select {
		case job := <-highQueue:
			jobQueue <- job
		default:
			select {
			case job := <-mediumQueue:
				jobQueue <- job
			default:
				select {
				case job := <-lowQueue:
					jobQueue <- job
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}
}

func exponentialBackoff(attempt int) time.Duration {
	base := time.Second
	jitter := time.Duration(rand.Intn(500)) * time.Millisecond
	return time.Duration(math.Pow(2, float64(attempt)))*base + jitter
}
