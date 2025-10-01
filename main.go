package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/db"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/generator"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/monitoring"
	urlutil "github.com/kweusuf/pocketbase-demo/pkg/utils/url"
	wsutil "github.com/kweusuf/pocketbase-demo/pkg/utils/ws"
	"github.com/kweusuf/pocketbase-demo/transport/ws"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	// Start the cleanup scheduler for old URLs
	db.StartCleanupScheduler(app)

	// Start the WebSocket hub
	go wsutil.GlobalHub.Run()

	// Add custom routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Initialize database table
		if err := db.InitDatabase(app); err != nil {
			return err
		}

		// Register health check and monitoring routes
		if err := monitoring.RegisterHealthRoutes(app, e); err != nil {
			return err
		}

		// Register authentication routes
		if err := auth.RegisterAuthRoutes(app, e); err != nil {
			return err
		}
		// Serve static files from current directory
		e.Router.GET("/", func(e *core.RequestEvent) error {
			// Try to serve index.html first
			http.ServeFile(e.Response, e.Request, "index.html")
			return nil
		})

		// Serve other static files
		e.Router.GET("/static/*", func(e *core.RequestEvent) error {
			http.StripPrefix("/static/", http.FileServer(http.Dir(".")))
			return nil
		})

		// Basic GET endpoint at /api/hello
		e.Router.GET("/api/hello", func(e *core.RequestEvent) error {
			return e.JSON(constants.HTTPStatusOK, map[string]string{
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
				return e.JSON(constants.HTTPStatusInternalServerError, map[string]string{
					constants.JSONError: constants.CleanupErrorMsg,
				})
			}

			rowsAffected, _ := result.RowsAffected()

			// Set cache-busting headers
			e.Response.Header().Set(constants.CacheControl, constants.NoCache)
			e.Response.Header().Set(constants.Pragma, constants.NoCache)
			e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

			return e.JSON(constants.HTTPStatusOK, map[string]interface{}{
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
				return e.JSON(constants.HTTPStatusNotFound, map[string]string{
					constants.JSONError: constants.ShortCodeRequired,
				})
			}

			// Get current click count
			urlData, err := db.GetURLFromDB(app, shortCode)
			if err != nil || urlData == nil {
				return e.JSON(constants.HTTPStatusNotFound, map[string]string{
					constants.JSONError: constants.URLNotFound,
				})
			}

			currentClicks := urlData[constants.ColumnClicks].(int)

			// Increment click count
			if err := db.IncrementClickCount(app, shortCode); err != nil {
				return e.JSON(constants.HTTPStatusInternalServerError, map[string]string{
					constants.JSONError: constants.IncrementError,
				})
			}

			// Get updated click count
			updatedData, err := db.GetURLFromDB(app, shortCode)
			if err != nil || updatedData == nil {
				return e.JSON(constants.HTTPStatusInternalServerError, map[string]string{
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

			return e.JSON(constants.HTTPStatusOK, map[string]interface{}{
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
				return e.JSON(constants.HTTPStatusNotFound, map[string]string{
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

				return e.JSON(constants.HTTPStatusOK, map[string]interface{}{
					constants.JSONShortCode:   shortCode,
					constants.JSONOriginalURL: urlData[constants.ColumnOriginalURL].(string),
					constants.JSONClicks:      urlData[constants.ColumnClicks].(int),
					constants.JSONCreated:     urlData[constants.ColumnCreated].(time.Time).Format(constants.TimeFormat),
				})
			}

			return e.JSON(constants.HTTPStatusNotFound, map[string]string{
				constants.JSONError: constants.URLNotFound,
			})
		})

		// GET endpoint to get recent URLs (requires authentication)
		e.Router.GET("/api/recent", func(e *core.RequestEvent) error {
			// Try to get user info from JWT token manually since e.Auth might not be set
			authHeader := e.Request.Header.Get("Authorization")
			userID := ""

			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				// Parse JWT token to get user ID
				// For now, just log it and proceed
				if len(token) > 50 {
					log.Printf("Received JWT token (first 50 chars): %s...", token[:50])
				} else {
					log.Printf("Received JWT token: %s", token)
				}

				// TODO: Actually parse the JWT token to extract user ID
				// For now, let's hardcode getting a recent user or return empty
				if token != "" {
					// Try to find the most recent user with an auth token
					// This is a temporary workaround
					userID = "slcyfptd23bi6uz" // Hardcoded for testing
				}
			}

			// Get user ID if authenticated, otherwise use empty string for anonymous URLs
			if userID == "" && e.Auth != nil && e.Auth.Collection().Name == "users" {
				userID = e.Auth.Id
				log.Printf("Using e.Auth user ID: %s", userID)
			} else if userID != "" {
				log.Printf("Using extracted user ID: %s", userID)
			} else {
				log.Printf("No user ID found - will return recent URLs for all users")
			}

			// Fetch last 5 URLs for the authenticated user
			recentURLs, err := db.GetRecentURLsFromDB(app, 5, userID)
			if err != nil {
				return e.JSON(constants.HTTPStatusInternalServerError, map[string]string{
					constants.JSONError: constants.FetchURLError,
				})
			}

			// Set cache-busting headers
			e.Response.Header().Set(constants.CacheControl, constants.NoCache)
			e.Response.Header().Set(constants.Pragma, constants.NoCache)
			e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

			return e.JSON(constants.HTTPStatusOK, map[string]interface{}{
				constants.JSONUrls: recentURLs,
			})
		})

		// POST endpoint to create short URL
		e.Router.POST("/api/shorten", func(e *core.RequestEvent) error {
			// First try to extract user ID from JWT token manually
			authHeader := e.Request.Header.Get("Authorization")
			userID := ""

			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if len(token) > 20 {
					log.Printf("Shorten request received with JWT token (first 20 chars): %s...", token[:20])
				} else {
					log.Printf("Shorten request received with JWT token: %s...", token)
				}

				// TODO: Parse JWT token properly to extract user ID
				// For now, try to use e.Auth if available
				if e.Auth != nil && e.Auth.Collection().Name == "users" {
					userID = e.Auth.Id
					log.Printf("Using e.Auth user ID: %s", userID)
				} else {
					// Temporary workaround - assume authenticated if token present
					// In a real implementation, you'd parse the JWT
					userID = "slcyfptd23bi6uz" // Use the tester user ID
					log.Printf("Using assumed user ID (token present but e.Auth nil): %s", userID)
				}
			} else {
				log.Printf("No authorization header found - anonymous URL creation")
			}

			data := struct {
				URL string `json:"url"`
			}{}

			if err := e.BindBody(&data); err != nil {
				return e.JSON(constants.HTTPStatusBadRequest, map[string]string{
					constants.JSONError: constants.InvalidJSON,
				})
			}

			if data.URL == "" {
				return e.JSON(constants.HTTPStatusBadRequest, map[string]string{
					constants.JSONError: constants.URLRequired,
				})
			}

			// Validate and normalize URL format
			parsedURL, err := url.Parse(data.URL)
			if err != nil {
				return e.JSON(constants.HTTPStatusBadRequest, map[string]string{
					constants.JSONError: constants.InvalidURLFormat,
				})
			}

			// Add https:// protocol if missing
			if parsedURL.Scheme == "" {
				// Check if it looks like it has a protocol but is missing
				if strings.HasPrefix(data.URL, constants.HTTPProtocol) || strings.HasPrefix(data.URL, constants.HTTPSProtocol) {
					return e.JSON(constants.HTTPStatusBadRequest, map[string]string{
						constants.JSONError: constants.InvalidURLFormat,
					})
				}
				// Add https:// protocol for bare URLs
				data.URL = constants.DefaultProtocol + data.URL
			} else if parsedURL.Scheme != constants.HTTPScheme && parsedURL.Scheme != constants.HTTPSScheme {
				return e.JSON(constants.HTTPStatusBadRequest, map[string]string{
					constants.JSONError: constants.HTTPSOnly,
				})
			}

			// Generate short code
			shortCode := generator.GenerateShortCode()
			log.Printf("Generated short code: %s for URL: %s", shortCode, data.URL)

			// Store URL in PocketBase database with proper user association
			err = db.StoreURLInDB(app, shortCode, data.URL, userID)
			if err != nil {
				log.Printf("Failed to store URL with userID %s: %v", userID, err)
				return e.JSON(constants.HTTPStatusInternalServerError, map[string]string{
					constants.JSONError: constants.StoreURLError,
				})
			}

			log.Printf("Successfully stored URL: %s -> %s", shortCode, data.URL)
			baseURL := urlutil.GetBaseURL()
			return e.JSON(constants.HTTPStatusCreated, map[string]interface{}{
				constants.JSONOriginalURL: data.URL,
				constants.JSONShortCode:   shortCode,
				constants.JSONShortURL:    baseURL + "/" + shortCode,
			})
		})

		// WebSocket endpoint for real-time stats updates
		e.Router.GET("/ws", func(e *core.RequestEvent) error {
			ws.HandleWebSocket(e.Response, e.Request)
			return nil
		})

		// GET endpoint to redirect short URLs
		e.Router.GET("/{"+constants.PathParamShortCode+"}", func(e *core.RequestEvent) error {
			shortCode := e.Request.PathValue(constants.PathParamShortCode)

			if shortCode == "" {
				return e.JSON(constants.HTTPStatusNotFound, map[string]string{
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
					log.Printf("Failed to increment click count in database: %v", err)
				} else {
					log.Printf("Incremented click count for short code: %s", shortCode)
					// Broadcast the click update via WebSocket
					if urlData[constants.ColumnClicks] != nil {
						clicks := urlData[constants.ColumnClicks].(int) + 1
						wsutil.GlobalHub.BroadcastStatsUpdate(shortCode, clicks)
					}
				}
			}

			if !found {
				return e.JSON(constants.HTTPStatusNotFound, map[string]string{
					constants.JSONErrorKey: constants.ErrorURLNotFound,
				})
			}

			// Validate URL before redirecting
			if originalURL == "" {
				return e.JSON(constants.HTTPStatusInternalServerError, map[string]string{
					constants.JSONErrorKey: constants.ErrorInvalidURLStored,
				})
			}

			// Set cache-busting headers to prevent caching
			e.Response.Header().Set(constants.CacheControl, constants.NoCache)
			e.Response.Header().Set(constants.Pragma, constants.NoCache)
			e.Response.Header().Set(constants.Expires, constants.HeaderValueZero)

			// Redirect to original URL
			return e.Redirect(constants.HTTPStatusMovedPermanently, originalURL)
		})

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
