package service

import (
	"cruder/internal/errors"
	"cruder/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetAll() ([]model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(req *model.CreateUserRequest) (*model.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// =============================================================================
// GetAll Tests
// =============================================================================

func TestGetAll_Success(t *testing.T) {
	// Given: A service with a mock repository that returns users
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	expectedUsers := []model.User{
		{
			ID:       uuid.New(),
			Username: "user1",
			Email:    "user1@example.com",
			FullName: "User One",
		},
		{
			ID:       uuid.New(),
			Username: "user2",
			Email:    "user2@example.com",
			FullName: "User Two",
		},
	}

	mockRepo.On("GetAll").Return(expectedUsers, nil)

	// When: Calling GetAll
	users, err := service.GetAll()

	// Then: Should return users and no error
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
	mockRepo.AssertExpectations(t)
}

func TestGetAll_RepositoryError(t *testing.T) {
	// Given: A service with a mock repository that returns an error
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetAll").Return(nil, assert.AnError)

	// When: Calling GetAll
	users, err := service.GetAll()

	// Then: Should return error and nil users
	assert.Error(t, err)
	assert.Nil(t, users)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// GetByUsername Tests
// =============================================================================

func TestGetByUsername_Success(t *testing.T) {
	// Given: A service with a mock repository that returns a user
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	expectedUser := &model.User{
		ID:       uuid.New(),
		Username: "johndoe",
		Email:    "john@example.com",
		FullName: "John Doe",
	}

	// Service normalizes username to lowercase and trimmed
	mockRepo.On("GetByUsername", "johndoe").Return(expectedUser, nil)

	// When: Calling GetByUsername with mixed case and spaces
	user, err := service.GetByUsername("  JohnDoe  ")

	// Then: Should return user and no error
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByUsername_UserNotFound(t *testing.T) {
	// Given: A service with a mock repository that returns ErrUserNotFound
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetByUsername", "nonexistent").Return(nil, errors.ErrUserNotFound)

	// When: Calling GetByUsername for non-existent user
	user, err := service.GetByUsername("nonexistent")

	// Then: Should return ErrUserNotFound
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByUsername_RepositoryError(t *testing.T) {
	// Given: A service with a mock repository that returns a generic error
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetByUsername", "johndoe").Return(nil, assert.AnError)

	// When: Calling GetByUsername
	user, err := service.GetByUsername("johndoe")

	// Then: Should return the error
	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestGetByID_Success(t *testing.T) {
	// Given: A service with a mock repository that returns a user
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	expectedUser := &model.User{
		ID:       userID,
		Username: "johndoe",
		Email:    "john@example.com",
		FullName: "John Doe",
	}

	mockRepo.On("GetByID", userID).Return(expectedUser, nil)

	// When: Calling GetByID with valid UUID
	user, err := service.GetByID(userID)

	// Then: Should return user and no error
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_UserNotFound(t *testing.T) {
	// Given: A service with a mock repository that returns ErrUserNotFound
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetByID", userID).Return(nil, errors.ErrUserNotFound)

	// When: Calling GetByID for non-existent user
	user, err := service.GetByID(userID)

	// Then: Should return ErrUserNotFound
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_RepositoryError(t *testing.T) {
	// Given: A service with a mock repository that returns a generic error
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetByID", userID).Return(nil, assert.AnError)

	// When: Calling GetByID
	user, err := service.GetByID(userID)

	// Then: Should return the error
	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Create Tests
// =============================================================================

func TestCreate_Success(t *testing.T) {
	// Given: A service with a mock repository that successfully creates a user
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	now := time.Now()
	createdUser := &model.User{
		ID:        uuid.New(),
		Username:  "johndoe",
		Email:     "john@example.com",
		FullName:  "John Doe",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Repository expects normalized input (lowercase username/email, trimmed)
	mockRepo.On("Create", mock.MatchedBy(func(req *model.CreateUserRequest) bool {
		return req.Username == "johndoe" &&
			req.Email == "john@example.com" &&
			req.FullName == "John Doe"
	})).Return(createdUser, nil)

	// When: Calling Create with valid request (service will normalize)
	req := &model.CreateUserRequest{
		Username: "  JohnDoe  ",          // Will be normalized
		Email:    "  John@Example.COM  ", // Will be normalized
		FullName: "  John Doe  ",         // Will be trimmed
	}
	user, err := service.Create(req)

	// Then: Should return created user and no error
	assert.NoError(t, err)
	assert.Equal(t, createdUser, user)
	mockRepo.AssertExpectations(t)
}

func TestCreate_UsernameExists(t *testing.T) {
	// Given: A service with a mock repository that returns ErrUsernameExists
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.Anything).Return(nil, errors.ErrUsernameExists)

	// When: Calling Create with duplicate username
	req := &model.CreateUserRequest{
		Username: "existing",
		Email:    "new@example.com",
		FullName: "New User",
	}
	user, err := service.Create(req)

	// Then: Should return ErrUsernameExists
	assert.ErrorIs(t, err, errors.ErrUsernameExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestCreate_EmailExists(t *testing.T) {
	// Given: A service with a mock repository that returns ErrEmailExists
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.Anything).Return(nil, errors.ErrEmailExists)

	// When: Calling Create with duplicate email
	req := &model.CreateUserRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		FullName: "New User",
	}
	user, err := service.Create(req)

	// Then: Should return ErrEmailExists
	assert.ErrorIs(t, err, errors.ErrEmailExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestCreate_InvalidFullName(t *testing.T) {
	// Given: A service with a mock repository (should not be called)
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	testCases := []struct {
		name     string
		fullName string
	}{
		{"numbers", "John123 Doe"},
		{"special chars", "John@Doe"},
		{"symbols", "John$Doe"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Calling Create with invalid full_name
			req := &model.CreateUserRequest{
				Username: "johndoe",
				Email:    "john@example.com",
				FullName: tc.fullName,
			}
			user, err := service.Create(req)

			// Then: Should return ErrInvalidInput without calling repository
			assert.ErrorIs(t, err, errors.ErrInvalidInput)
			assert.Nil(t, user)
			mockRepo.AssertNotCalled(t, "Create", mock.Anything)
		})
	}
}

func TestCreate_RepositoryError(t *testing.T) {
	// Given: A service with a mock repository that returns a generic error
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.Anything).Return(nil, assert.AnError)

	// When: Calling Create
	req := &model.CreateUserRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		FullName: "John Doe",
	}
	user, err := service.Create(req)

	// Then: Should return the error
	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Update Tests
// =============================================================================

func TestUpdate_Success_AllFields(t *testing.T) {
	// Given: A service with a mock repository that successfully updates a user
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	newUsername := "newusername"
	newEmail := "new@example.com"
	newFullName := "New Full Name"

	updatedUser := &model.User{
		ID:       userID,
		Username: newUsername,
		Email:    newEmail,
		FullName: newFullName,
	}

	// Repository expects normalized input
	mockRepo.On("Update", userID, mock.MatchedBy(func(req *model.UpdateUserRequest) bool {
		return req.Username != nil && *req.Username == newUsername &&
			req.Email != nil && *req.Email == newEmail &&
			req.FullName != nil && *req.FullName == newFullName
	})).Return(updatedUser, nil)

	// When: Calling Update with all fields
	req := &model.UpdateUserRequest{
		Username: &newUsername,
		Email:    &newEmail,
		FullName: &newFullName,
	}
	user, err := service.Update(userID, req)

	// Then: Should return updated user and no error
	assert.NoError(t, err)
	assert.Equal(t, updatedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_Success_PartialUpdate(t *testing.T) {
	// Given: A service with a mock repository
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	newFullName := "Updated Name"

	updatedUser := &model.User{
		ID:       userID,
		Username: "original",
		Email:    "original@example.com",
		FullName: newFullName,
	}

	mockRepo.On("Update", userID, mock.MatchedBy(func(req *model.UpdateUserRequest) bool {
		return req.Username == nil &&
			req.Email == nil &&
			req.FullName != nil && *req.FullName == newFullName
	})).Return(updatedUser, nil)

	// When: Calling Update with only one field
	req := &model.UpdateUserRequest{
		FullName: &newFullName,
	}
	user, err := service.Update(userID, req)

	// Then: Should succeed
	assert.NoError(t, err)
	assert.Equal(t, updatedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_UserNotFound(t *testing.T) {
	// Given: A service with a mock repository that returns ErrUserNotFound
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, errors.ErrUserNotFound)

	// When: Calling Update for non-existent user
	fullName := "New Name"
	req := &model.UpdateUserRequest{
		FullName: &fullName,
	}
	user, err := service.Update(userID, req)

	// Then: Should return ErrUserNotFound
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_UsernameExists(t *testing.T) {
	// Given: A service with a mock repository that returns ErrUsernameExists
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, errors.ErrUsernameExists)

	// When: Calling Update with duplicate username
	username := "existinguser"
	req := &model.UpdateUserRequest{
		Username: &username,
	}
	user, err := service.Update(userID, req)

	// Then: Should return ErrUsernameExists
	assert.ErrorIs(t, err, errors.ErrUsernameExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_EmailExists(t *testing.T) {
	// Given: A service with a mock repository that returns ErrEmailExists
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, errors.ErrEmailExists)

	// When: Calling Update with duplicate email
	email := "existing@example.com"
	req := &model.UpdateUserRequest{
		Email: &email,
	}
	user, err := service.Update(userID, req)

	// Then: Should return ErrEmailExists
	assert.ErrorIs(t, err, errors.ErrEmailExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_InvalidFullName(t *testing.T) {
	// Given: A service with a mock repository (should not be called)
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()

	// When: Calling Update with invalid full_name (contains numbers)
	fullName := "Invalid123"
	req := &model.UpdateUserRequest{
		FullName: &fullName,
	}
	user, err := service.Update(userID, req)

	// Then: Should return ErrInvalidInput without calling repository
	assert.ErrorIs(t, err, errors.ErrInvalidInput)
	assert.Nil(t, user)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdate_RepositoryError(t *testing.T) {
	// Given: A service with a mock repository that returns a generic error
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, assert.AnError)

	// When: Calling Update
	fullName := "New Name"
	req := &model.UpdateUserRequest{
		FullName: &fullName,
	}
	user, err := service.Update(userID, req)

	// Then: Should return the error
	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestDelete_Success(t *testing.T) {
	// Given: A service with a mock repository that successfully deletes a user
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(nil)

	// When: Calling Delete
	err := service.Delete(userID)

	// Then: Should return no error
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDelete_UserNotFound(t *testing.T) {
	// Given: A service with a mock repository that returns ErrUserNotFound (user doesn't exist)
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(errors.ErrUserNotFound)

	// When: Calling Delete on non-existent user
	err := service.Delete(userID)

	// Then: Should return ErrUserNotFound (informative - reports the fact)
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestDelete_RepositoryError(t *testing.T) {
	// Given: A service with a mock repository that returns a database error
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(assert.AnError)

	// When: Calling Delete
	err := service.Delete(userID)

	// Then: Should return the error
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
