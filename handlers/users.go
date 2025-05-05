package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/XEDJK/ToT/auth"
	"github.com/XEDJK/ToT/db/database"
	"github.com/XEDJK/ToT/middleware"
	"github.com/XEDJK/ToT/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// SignupHandler registers a new user
func (cfg *APIConfig) SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Invalid request format"))
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.Username == "" {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Email, password, and username are required"))
		return
	}

	// Check if email already exists
	_, err := cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		RespondWithJSON(w, http.StatusConflict, models.NewErrorResponse("Email already registered"))
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		// Other database error
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
		return
	}

	// Check if username already exists
	_, err = cfg.DB.GetUserByUsername(r.Context(), req.Username)
	if err == nil {
		RespondWithJSON(w, http.StatusConflict, models.NewErrorResponse("Username already taken"))
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		// Other database error
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error processing password"))
		return
	}

	// Create user in database
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Username:     req.Username,
	})
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error creating user"))
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error generating authentication token"))
		return
	}

	// Convert to API model
	userModel := models.DatabaseUserToUser(user)

	// Return the user and token
	RespondWithJSON(w, http.StatusCreated, models.NewSuccessResponse(map[string]any{
		"user":  userModel,
		"token": token,
	}))
}

// LoginHandler handles user authentication
func (cfg *APIConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Invalid request format"))
		return
	}

	// Find user by email
	user, err := cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		RespondWithJSON(w, http.StatusUnauthorized, models.NewErrorResponse("Invalid email or password"))
		return
	} else if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, models.NewErrorResponse("Invalid email or password"))
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error generating authentication token"))
		return
	}

	// Convert to API model
	userModel := models.DatabaseUserToUser(user)

	// Return user and token
	RespondWithJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]any{
		"user":  userModel,
		"token": token,
	}))
}

// GetMeHandler returns the authenticated user's profile
func (cfg *APIConfig) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by AuthMiddleware)
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		RespondWithJSON(w, http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
		return
	}

	// Get updated user data from database
	user, err := cfg.DB.GetUserByID(r.Context(), claims.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		RespondWithJSON(w, http.StatusNotFound, models.NewErrorResponse("User not found"))
		return
	} else if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
		return
	}

	// Return user data
	RespondWithJSON(w, http.StatusOK, models.NewSuccessResponse(models.DatabaseUserToUser(user)))
}

// GetUserByIDHandler returns a user by ID
func (cfg *APIConfig) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Missing user ID"))
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Invalid user ID format"))
		return
	}

	// Get user from database
	user, err := cfg.DB.GetUserByID(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		RespondWithJSON(w, http.StatusNotFound, models.NewErrorResponse("User not found"))
		return
	} else if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
		return
	}

	// Return user data
	RespondWithJSON(w, http.StatusOK, models.NewSuccessResponse(models.DatabaseUserToUser(user)))
}

// GetUserByUsernameHandler returns a user by username
func (cfg *APIConfig) GetUserByUsernameHandler(w http.ResponseWriter, r *http.Request) {
	// Extract username from path
	username := r.PathValue("username")
	if username == "" {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Missing username"))
		return
	}

	// Get user from database
	user, err := cfg.DB.GetUserByUsername(r.Context(), username)
	if errors.Is(err, pgx.ErrNoRows) {
		RespondWithJSON(w, http.StatusNotFound, models.NewErrorResponse("User not found"))
		return
	} else if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
		return
	}

	// Return user data
	RespondWithJSON(w, http.StatusOK, models.NewSuccessResponse(models.DatabaseUserToUser(user)))
}

// UpdateUserHandler updates user information
func (cfg *APIConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		RespondWithJSON(w, http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
		return
	}

	// Extract ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Missing user ID"))
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Invalid user ID format"))
		return
	}

	// Verify user is updating their own profile
	if claims.UserID != id {
		RespondWithJSON(w, http.StatusForbidden, models.NewErrorResponse("Cannot update another user's profile"))
		return
	}

	// Parse request
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Invalid request format"))
		return
	}

	// Get current user data
	currentUser, err := cfg.DB.GetUserByID(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		RespondWithJSON(w, http.StatusNotFound, models.NewErrorResponse("User not found"))
		return
	} else if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
		return
	}

	// Prepare update params
	updateParams := database.UpdateUserParams{
		ID:           id,
		Email:        currentUser.Email,        // Default to current value
		PasswordHash: currentUser.PasswordHash, // Default to current value
		Username:     currentUser.Username,     // Default to current value
	}

	// Update fields if provided
	if req.Email != "" && req.Email != currentUser.Email {
		// Check if new email is already taken
		_, err := cfg.DB.GetUserByEmail(r.Context(), req.Email)
		if err == nil {
			RespondWithJSON(w, http.StatusConflict, models.NewErrorResponse("Email already in use"))
			return
		} else if !errors.Is(err, pgx.ErrNoRows) {
			RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
			return
		}
		updateParams.Email = req.Email
	}

	if req.Username != "" && req.Username != currentUser.Username {
		// Check if new username is already taken
		_, err := cfg.DB.GetUserByUsername(r.Context(), req.Username)
		if err == nil {
			RespondWithJSON(w, http.StatusConflict, models.NewErrorResponse("Username already in use"))
			return
		} else if !errors.Is(err, pgx.ErrNoRows) {
			RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Database error"))
			return
		}
		updateParams.Username = req.Username
	}

	if req.Password != "" {
		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error processing password"))
			return
		}
		updateParams.PasswordHash = string(hashedPassword)
	}

	// Update user in database
	updatedUser, err := cfg.DB.UpdateUser(r.Context(), updateParams)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error updating user"))
		return
	}

	// Return updated user
	RespondWithJSON(w, http.StatusOK, models.NewSuccessResponse(models.DatabaseUserToUser(updatedUser)))
}

// DeleteUserHandler deletes a user account
func (cfg *APIConfig) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		RespondWithJSON(w, http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
		return
	}

	// Extract ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Missing user ID"))
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithJSON(w, http.StatusBadRequest, models.NewErrorResponse("Invalid user ID format"))
		return
	}

	// Verify user is deleting their own account
	if claims.UserID != id {
		RespondWithJSON(w, http.StatusForbidden, models.NewErrorResponse("Cannot delete another user's account"))
		return
	}

	// Delete user from database
	err = cfg.DB.DeleteUser(r.Context(), id)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error deleting user"))
		return
	}

	// Return success message
	RespondWithJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message": "User deleted successfully",
	}))
}

// ListUsersHandler returns a paginated list of users
func (cfg *APIConfig) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page := 1
	perPage := 10

	// Get page from query string
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	// Get per_page from query string
	perPageStr := r.URL.Query().Get("per_page")
	if perPageStr != "" {
		if parsedPerPage, err := strconv.Atoi(perPageStr); err == nil && parsedPerPage > 0 && parsedPerPage <= 100 {
			perPage = parsedPerPage
		}
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Get users with pagination
	users, err := cfg.DB.ListUsers(r.Context(), database.ListUsersParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error fetching users"))
		return
	}

	// Get total count for pagination
	totalCount, err := cfg.DB.CountUsers(r.Context())
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, models.NewErrorResponse("Error counting users"))
		return
	}

	// Convert database users to API models
	userModels := models.DatabaseUsersToUsers(users)

	// Return paginated response
	response := models.NewPaginatedResponse(
		userModels,
		int(totalCount),
		perPage,
		page,
	)

	RespondWithJSON(w, http.StatusOK, response)
}
