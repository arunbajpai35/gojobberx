package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ http server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ http shutdown: %v", err)
	}

	close(shutdownCh)
	workerWG.Wait()
	DB.Close()
	log.Println("✅ shutdown complete")
}
