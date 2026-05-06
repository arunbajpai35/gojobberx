package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	jobsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "jobs_total", Help: "Total jobs processed by terminal status"},
		[]string{"status"},
	)

	retriesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "job_retries_total", Help: "Total job retries scheduled"},
	)

	queueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "queue_depth", Help: "Current depth of each queue"},
		[]string{"queue"},
	)

	processingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_processing_duration_seconds",
			Help:    "Time spent in processJob, labeled by per-attempt outcome",
			Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120, 300, 600},
		},
		[]string{"outcome"},
	)
)

func init() {
	prometheus.MustRegister(jobsTotal, retriesTotal, queueDepth, processingDuration)
}

func startMetricsObserver() {
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-shutdownCh:
				return
			case <-t.C:
				queueDepth.WithLabelValues("high").Set(float64(len(highQueue)))
				queueDepth.WithLabelValues("medium").Set(float64(len(mediumQueue)))
				queueDepth.WithLabelValues("low").Set(float64(len(lowQueue)))
				queueDepth.WithLabelValues("main").Set(float64(len(jobQueue)))
			}
		}
	}()
}

func prometheusHandler() http.Handler {
	return promhttp.Handler()
}
