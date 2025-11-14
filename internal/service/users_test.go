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
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetAll").Return(nil, assert.AnError)

	users, err := service.GetAll()

	assert.Error(t, err)
	assert.Nil(t, users)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// GetByUsername Tests
// =============================================================================

func TestGetByUsername_Success(t *testing.T) {
	// Given: A repository that returns a user
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

	// When: Calling with mixed case and spaces
	user, err := service.GetByUsername("  JohnDoe  ")

	// Then: Should normalize and return the user
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByUsername_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetByUsername", "nonexistent").Return(nil, errors.ErrUserNotFound)

	user, err := service.GetByUsername("nonexistent")

	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByUsername_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetByUsername", "johndoe").Return(nil, assert.AnError)

	user, err := service.GetByUsername("johndoe")

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestGetByID_Success(t *testing.T) {
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

	user, err := service.GetByID(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetByID", userID).Return(nil, errors.ErrUserNotFound)

	user, err := service.GetByID(userID)

	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetByID", userID).Return(nil, assert.AnError)

	user, err := service.GetByID(userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Create Tests
// =============================================================================

func TestCreate_Success(t *testing.T) {
	// Given: A service that can create users
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

	// When: Creating a user with mixed case and spaces (service will normalize)
	req := &model.CreateUserRequest{
		Username: "  JohnDoe  ",
		Email:    "  John@Example.COM  ",
		FullName: "  John Doe  ",
	}
	user, err := service.Create(req)

	// Then: Should normalize input and create the user
	assert.NoError(t, err)
	assert.Equal(t, createdUser, user)
	mockRepo.AssertExpectations(t)
}

func TestCreate_UsernameExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.Anything).Return(nil, errors.ErrUsernameExists)

	req := &model.CreateUserRequest{
		Username: "existing",
		Email:    "new@example.com",
		FullName: "New User",
	}
	user, err := service.Create(req)

	assert.ErrorIs(t, err, errors.ErrUsernameExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestCreate_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.Anything).Return(nil, errors.ErrEmailExists)

	req := &model.CreateUserRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		FullName: "New User",
	}
	user, err := service.Create(req)

	assert.ErrorIs(t, err, errors.ErrEmailExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestCreate_InvalidFullName(t *testing.T) {
	// Business rule: full name must contain only letters, spaces, hyphens, and apostrophes
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
			req := &model.CreateUserRequest{
				Username: "johndoe",
				Email:    "john@example.com",
				FullName: tc.fullName,
			}
			user, err := service.Create(req)

			// Should reject invalid names before calling repository
			assert.ErrorIs(t, err, errors.ErrInvalidInput)
			assert.Nil(t, user)
			mockRepo.AssertNotCalled(t, "Create", mock.Anything)
		})
	}
}

func TestCreate_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.Anything).Return(nil, assert.AnError)

	req := &model.CreateUserRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		FullName: "John Doe",
	}
	user, err := service.Create(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Update Tests
// =============================================================================

func TestUpdate_Success_AllFields(t *testing.T) {
	// Given: A service that can update users
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

	mockRepo.On("Update", userID, mock.MatchedBy(func(req *model.UpdateUserRequest) bool {
		return req.Username != nil && *req.Username == newUsername &&
			req.Email != nil && *req.Email == newEmail &&
			req.FullName != nil && *req.FullName == newFullName
	})).Return(updatedUser, nil)

	// When: Updating all fields
	req := &model.UpdateUserRequest{
		Username: &newUsername,
		Email:    &newEmail,
		FullName: &newFullName,
	}
	user, err := service.Update(userID, req)

	// Then: Should update all fields
	assert.NoError(t, err)
	assert.Equal(t, updatedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_Success_PartialUpdate(t *testing.T) {
	// PATCH semantics: only update provided fields
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

	req := &model.UpdateUserRequest{
		FullName: &newFullName,
	}
	user, err := service.Update(userID, req)

	assert.NoError(t, err)
	assert.Equal(t, updatedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, errors.ErrUserNotFound)

	fullName := "New Name"
	req := &model.UpdateUserRequest{
		FullName: &fullName,
	}
	user, err := service.Update(userID, req)

	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_UsernameExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, errors.ErrUsernameExists)

	username := "existinguser"
	req := &model.UpdateUserRequest{
		Username: &username,
	}
	user, err := service.Update(userID, req)

	assert.ErrorIs(t, err, errors.ErrUsernameExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, errors.ErrEmailExists)

	email := "existing@example.com"
	req := &model.UpdateUserRequest{
		Email: &email,
	}
	user, err := service.Update(userID, req)

	assert.ErrorIs(t, err, errors.ErrEmailExists)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_InvalidFullName(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()

	fullName := "Invalid123"
	req := &model.UpdateUserRequest{
		FullName: &fullName,
	}
	user, err := service.Update(userID, req)

	assert.ErrorIs(t, err, errors.ErrInvalidInput)
	assert.Nil(t, user)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdate_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Update", userID, mock.Anything).Return(nil, assert.AnError)

	fullName := "New Name"
	req := &model.UpdateUserRequest{
		FullName: &fullName,
	}
	user, err := service.Update(userID, req)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestDelete_Success(t *testing.T) {
	// Given: A repository that can delete users
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(nil)

	// When: Deleting a user
	err := service.Delete(userID)

	// Then: Should succeed
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDelete_UserNotFound(t *testing.T) {
	// Repository reports that user doesn't exist
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(errors.ErrUserNotFound)

	err := service.Delete(userID)

	assert.ErrorIs(t, err, errors.ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestDelete_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(assert.AnError)

	err := service.Delete(userID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
