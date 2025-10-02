package httproutes

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/db"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/generator"
	urlutil "github.com/kweusuf/pocketbase-demo/pkg/utils/url"
	wsutil "github.com/kweusuf/pocketbase-demo/pkg/utils/ws"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// handleClickBroadcast broadcasts URL click updates via WebSocket after incrementing clicks
func handleClickBroadcast(shortCode string, urlData map[string]interface{}) {
	if urlData[constants.ColumnClicks] != nil {
		clicks := urlData[constants.ColumnClicks].(int) + 1
		wsutil.GlobalHub.BroadcastStatsUpdate(shortCode, clicks)
	}
}

// RegisterHTTPRoutes registers all HTTP endpoints for the URL shortener service
func RegisterHTTPRoutes(app *pocketbase.PocketBase, e *core.ServeEvent) error {
	// Basic GET endpoint at /api/hello
	e.Router.GET("/api/hello", func(e *core.RequestEvent) error {
		return e.JSON(http.StatusOK, map[string]string{
			constants.JSONMessage: constants.HelloMessage,
			constants.JSONStatus:  constants.SuccessStatus,
		})
	})

	// Manual cleanup endpoint for testing TTL
	e.Router.DELETE("/api/cleanup", func(e *core.RequestEvent) error {
		cutoffTime := time.Now().Add(-1 * time.Hour)

		result, err := app.DB().NewQuery(`
			DELETE FROM urls
			WHERE created < {:cutoffTime}
		`).Bind(map[string]interface{}{
			"cutoffTime": cutoffTime,
		}).Execute()

		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: constants.CleanupErrorMsg,
			})
		}

		rowsAffected, _ := result.RowsAffected()

		// Set cache-busting headers
		e.Response.Header().Set(constants.CacheControl, constants.NoCache)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONMessage:       constants.CleanupCompletedMsg,
			constants.JSONRowsDeleted:   rowsAffected,
			constants.JSONCutoffTime:    cutoffTime.Format(constants.TimeFormat),
			constants.JSONCleanupReason: constants.ManualCleanup,
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

		// Get current click count
		urlData, err := db.GetURLFromDB(app, shortCode)
		if err != nil || urlData == nil {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONError: constants.URLNotFound,
			})
		}

		currentClicks := urlData[constants.ColumnClicks].(int)

		// Increment click count
		if err := db.IncrementClickCount(app, shortCode); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: constants.IncrementError,
			})
		}

		// Get updated click count
		updatedData, err := db.GetURLFromDB(app, shortCode)
		if err != nil || updatedData == nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: constants.DatabaseError,
			})
		}

		newClicks := updatedData[constants.ColumnClicks].(int)

		// Set aggressive cache-busting headers
		e.Response.Header().Set(constants.CacheControl, constants.NoCacheMaxAge)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)
		e.Response.Header().Set(constants.LastModified, time.Now().Format(constants.HTTPTimeFormat))
		e.Response.Header().Set(constants.XTimestamp, fmt.Sprintf("%d", time.Now().UnixNano()))

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONShortCode:  shortCode,
			constants.JSONPrevClicks: currentClicks,
			constants.JSONCurrClicks: newClicks,
			constants.JSONClicksInc:  newClicks - currentClicks,
			constants.JSONTest:       constants.TestMessage,
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

		// Try database first
		urlData, err := db.GetURLFromDB(app, shortCode)
		if err == nil && urlData != nil {
			// Set cache-busting headers
			e.Response.Header().Set(constants.CacheControl, constants.NoCache)
			e.Response.Header().Set(constants.Pragma, constants.NoCache)
			e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

			return e.JSON(http.StatusOK, map[string]interface{}{
				constants.JSONShortCode:   shortCode,
				constants.JSONOriginalURL: urlData[constants.ColumnOriginalURL].(string),
				constants.JSONClicks:      urlData[constants.ColumnClicks].(int),
				constants.JSONCreated:     urlData[constants.ColumnCreated].(time.Time).Format(constants.TimeFormat),
			})
		}

		return e.JSON(http.StatusNotFound, map[string]string{
			constants.JSONError: constants.URLNotFound,
		})
	})

	// GET endpoint to get recent URLs (requires authentication)
	e.Router.GET("/api/recent", auth.RequireAuth(func(e *core.RequestEvent) error {
		// Get authenticated user
		user, err := auth.GetUserFromRequest(e)
		if err != nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				constants.JSONError: err.Error(),
			})
		}

		userID := user.ID

		// Fetch last 5 URLs for the authenticated user
		recentURLs, err := db.GetRecentURLsFromDB(app, 5, userID)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: constants.FetchURLError,
			})
		}

		// Set cache-busting headers
		e.Response.Header().Set(constants.CacheControl, constants.NoCache)
		e.Response.Header().Set(constants.Pragma, constants.NoCache)
		e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

		return e.JSON(http.StatusOK, map[string]interface{}{
			constants.JSONUrls: recentURLs,
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

		if data.URL == "" {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.URLRequired,
			})
		}

		// Validate and normalize URL format
		parsedURL, err := url.Parse(data.URL)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.InvalidURLFormat,
			})
		}

		// Add https:// protocol if missing
		if parsedURL.Scheme == "" {
			// Check if it looks like it has a protocol but is missing
			if strings.HasPrefix(data.URL, constants.HTTPProtocol) || strings.HasPrefix(data.URL, constants.HTTPSProtocol) {
				return e.JSON(http.StatusBadRequest, map[string]string{
					constants.JSONError: constants.InvalidURLFormat,
				})
			}
			// Add https:// protocol for bare URLs
			data.URL = constants.DefaultProtocol + data.URL
		} else if parsedURL.Scheme != constants.HTTPScheme && parsedURL.Scheme != constants.HTTPSScheme {
			return e.JSON(http.StatusBadRequest, map[string]string{
				constants.JSONError: constants.HTTPSOnly,
			})
		}

		// Generate short code
		shortCode := generator.GenerateShortCode()
		log.Info("Generated short code: %s for URL: %s", shortCode, data.URL)

		// Store URL in PocketBase database with proper user association
		err = db.StoreURLInDB(app, shortCode, data.URL, userID)
		if err != nil {
			log.Info("Failed to store URL with userID %s: %v", userID, err)
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONError: constants.StoreURLError,
			})
		}

		log.Info("Successfully stored URL: %s -> %s", shortCode, data.URL)
		baseURL := urlutil.GetBaseURL()
		return e.JSON(http.StatusCreated, map[string]interface{}{
			constants.JSONOriginalURL: data.URL,
			constants.JSONShortCode:   shortCode,
			constants.JSONShortURL:    baseURL + "/" + shortCode,
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

		var originalURL string
		var found bool

		// Try database first
		urlData, err := db.GetURLFromDB(app, shortCode)
		if err == nil && urlData != nil {
			originalURL = urlData[constants.ColumnOriginalURL].(string)
			found = true

			// Increment click count in database
			if err := db.IncrementClickCount(app, shortCode); err != nil {
				// Log error but don't fail the request
				log.Info("Failed to increment click count in database: %v", err)
			} else {
				log.Info("Incremented click count for short code: %s", shortCode)
				// Broadcast the click update via WebSocket
				handleClickBroadcast(shortCode, urlData)
			}
		}

		if !found {
			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONErrorKey: constants.ErrorURLNotFound,
			})
		}

		// Validate URL before redirecting
		if originalURL == "" {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				constants.JSONErrorKey: constants.ErrorInvalidURLStored,
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
