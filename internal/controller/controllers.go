package controller

import (
	"cruder/internal/repository"
	"cruder/internal/service"
)

type Controller struct {
	Users  *UserController
	Health *HealthController
}

func NewController(services *service.Service, dbConn *repository.PostgresConnection) *Controller {
	return &Controller{
		Users:  NewUserController(services.Users),
		Health: NewHealthController(dbConn),
	}
}
