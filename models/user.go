package models

import (
	"time"

	"github.com/XEDJK/ToT/db/database"
	"github.com/google/uuid"
)

// User represents the API-friendly user model
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRequest represents the request payload for user-related operations
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Username string `json:"username" validate:"required,min=3"`
}

// UpdateUserRequest represents the request payload for updating user information
type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=6"`
	Username string `json:"username" validate:"omitempty,min=3"`
}

// DatabaseUserToUser converts a database user to an API user
func DatabaseUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}
}

// Multiple conversion helper for slices of users
func DatabaseUsersToUsers(dbUsers []database.User) []User {
	users := make([]User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = DatabaseUserToUser(dbUser)
	}
	return users
}
