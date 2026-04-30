package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	InitDB()
	go priorityDispatcher()
	StartWorkerPool(3)
	RecoverPendingJobs()

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.POST("/job", EnqueueJob)
	r.GET("/job/:id", GetJobStatus)
	r.GET("/jobs", ListJobs)
	r.GET("/dead-jobs", ListDeadJobs)
	r.GET("/metrics", gin.WrapH(prometheusHandler()))

	r.Run(":8080")
}
