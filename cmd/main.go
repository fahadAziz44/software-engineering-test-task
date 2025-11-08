package main

import (
	"context"
	"cruder/internal/config"
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/middleware"
	"cruder/internal/repository"
	"cruder/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Creating structured JSON logger early for consistent logging
	logger := middleware.NewStructuredLogger()

	// Load all configuration from environment variables
	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("Failed to load configuration from environment",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	dsn := cfg.BuildDSN()

	dbConn, err := repository.NewPostgresConnection(dsn)
	if err != nil {
		logger.Error("Failed to connect to database",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	repositories := repository.NewRepository(dbConn.DB())
	services := service.NewService(repositories)
	controllers := controller.NewController(services, dbConn)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.APIKeyAuth())

	handler.New(r, controllers.Users, controllers.Health)

	addr := ":" + cfg.Server.Port
	logger.Info("Starting server",
		slog.String("address", addr),
		slog.String("environment", config.GetEnvironment()))

	// HTTP server with timeouts
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server",
				slog.String("error", err.Error()),
				slog.String("address", addr))
			os.Exit(1)
		}
	}()

	// Graceful shutdown: Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server gracefully...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Close database connection
	if err := dbConn.Close(); err != nil {
		logger.Error("Error closing database connection",
			slog.String("error", err.Error()))
	}

	logger.Info("Server stopped gracefully")
}
