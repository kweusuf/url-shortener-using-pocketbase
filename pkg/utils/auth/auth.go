package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New(constants.ErrorEmailRequired)
	}
	if !strings.Contains(email, "@") {
		return errors.New(constants.ErrorInvalidEmail)
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New(constants.ErrorPasswordRequired)
	}
	if len(password) < 6 {
		return errors.New(constants.ErrorPasswordTooShort)
	}
	return nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// CreateUser creates a new user in PocketBase
func CreateUser(app *pocketbase.PocketBase, email, password string) (*core.Record, error) {
	// Get the users collection
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, _ := app.FindFirstRecordByFilter(
		collection,
		"email = {:email}",
		map[string]interface{}{"email": email},
	)
	if existingUser != nil {
		return nil, errors.New(constants.ErrorUserExists)
	}

	// Create user record using PocketBase's auth collection
	user := core.NewRecord(collection)
	user.SetEmail(email)
	user.SetPassword(password)
	user.Set("verified", false)

	// Save user
	if err := app.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func AuthenticateUser(app *pocketbase.PocketBase, email, password string) (*core.Record, error) {
	// Get the users collection
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		log.Info("AuthenticateUser: Failed to get users collection: %v", err)
		return nil, err
	}

	log.Info("AuthenticateUser: Looking for user with email: %s", email)

	// Find user by email - try to find auth record directly
	record, err := app.FindFirstRecordByFilter(
		collection,
		"email = '"+email+"'",
		nil,
	)
	if err != nil {
		log.Info("AuthenticateUser: FindFirstRecordByFilter failed: %v", err)
		return nil, errors.New(constants.ErrorInvalidCredentials)
	}

	if record == nil {
		log.Info("AuthenticateUser: No record found for email: %s", email)
		return nil, errors.New(constants.ErrorInvalidCredentials)
	}

	log.Info("AuthenticateUser: Record found, ID: %s", record.Id)

	// Debug: Log all fields in the record
	log.Info("AuthenticateUser: Record fields: %+v", record)

	// Verify password hash directly
	storedHash := record.GetString("password")
	log.Info("AuthenticateUser: Password field value: '%s'", storedHash)
	if storedHash == "" {
		log.Info("AuthenticateUser: Stored hash is empty - checking other password fields")
		// Try other possible field names
		storedHash = record.GetString("passwd")
		if storedHash == "" {
			storedHash = record.GetString("password_hash")
		}
		if storedHash == "" {
			log.Info("AuthenticateUser: No password hash found in any field")
			return nil, errors.New(constants.ErrorInvalidCredentials)
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		log.Info("AuthenticateUser: Password verification failed: %v", err)
		return nil, errors.New(constants.ErrorInvalidCredentials)
	}

	log.Info("AuthenticateUser: Password verification successful")
	return record, nil
}

// GetUserFromRequest extracts user information from request context
func GetUserFromRequest(e *core.RequestEvent) (*User, error) {
	// Try to get auth record from request context
	authRecord := e.Auth
	if authRecord == nil {
		return nil, errors.New(constants.ErrorUnauthorized)
	}

	// Check if record is from users collection
	if authRecord.Collection().Name != "users" {
		return nil, errors.New(constants.ErrorUnauthorized)
	}

	user := &User{
		ID:       authRecord.Id,
		Email:    authRecord.Email(),
		Verified: authRecord.GetBool("verified"),
	}

	return user, nil
}

// RequireAuth middleware function to protect routes
func RequireAuth(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Check if user is authenticated via PocketBase auth
		authRecord := e.Auth
		if authRecord == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				constants.JSONError: constants.ErrorSessionRequired,
			})
		}

		// Check if auth record is from users collection
		if authRecord.Collection().Name != "users" {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				constants.JSONError: constants.ErrorUnauthorized,
			})
		}

		// Call the next handler
		return next(e)
	}
}
