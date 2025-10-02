package httproutes

import (
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(app *pocketbase.PocketBase, e *core.ServeEvent) error {
	// POST /api/auth/register - User registration
	e.Router.POST("/api/auth/register", func(e *core.RequestEvent) error {
		data := auth.RegisterRequest{}

		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.InvalidJSON,
			})
		}

		// Validate input
		if err := auth.ValidateEmail(data.Email); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		if err := auth.ValidatePassword(data.Password); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Create user
		user, err := auth.CreateUser(app, data.Email, data.Password)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set the authenticated user in the request context for immediate login
		e.Auth = user

		// Generate JWT token manually
		secret := "pocketbase-secret" // Use a consistent secret

		claims := &jwt.MapClaims{
			"id":           user.Id,
			"type":         "authRecord",
			"collectionId": user.Collection().Id,
			"exp":          time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		}

		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		token, err := tokenObj.SignedString([]byte(secret))
		if err != nil {
			log.Printf("Failed to generate auth token: %v", err)
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: "Failed to generate authentication token",
			})
		}

		// Return success response with token
		return e.JSON(http.StatusCreated, auth.AuthResponse{
			User: auth.User{
				ID:       user.Id,
				Email:    user.Email(),
				Verified: user.GetBool("verified"),
			},
			Token: token,
		})
	})

	// Add debug logging to help troubleshoot
	log.Printf("Registering auth routes: /api/auth/login")

	// POST /api/auth/login - User login
	e.Router.POST("/api/auth/login", func(e *core.RequestEvent) error {
		data := auth.LoginRequest{}

		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.InvalidJSON,
			})
		}

		// Validate input
		if err := auth.ValidateEmail(data.Email); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		if err := auth.ValidatePassword(data.Password); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Authenticate user
		user, err := auth.AuthenticateUser(app, data.Email, data.Password)
		if err != nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set the authenticated user in the request context for session persistence
		e.Auth = user

		// Generate JWT token manually
		secret := "pocketbase-secret" // Use a consistent secret

		claims := &jwt.MapClaims{
			"id":           user.Id,
			"type":         "authRecord",
			"collectionId": user.Collection().Id,
			"exp":          time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		}

		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		token, err := tokenObj.SignedString([]byte(secret))
		if err != nil {
			log.Printf("Failed to generate auth token: %v", err)
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: "Failed to generate authentication token",
			})
		}

		// Return success response with token
		return e.JSON(http.StatusOK, auth.AuthResponse{
			User: auth.User{
				ID:       user.Id,
				Email:    user.Email(),
				Verified: user.GetBool("verified"),
			},
			Token: token,
		})
	})

	// POST /api/auth/logout - User logout
	e.Router.POST("/api/auth/logout", auth.RequireAuth(func(e *core.RequestEvent) error {
		// In PocketBase, logout is typically handled client-side by removing the token
		// But we can provide a logout endpoint for consistency
		return e.JSON(http.StatusOK, map[string]string{
			constants.JSONMessage: "Logout successful",
		})
	}))

	// GET /api/auth/me - Get current user info
	e.Router.GET("/api/auth/me", auth.RequireAuth(func(e *core.RequestEvent) error {
		user, err := auth.GetUserFromRequest(e)
		if err != nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONUser: user,
		})
	}))

	return nil
}
