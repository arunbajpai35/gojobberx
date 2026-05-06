package main

import (
	"log/slog"
	"math"
	"math/rand"
	"strings"
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
	slog.Info("workers started", "count", n)
}

func worker(id int) {
	defer workerWG.Done()
	for job := range jobQueue {
		processJob(id, job)
	}
}

func processJob(workerID int, job *Job) {
	log := slog.With("worker", workerID, "job_id", job.ID, "type", job.Type)
	log.Info("job picked")

	job.Status = StatusRunning
	UpdateJobStatus(job)

	success := executeJob(job)

	if success {
		job.Status = StatusCompleted
		UpdateJobStatus(job)
		log.Info("job completed")
		jobsTotal.WithLabelValues(string(StatusCompleted)).Inc()
		return
	}

	job.Retries++
	if job.Retries > job.MaxRetries {
		job.Status = StatusFailed
		UpdateJobStatus(job)
		log.Error("job failed", "retries", job.Retries)
		jobsTotal.WithLabelValues(string(StatusFailed)).Inc()

		if err := SaveToDeadLetterQueue(job); err != nil {
			log.Error("dlq insert failed", "error", err)
		} else {
			log.Warn("job moved to dlq")
		}
		return
	}

	backoff := exponentialBackoff(job.Retries)
	log.Info("job retry scheduled", "backoff", backoff, "attempt", job.Retries)

	time.AfterFunc(backoff, func() {
		queueByPriority(job)
	})
}

func executeJob(job *Job) bool {
	switch job.Type {
	case "send_email":
		slog.Info("sending email", "job_id", job.ID, "payload", job.Payload)
	case "generate_invoice":
		slog.Info("generating invoice", "job_id", job.ID, "payload", job.Payload)
	default:
		slog.Warn("unknown job type", "job_id", job.ID, "type", job.Type)
		return false
	}

	time.Sleep(time.Duration(job.Duration) * time.Second)

	// payloads starting with "fail" always fail — used to demo the dlq flow
	if strings.HasPrefix(job.Payload, "fail") {
		return false
	}

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
