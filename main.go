package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	InitDB()
	go priorityDispatcher()
	StartWorkerPool(3)
	startMetricsObserver()
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Warn("http shutdown failed", "error", err)
	}

	close(shutdownCh)
	workerWG.Wait()
	DB.Close()
	slog.Info("shutdown complete")
}
