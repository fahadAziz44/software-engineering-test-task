package service

import (
	"cruder/internal/errors"
	"cruder/internal/model"
	"cruder/internal/repository"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id uuid.UUID) (*model.User, error)
	Create(req *model.CreateUserRequest) (*model.User, error)
	Update(id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error)
	Delete(id uuid.UUID) error
}

type userService struct {
	repo repository.UserRepository
	// fullNamePattern caches the compiled regex for validating full names for performance gains (initialized once for efficiency).
	fullNamePattern *regexp.Regexp
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
		// Compile regex pattern once for performance (expensive operation)
		fullNamePattern: regexp.MustCompile(`^[a-zA-Z\s\-']+$`),
	}
}

func (s *userService) GetAll() ([]model.User, error) {
	return s.repo.GetAll()
}

func (s *userService) GetByUsername(username string) (*model.User, error) {
	// assuming username is case insensitive can serve as good example of business logic being validated in service layer.
	normalizedUsername := strings.TrimSpace(strings.ToLower(username))
	var user, err = s.repo.GetByUsername(normalizedUsername)
	if err != nil {
		// Repository layer maps storage errors to domain errors; service simply propagates.
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByID(id uuid.UUID) (*model.User, error) {
	var user, err = s.repo.GetByID(id)
	if err != nil {
		// Repository layer maps storage errors to domain errors; service simply propagates.
		return nil, err
	}
	return user, nil
}

func (s *userService) Create(req *model.CreateUserRequest) (*model.User, error) {
	// Normalize input
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	// Business logic validation: full name must contain only letters, spaces, hyphens, and apostrophes
	if err := s.validateFullName(req.FullName); err != nil {
		return nil, err
	}

	// Create user in repository
	user, err := s.repo.Create(req)
	if err != nil {
		// Repository layer maps storage error to domain errors; service simply propagates.
		return nil, err
	}

	return user, nil
}

// controller layer already does format check for alphanumeric characters and email format
// Stating my assumption here that, if we have any checl/validation related to business logic , we can add it here.
// forexample if we have a rule that validateFullName validates that a full name contains only allowed characters
// Allowed: letters, spaces, hyphens, and apostrophes
func (s *userService) validateFullName(fullName string) error {
	if !s.fullNamePattern.MatchString(fullName) {
		return fmt.Errorf("%w: full name must contain only letters, spaces, hyphens, and apostrophes", errors.ErrInvalidInput)
	}
	return nil
}

func (s *userService) Update(id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	// Normalize input for fields that are present
	if req.Username != nil {
		normalized := strings.TrimSpace(strings.ToLower(*req.Username))
		req.Username = &normalized
	}
	if req.Email != nil {
		normalized := strings.TrimSpace(strings.ToLower(*req.Email))
		req.Email = &normalized
	}
	if req.FullName != nil {
		normalized := strings.TrimSpace(*req.FullName)
		req.FullName = &normalized

		// Business logic validation: full name must contain only letters, spaces, hyphens, and apostrophes
		if err := s.validateFullName(*req.FullName); err != nil {
			return nil, err
		}
	}

	// Update user in repository
	user, err := s.repo.Update(id, req)
	if err != nil {
		// Repository layer maps storage errors to domain errors; service simply propagates.
		return nil, err
	}

	return user, nil
}

func (s *userService) Delete(id uuid.UUID) error {
	// Delete user from repository
	return s.repo.Delete(id)
}
