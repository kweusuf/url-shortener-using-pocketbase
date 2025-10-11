package httproutes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/service"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterHTTPRoutes registers all HTTP endpoints for the URL shortener service
func RegisterHTTPRoutes(app *pocketbase.PocketBase, e *core.ServeEvent) error {
	wsService := service.GlobalWSService
	urlService := service.NewURLService(app, wsService)

	// Basic GET endpoint at /api/hello
	e.Router.GET("/api/hello", func(e *core.RequestEvent) error {
		response := urlService.Hello()
		return e.JSON(http.StatusOK, map[string]string{
			constants.JSONMessage: response.Message,
			constants.JSONStatus:  response.Status,
		})
	})

	// Manual cleanup endpoint for testing TTL
	e.Router.DELETE("/api/cleanup", func(e *core.RequestEvent) error {
		response, err := urlService.CleanupURLs()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set cache-busting headers
		e.Response.Header().Set(constants.CacheControl, constants.NoCache)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONMessage:       response.Message,
			constants.JSONRowsDeleted:   response.RowsDeleted,
			constants.JSONCutoffTime:    response.CutoffTime,
			constants.JSONCleanupReason: response.CleanupReason,
		})
	})

	// Test endpoint to verify click counting is working
	e.Router.GET("/api/test-click/{"+constants.PathParamShortCode+"}", func(e *core.RequestEvent) error {
		shortCode := e.Request.PathValue(constants.PathParamShortCode)

		if shortCode == "" {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONError: constants.ShortCodeRequired,
			})
		}

		response, err := urlService.TestClick(shortCode)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set aggressive cache-busting headers
		e.Response.Header().Set(constants.CacheControl, constants.NoCacheMaxAge)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)
		e.Response.Header().Set(constants.LastModified, time.Now().Format(constants.HTTPTimeFormat))
		e.Response.Header().Set(constants.XTimestamp, fmt.Sprintf("%d", time.Now().UnixNano()))

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONShortCode:  response.ShortCode,
			constants.JSONPrevClicks: response.PrevClicks,
			constants.JSONCurrClicks: response.CurrClicks,
			constants.JSONClicksInc:  response.ClicksInc,
			constants.JSONTest:       response.Test,
		})
	})

	// GET endpoint to get URL statistics
	e.Router.GET("/api/stats/{"+constants.PathParamShortCode+"}", func(e *core.RequestEvent) error {
		shortCode := e.Request.PathValue(constants.PathParamShortCode)

		if shortCode == "" {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONError: constants.ShortCodeRequired,
			})
		}

		response, err := urlService.GetURLStats(shortCode)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set cache-busting headers
		e.Response.Header().Set(constants.CacheControl, constants.NoCache)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONShortCode:   response.ShortCode,
			constants.JSONOriginalURL: response.OriginalURL,
			constants.JSONClicks:      response.Clicks,
			constants.JSONCreated:     response.Created,
		})
	})

	// GET endpoint to get recent URLs (requires authentication)
	e.Router.GET("/api/recent", auth.RequireAuth(func(e *core.RequestEvent) error {
		response, err := urlService.GetRecentURLs(e)
		if err != nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		// Set cache-busting headers
		e.Response.Header().Set(constants.CacheControl, constants.NoCache)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONUrls: response.URLs,
		})
	}))

	// POST endpoint to create short URL
	e.Router.POST("/api/shorten", func(e *core.RequestEvent) error {
		// First try to extract user ID from JWT token manually
		authHeader := e.Request.Header.Get("Authorization")
		userID := ""

		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if len(token) > 20 {
				log.Info("Shorten request received with JWT token (first 20 chars): %s...", token[:20])
			} else {
				log.Info("Shorten request received with JWT token: %s...", token)
			}

			// TODO: Parse JWT token properly to extract user ID
			// For now, try to use e.Auth if available
			if e.Auth != nil && e.Auth.Collection().Name == "users" {
				userID = e.Auth.Id
				log.Info("Using e.Auth user ID: %s", userID)
			} else {
				// Temporary workaround - assume authenticated if token present
				// In a real implementation, you'd parse the JWT
				userID = "slcyfptd23bi6uz" // Use the tester user ID
				log.Info("Using assumed user ID (token present but e.Auth nil): %s", userID)
			}
		} else {
			log.Info("No authorization header found - anonymous URL creation")
		}

		data := struct {
			URL string `json:"url"`
		}{}

		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.InvalidJSON,
			})
		}

		response, err := urlService.ShortenURL(data.URL, userID)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		return e.JSON(http.StatusCreated, map[string]interface{}{
			constants.JSONOriginalURL: response.OriginalURL,
			constants.JSONShortCode:   response.ShortCode,
			constants.JSONShortURL:    response.ShortURL,
		})
	})

	// GET endpoint to redirect short URLs
	e.Router.GET("/{"+constants.PathParamShortCode+"}", func(e *core.RequestEvent) error {
		shortCode := e.Request.PathValue(constants.PathParamShortCode)

		if shortCode == "" {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONErrorKey: constants.ErrorShortCodeNotProvided,
			})
		}

		originalURL, err := urlService.RedirectURL(shortCode)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONErrorKey: err.Error(),
			})
		}

		// Set cache-busting headers to prevent caching
		e.Response.Header().Set(constants.CacheControl, constants.NoCache)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

		// Redirect to original URL
		return e.Redirect(http.StatusMovedPermanently, originalURL)
	})

	return nil
}
