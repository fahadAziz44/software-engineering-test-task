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
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
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
	// Additional business logic validation beyond struct tags
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}
	// Normalize input
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	// Create user in repository
	user, err := s.repo.Create(req)
	if err != nil {
		// Repository layer maps storage error  to domain errors; service simply propagates.
		return nil, err
	}

	return user, nil
}

func (s *userService) validateCreateUserRequest(req *model.CreateUserRequest) error {
	// controller layer already does format check for alphanumeric characters and email format
	// Stating my assumption , if we have any validation related to our product , we will add it here.
	// forexample if we have a rule that full name must not contain special characters except spaces, hyphens, and apostrophes we give validation error here.
	fullNameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !fullNameRegex.MatchString(req.FullName) {
		return fmt.Errorf("%w: full name must contain only letters, spaces, hyphens, and apostrophes", errors.ErrInvalidInput)
	}

	return nil
}
