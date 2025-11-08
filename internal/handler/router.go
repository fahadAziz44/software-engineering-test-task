package handler

import (
	"cruder/internal/controller"

	"github.com/gin-gonic/gin"
)

func New(router *gin.Engine, userController *controller.UserController, healthController *controller.HealthController) *gin.Engine {
	// Health endpoints for Kubernetes probes and NO authentication required
	router.GET("/health", healthController.LivenessProbe)
	router.GET("/ready", healthController.ReadinessProbe)

	v1 := router.Group("/api/v1")
	{
		userGroup := v1.Group("/users")
		{
			userGroup.GET("", userController.GetAllUsers)
			userGroup.GET("/username/:username", userController.GetUserByUsername)
			userGroup.GET("/id/:id", userController.GetUserByID)
			userGroup.POST("", userController.CreateUser)
			userGroup.PATCH("/id/:id", userController.UpdateUser)
			userGroup.DELETE("/id/:id", userController.DeleteUser)
		}
	}
	return router
}
