package errors

import "errors"

// Domain/Business Logic Errors
var (
	// User-related errors
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidInput = errors.New("invalid input")

	// Uniqueness constraint errors (from repository/database layer)
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")

	// Database errors
	ErrDatabaseOperation = errors.New("database operation failed")
)
