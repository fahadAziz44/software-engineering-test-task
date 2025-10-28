package repository

import (
	"context"
	"cruder/internal/errors"
	"cruder/internal/model"
	"database/sql"
)

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.QueryContext(context.Background(), `SELECT id, username, email, full_name, created_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.CreatedAt); err != nil {
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
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, username, email, full_name, created_at FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(id int64) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, username, email, full_name, created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}
