package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const jwtSecret = "pocketbase-secret" // Use a consistent secret

// AuthService handles authentication-related operations
type AuthService struct {
	app *pocketbase.PocketBase
}

// NewAuthService creates a new AuthService instance
func NewAuthService(app *pocketbase.PocketBase) *AuthService {
	return &AuthService{app: app}
}

// RegisterUser handles user registration logic
func (s *AuthService) RegisterUser(email, password string) (*auth.AuthResponse, *core.Record, error) {
	// Validate input
	if err := auth.ValidateEmail(email); err != nil {
		return nil, nil, fmt.Errorf("email validation failed: %w", err)
	}

	if err := auth.ValidatePassword(password); err != nil {
		return nil, nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Create user
	user, err := auth.CreateUser(s.app, email, password)
	if err != nil {
		return nil, nil, fmt.Errorf("user creation failed: %w", err)
	}

	// Generate JWT token
	token, err := s.generateJWTToken(user)
	if err != nil {
		return nil, nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Return response and user record
	return &auth.AuthResponse{
		User: auth.User{
			ID:       user.Id,
			Email:    user.Email(),
			Verified: user.GetBool("verified"),
		},
		Token: token,
	}, user, nil
}

// LoginUser handles user login logic
func (s *AuthService) LoginUser(email, password string) (*auth.AuthResponse, *core.Record, error) {
	// Validate input
	if err := auth.ValidateEmail(email); err != nil {
		return nil, nil, fmt.Errorf("email validation failed: %w", err)
	}

	if err := auth.ValidatePassword(password); err != nil {
		return nil, nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Authenticate user
	user, err := auth.AuthenticateUser(s.app, email, password)
	if err != nil {
		return nil, nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Generate JWT token
	token, err := s.generateJWTToken(user)
	if err != nil {
		return nil, nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Return response and user record
	return &auth.AuthResponse{
		User: auth.User{
			ID:       user.Id,
			Email:    user.Email(),
			Verified: user.GetBool("verified"),
		},
		Token: token,
	}, user, nil
}

// GetCurrentUser gets the current user information
func (s *AuthService) GetCurrentUser(e *core.RequestEvent) (*auth.User, error) {
	user, err := auth.GetUserFromRequest(e)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from request: %w", err)
	}
	return user, nil
}

// generateJWTToken generates a JWT token for the user
func (s *AuthService) generateJWTToken(user *core.Record) (string, error) {
	claims := &jwt.MapClaims{
		"id":           user.Id,
		"type":         "authRecord",
		"collectionId": user.Collection().Id,
		"exp":          time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	}

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenObj.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Info("Failed to generate auth token: %v", err)
		return "", err
	}

	return token, nil
}
