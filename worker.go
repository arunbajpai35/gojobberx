package main

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

var jobQueue = make(chan *Job, 100)

var (
	highQueue   = make(chan *Job, 100)
	mediumQueue = make(chan *Job, 100)
	lowQueue    = make(chan *Job, 100)
)

var (
	shutdownCh = make(chan struct{})
	workerWG   sync.WaitGroup
)

func StartWorkerPool(n int) {
	for i := 0; i < n; i++ {
		workerWG.Add(1)
		go worker(i)
	}
	log.Printf("👷 Started %d workers\n", n)
}

func worker(id int) {
	defer workerWG.Done()
	for job := range jobQueue {
		processJob(id, job)
	}
}

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

	job.Retries++
	if job.Retries > job.MaxRetries {
		job.Status = StatusFailed
		UpdateJobStatus(job)
		log.Printf("❌ Job %s failed after %d retries", job.ID, job.Retries)
		jobsTotal.WithLabelValues(string(StatusFailed)).Inc()

		err := SaveToDeadLetterQueue(job)
		if err != nil {
			log.Printf("⚠️ Failed to insert into dead_jobs: %v", err)
		} else {
			log.Printf("☠️ Job %s moved to DLQ", job.ID)
		}
		return
	}

	backoff := exponentialBackoff(job.Retries)
	log.Printf("🔁 Retrying job %s after %v (attempt %d)", job.ID, backoff, job.Retries)

	time.AfterFunc(backoff, func() {
		queueByPriority(job)
	})
}

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

	time.Sleep(time.Duration(job.Duration) * time.Second)

	if job.Retries == 0 && time.Now().UnixNano()%10 < 3 {
		return false
	}

	return true
}

func priorityDispatcher() {
	defer close(jobQueue)
	for {
		var job *Job
		select {
		case <-shutdownCh:
			return
		case job = <-highQueue:
		default:
		}
		if job == nil {
			select {
			case <-shutdownCh:
				return
			case job = <-mediumQueue:
			default:
			}
		}
		if job == nil {
			select {
			case <-shutdownCh:
				return
			case job = <-lowQueue:
			default:
			}
		}
		if job == nil {
			select {
			case <-shutdownCh:
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}

		select {
		case jobQueue <- job:
		case <-shutdownCh:
			return
		}
	}
}

func exponentialBackoff(attempt int) time.Duration {
	base := time.Second
	jitter := time.Duration(rand.Intn(500)) * time.Millisecond
	return time.Duration(math.Pow(2, float64(attempt)))*base + jitter
}
