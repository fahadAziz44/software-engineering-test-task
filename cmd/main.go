package main

import (
	"cruder/internal/config"
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/middleware"
	"cruder/internal/repository"
	"cruder/internal/service"
	"log/slog"
	"os"

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
	controllers := controller.NewController(services)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.APIKeyAuth())

	handler.New(r, controllers.Users)

	addr := ":" + cfg.Server.Port
	logger.Info("Starting server",
		slog.String("address", addr),
		slog.String("environment", config.GetEnvironment()))

	if err := r.Run(addr); err != nil {
		logger.Error("Failed to start server",
			slog.String("error", err.Error()),
			slog.String("address", addr))
		os.Exit(1)
	}
}
