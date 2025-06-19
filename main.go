package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	InitDB()
	go StartWorkerPool(3)

	r := gin.Default()
	r.Use(cors.Default()) // 👈 Add this

	r.POST("/job", EnqueueJob)
	r.GET("/job/:id", GetJobStatus)
	r.GET("/jobs", ListJobs)
	r.GET("/metrics", gin.WrapH(prometheusHandler()))
	r.GET("/dead-jobs", ListDeadJobs)

	r.Run(":8080")
}
