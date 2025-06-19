package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	jobsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "jobs_total", Help: "Total jobs processed"},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(jobsTotal)
}

func prometheusHandler() http.Handler {
	return promhttp.Handler()
}
