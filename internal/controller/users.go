package controller

import (
	"cruder/internal/errors"
	"cruder/internal/model"
	"cruder/internal/service"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	stdErrors "errors"
)

type UserController struct {
	service service.UserService
}

func NewUserController(service service.UserService) *UserController {
	return &UserController{service: service}
}

// formatValidationErrors converts validator.ValidationErrors to a map of user-friendly error messages
func formatValidationErrors(ve validator.ValidationErrors) map[string]string {
	validationErrors := make(map[string]string)
	for _, fe := range ve {
		switch fe.Tag() {
		case "required":
			validationErrors[fe.Field()] = "This field is required"
		case "email":
			validationErrors[fe.Field()] = "Invalid email format"
		case "min":
			validationErrors[fe.Field()] = "Value is too short (minimum " + fe.Param() + " characters)"
		case "max":
			validationErrors[fe.Field()] = "Value is too long (maximum " + fe.Param() + " characters)"
		case "alphanum":
			validationErrors[fe.Field()] = "Must contain only alphanumeric characters"
		default:
			validationErrors[fe.Field()] = "Invalid value"
		}
	}
	return validationErrors
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"message": "ID must be a valid UUID",
		})
		return
	}

	user, err := c.service.GetByID(id)
	if err != nil {
		if stdErrors.Is(err, errors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": fmt.Sprintf("user with id '%s' not found", id),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": fmt.Sprintf("failed to retrieve user with id '%s': %v", id, err),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	var req model.CreateUserRequest

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Handle validation errors
		var ve validator.ValidationErrors
		if stdErrors.As(err, &ve) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": "Invalid input data",
				"details": formatValidationErrors(ve),
			})
			return
		}

		// Handle JSON parsing errors
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": fmt.Sprintf("Failed to parse request body: %v", err.Error()),
		})
		return
	}

	user, err := c.service.Create(&req)
	if err != nil {
		// Handle specific business logic errors
		if stdErrors.Is(err, errors.ErrUsernameExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Username already exists",
			})
			return
		}

		if stdErrors.Is(err, errors.ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Email already exists",
			})
			return
		}

		if stdErrors.Is(err, errors.ErrInvalidInput) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid input",
				"message": err.Error(),
			})
			return
		}

		// Generic error
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": fmt.Sprintf("failed to create user: %v", err.Error()),
		})
		return
	}

	// Return created user with 201 status
	ctx.JSON(http.StatusCreated, user)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"message": "ID must be a valid UUID",
		})
		return
	}

	var req model.UpdateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if stdErrors.As(err, &ve) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": "Invalid input data",
				"details": formatValidationErrors(ve),
			})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": fmt.Sprintf("Failed to parse request body: %v", err.Error()),
		})
		return
	}

	user, err := c.service.Update(id, &req)
	if err != nil {
		if stdErrors.Is(err, errors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": fmt.Sprintf("user with id '%s' not found", id),
			})
			return
		}

		if stdErrors.Is(err, errors.ErrUsernameExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Username already exists",
			})
			return
		}

		if stdErrors.Is(err, errors.ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Email already exists",
			})
			return
		}

		if stdErrors.Is(err, errors.ErrInvalidInput) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid input",
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": fmt.Sprintf("failed to update user: %v", err.Error()),
		})
		return
	}

	// Return updated user with 200 status
	ctx.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE requests to remove a user by ID.
// Policy: Idempotent at HTTP level - returns 204 whether user existed or not.
// Why: Client's goal is to "ensure user is absent".
// Note: Repository reports facts (ErrUserNotFound), but we treat it as success here.
func (c *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"message": "ID must be a valid UUID",
		})
		return
	}

	err = c.service.Delete(id)
	if err != nil {
		// User didn't exist - still return success (idempotent behavior)
		if stdErrors.Is(err, errors.ErrUserNotFound) {
			// Log for observability: track attempts to delete non-existent users
			// This helps identify client bugs, typos, or potential probing attacks
			log.Printf("INFO: Attempted deletion of non-existent user (id=%s)", id)
			ctx.Status(http.StatusNoContent)
			return
		}

		// Real database errors
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": fmt.Sprintf("failed to delete user: %v", err.Error()),
		})
		return
	}

	// Success: user was deleted
	ctx.Status(http.StatusNoContent)
}
