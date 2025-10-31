package main

import (
	"cruder/internal/config"
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/repository"
	"cruder/internal/service"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn, err := cfg.BuildDSN()
	if err != nil {
		log.Fatalf("failed to build database connection string: %v", err)
	}

	dbConn, err := repository.NewPostgresConnection(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	repositories := repository.NewRepository(dbConn.DB())
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	r := gin.Default()
	handler.New(r, controllers.Users)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
