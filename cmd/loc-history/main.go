// Package main is the entry point for the location history service.
// It exposes a REST API for querying travel distance and a gRPC server
// that receives location events from the location management service.
package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/grpcserver"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/handler"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/repository"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/service"
	pb "github.com/training/GOLANG_FOR_STUDENTS/pkg/proto"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	dbURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5433/loc_history?sslmode=disable")
	grpcPort := getEnv("GRPC_PORT", "50051")
	httpPort := getEnv("PORT", "8081")

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

	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc, logger)
	grpcSrv := grpcserver.New(svc, logger)

	// Start the gRPC server.
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("listen for grpc", zap.Error(err))
	}
	gs := grpc.NewServer()
	pb.RegisterLocationHistoryServiceServer(gs, grpcSrv)

	go func() {
		logger.Info("grpc server started", zap.String("port", grpcPort))
		if err := gs.Serve(lis); err != nil {
			logger.Fatal("grpc server error", zap.Error(err))
		}
	}()

	// Start the HTTP server.
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))

	api := router.Group("/api/v1")
	{
		api.GET("/users/:username/distance", h.GetDistance)
	}
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("location history service started", zap.String("port", httpPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully")
	gs.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful http shutdown failed", zap.Error(err))
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
