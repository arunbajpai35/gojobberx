package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	InitDB()              // Initialize PostgreSQL connection
	go StartWorkerPool(3) // Start job workers

	r := gin.Default()
	r.Use(cors.Default()) // Allow all CORS requests

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Enqueue a new job
	r.POST("/job", func(c *gin.Context) {
		var job Job
		if err := c.BindJSON(&job); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job payload"})
			return
		}

		jobID, err := EnqueueJobToDB(job) // helper function to insert into DB
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue job"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Job enqueued successfully",
			"job_id":  jobID,
		})
	})

	// Get job by ID
	r.GET("/job/:id", GetJobStatus)

	// List all jobs
	r.GET("/jobs", ListJobs)

	// Dead letter queue
	r.GET("/dead-jobs", ListDeadJobs)

	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(prometheusHandler()))

	// Run server
	r.Run(":8080")
}
