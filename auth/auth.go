package auth

import (
	"errors"
	"os"
	"time"

	"github.com/froggu-tantei/ToT/db/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims defines the JWT claim structure
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user database.User) (string, error) {
	// Get JWT secret from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET must be set in environment")
	}

	// Get JWT expiry from environment or use default
	jwtExpiry := os.Getenv("JWT_EXPIRY")
	if jwtExpiry == "" {
		jwtExpiry = "24h" // Default expiry time
	}

	// Parse expiration duration
	expiryDuration, err := time.ParseDuration(jwtExpiry)
	if err != nil {
		return "", err
	}

	// Set claims
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tot-api",
			Subject:   user.ID.String(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete token as a string
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET must be set in environment")
	}

	// Parse token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Get claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
