package repository

import (
	"context"
	"cruder/internal/errors"
	"cruder/internal/model"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	stdErrors "errors"
)

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id uuid.UUID) (*model.User, error)
	Create(req *model.CreateUserRequest) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.QueryContext(context.Background(), `SELECT id, username, email, full_name, created_at, updated_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, username, email, full_name, created_at, updated_at FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			// translate storage errors to domain errors
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, username, email, full_name, created_at, updated_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			// translate storage errors to domain errors
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(req *model.CreateUserRequest) (*model.User, error) {
	var user model.User

	query := `
		INSERT INTO users (username, email, full_name, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, username, email, full_name, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		context.Background(),
		query,
		req.Username,
		req.Email,
		req.FullName,
	).Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// map PostgreSQL unique constraint violations to domain errors ErrUsernameExists or ErrEmailExists
		var pqErr *pq.Error
		if stdErrors.As(err, &pqErr) {
			// 23505 is the PostgreSQL error code for unique_violation
			if pqErr.Code == "23505" {
				if strings.Contains(pqErr.Message, "username") {
					return nil, errors.ErrUsernameExists
				}
				if strings.Contains(pqErr.Message, "email") {
					return nil, errors.ErrEmailExists
				}
			}
		}
		return nil, err
	}

	return &user, nil
}
