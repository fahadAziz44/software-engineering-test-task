package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" binding:"required,email,max=100"`
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
}
