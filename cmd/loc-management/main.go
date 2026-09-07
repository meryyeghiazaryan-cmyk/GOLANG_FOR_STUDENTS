// Package main is the entry point for the location management service.
// It exposes a REST API for updating user locations and searching by radius,
// and sends location events to the history service via gRPC.
package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/grpcclient"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/handler"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/repository"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/service"
	"github.com/training/GOLANG_FOR_STUDENTS/pkg/metrics"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	dbURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/loc_management?sslmode=disable")
	historyAddr := getEnv("HISTORY_SERVICE_ADDR", "localhost:50051")
	port := getEnv("PORT", "8080")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("open database", zap.Error(err))
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		logger.Fatal("ping database", zap.Error(err))
	}
	logger.Info("connected to database")

	grpcClient, err := grpcclient.New(historyAddr)
	if err != nil {
		logger.Fatal("connect to history service", zap.Error(err))
	}
	defer grpcClient.Close()

	repo := repository.New(db)
	svc := service.New(repo, grpcClient)
	h := handler.New(svc, logger)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))
	router.Use(metrics.RequestCounter("loc-management"))

	api := router.Group("/api/v1")
	{
		api.PUT("/users/:username/location", h.UpdateLocation)
		api.GET("/users", h.SearchByRadius)
	}
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("location management service started", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
}

func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("handled request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
