package repository

import (
	"context"
	"cruder/internal/errors"
	"cruder/internal/model"
	"database/sql"
	"fmt"
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
	Update(id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error)
	Delete(id uuid.UUID) error
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

func (r *userRepository) Update(id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	// Build dynamic UPDATE query based on provided fields
	updates := []string{}
	args := []interface{}{}
	argPosition := 1

	if req.Username != nil {
		updates = append(updates, fmt.Sprintf("username = $%d", argPosition))
		args = append(args, *req.Username)
		argPosition++
	}
	if req.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argPosition))
		args = append(args, *req.Email)
		argPosition++
	}
	if req.FullName != nil {
		updates = append(updates, fmt.Sprintf("full_name = $%d", argPosition))
		args = append(args, *req.FullName)
		argPosition++
	}

	// Empty update (no fields provided) - fetch and return existing user
	if len(updates) == 0 {
		return r.GetByID(id)
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
		RETURNING id, username, email, full_name, created_at, updated_at
	`, strings.Join(updates, ", "), argPosition)

	var user model.User
	err := r.db.QueryRowContext(
		context.Background(),
		query,
		args...,
	).Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// User doesn't exist - UPDATE affected 0 rows
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}

		var pqErr *pq.Error
		if stdErrors.As(err, &pqErr) {
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

// Delete removes a user by ID. Returns ErrUserNotFound if the user doesn't exist.
// This is the "informative" approach - the repository reports facts, not policy.
// The controller layer decides whether to treat non-existence as idempotent or not.
func (r *userRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(context.Background(), query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Report the fact: user didn't exist
	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	return nil // Success: 1 row was deleted
}
