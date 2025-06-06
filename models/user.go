package models

import (
	"time"

	"github.com/froggu-tantei/ToT/db/database"
	"github.com/google/uuid"
)

// User represents the API-friendly user model
type User struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastPlaceCount int       `json:"last_place_count"`
	ProfilePicture string    `json:"profile_picture,omitempty"`
	Bio            string    `json:"bio,omitempty"`
}

// UserRequest represents the request payload for user-related operations
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Username string `json:"username" validate:"required,min=2"`
	Bio      string `json:"bio" validate:"omitempty,max=200"`
}

// UpdateUserRequest represents the request payload for updating user information
type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=6"`
	Username string `json:"username" validate:"omitempty,min=2"`
	Bio      string `json:"bio" validate:"omitempty,max=200"`
}

// DatabaseUserToUser converts a database user to an API user
func DatabaseUserToUser(dbUser database.User) User {
	return User{
		ID:             dbUser.ID,
		Username:       dbUser.Username,
		Email:          dbUser.Email,
		CreatedAt:      dbUser.CreatedAt.Time,
		UpdatedAt:      dbUser.UpdatedAt.Time,
		LastPlaceCount: int(dbUser.LastPlaceCount),
		ProfilePicture: dbUser.ProfilePicture.String,
		Bio:            dbUser.Bio.String,
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
