package controller

import (
	"cruder/internal/errors"
	"cruder/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	stdErrors "errors"
)

type UserController struct {
	service service.UserService
}

func NewUserController(service service.UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) GetAllUsers(ctx *gin.Context) {
	users, err := c.service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *UserController) GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	user, err := c.service.GetByUsername(username)
	if err != nil {
		if stdErrors.Is(err, errors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": fmt.Sprintf("user with username '%s' not found", username),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": fmt.Sprintf("failed to retrieve user with username '%s': %v", username, err),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) GetUserByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"message": "ID must be a valid integer",
		})
		return
	}

	user, err := c.service.GetByID(id)
	if err != nil {
		if stdErrors.Is(err, errors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": fmt.Sprintf("user with id '%d' not found", id),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": fmt.Sprintf("failed to retrieve user with id '%d': %v", id, err),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
