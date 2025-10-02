package httproutes

import (
	"net/http"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/service"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(app *pocketbase.PocketBase, e *core.ServeEvent) error {
	authService := service.NewAuthService(app)

	// POST /api/auth/register - User registration
	e.Router.POST("/api/auth/register", func(e *core.RequestEvent) error {
		data := auth.RegisterRequest{}

		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.InvalidJSON,
			})
		}

		// Call service to register user
		response, user, err := authService.RegisterUser(data.Email, data.Password)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set the authenticated user in the request context for immediate login
		e.Auth = user

		// Return success response with token
		return e.JSON(http.StatusCreated, response)
	})

	// POST /api/auth/login - User login
	e.Router.POST("/api/auth/login", func(e *core.RequestEvent) error {
		data := auth.LoginRequest{}

		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.InvalidJSON,
			})
		}

		// Call service to authenticate user
		response, user, err := authService.LoginUser(data.Email, data.Password)
		if err != nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set the authenticated user in the request context for session persistence
		e.Auth = user

		// Return success response
		return e.JSON(http.StatusOK, response)
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
		user, err := authService.GetCurrentUser(e)
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
