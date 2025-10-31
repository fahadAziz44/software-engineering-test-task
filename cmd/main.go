package main

import (
	"cruder/internal/config"
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/middleware"
	"cruder/internal/repository"
	"cruder/internal/service"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Creating structured JSON logger early for consistent logging
	logger := middleware.NewStructuredLogger()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load config",
			slog.String("error", err.Error()),
			slog.String("config_path", configPath))
		os.Exit(1)
	}

	dsn, err := cfg.BuildDSN()
	if err != nil {
		logger.Error("Failed to build database connection string",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

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

	handler.New(r, controllers.Users)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
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
