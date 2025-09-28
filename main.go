package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// initDatabase creates the database table for storing URLs
func initDatabase(app *pocketbase.PocketBase) error {
	log.Println(constants.InitializingDB)

	// Check if table exists
	exists, err := checkTableExists(app)
	if err != nil {
		log.Printf("Error checking if table exists: %v", err)
		return err
	}

	if !exists {
		log.Println(constants.IndexCreated)
		if err := createURLsTable(app); err != nil {
			log.Printf("Error creating URLs table: %v", err)
			return err
		}
		log.Println(constants.TableCreated)
	} else {
		log.Println(constants.TableExists)
		// Validate table schema
		valid, err := validateTableSchema(app)
		if err != nil {
			log.Printf("Error validating table schema: %v", err)
			return err
		}

		if !valid {
			log.Println(constants.TableSchemaInvalid)
			if err := dropAndRecreateTable(app); err != nil {
				log.Printf("Error recreating table: %v", err)
				return err
			}
			log.Println(constants.TableRecreated)
		} else {
			log.Println(constants.TableSchemaValid)
		}
	}

	// Create index if it doesn't exist
	if err := createIndexIfNotExists(app); err != nil {
		log.Printf("Error creating index: %v", err)
		return err
	}

	log.Println(constants.DatabaseInitialized)
	log.Println(constants.DashboardNote)
	log.Println(constants.DashboardStep1)
	log.Println(constants.DashboardStep2)
	log.Println(constants.DashboardStep3)
	log.Println(constants.DashboardStep4)
	return nil
}

// checkTableExists checks if the urls table exists
func checkTableExists(app *pocketbase.PocketBase) (bool, error) {
	var count int
	err := app.DB().NewQuery(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type='table' AND name='urls'
	`).Row(&count)

	return count > 0, err
}

// validateTableSchema checks if the table has the correct structure
func validateTableSchema(app *pocketbase.PocketBase) (bool, error) {
	var hasRequiredColumns int

	// Check if all required columns exist
	err := app.DB().NewQuery(`
		SELECT COUNT(*)
		FROM pragma_table_info('urls')
		WHERE name IN ('id', 'short_code', 'original_url', 'clicks', 'created', 'updated')
	`).Row(&hasRequiredColumns)

	if err != nil {
		return false, err
	}

	// Should have all 6 required columns
	return hasRequiredColumns == 6, nil
}

// createURLsTable creates the URLs table with correct structure
func createURLsTable(app *pocketbase.PocketBase) error {
	// _, err := app.DB().NewQuery(`
	// 	CREATE TABLE urls (
	// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 		short_code TEXT UNIQUE NOT NULL,
	// 		original_url TEXT NOT NULL,
	// 		clicks INTEGER DEFAULT 0,
	// 		created DATETIME DEFAULT CURRENT_TIMESTAMP,
	// 		updated DATETIME DEFAULT CURRENT_TIMESTAMP
	// 	)
	// `).Execute()

	// _, err := app.DB().NewQuery(`
	// 	CREATE TABLE urls (
	// 		id TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL,
	// 		short_code TEXT DEFAULT '' NOT NULL,
	// 		original_url TEXT DEFAULT '' NOT NULL,
	// 		clicks NUMERIC DEFAULT 0 NOT NULL,
	// 		created TEXT DEFAULT '' NOT NULL,
	// 		updated TEXT DEFAULT '' NOT NULL);
	// `).Execute()

	collection := core.NewBaseCollection("urls")
	// add text field
	collection.Fields.Add(
		&core.TextField{
			Name:       "id",
			Required:   true,
			PrimaryKey: true,
		},
		&core.TextField{
			Name:     "short_code",
			Required: true,
		},
		&core.TextField{
			Name:     "original_url",
			Required: true,
		},
		&core.NumberField{
			Name:     "clicks",
			Required: true,
		},
		&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		},
		&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
		},
	)
	collection.AddIndex("idx_urls_short_code", true, "short_code", "")

	err := app.Save(collection)

	return err
}

// dropAndRecreateTable drops and recreates the table
func dropAndRecreateTable(app *pocketbase.PocketBase) error {
	// Drop the table
	_, err := app.DB().NewQuery(`DROP TABLE IF EXISTS urls`).Execute()
	if err != nil {
		return err
	}

	// Recreate it
	return createURLsTable(app)
}

// createIndexIfNotExists creates the index if it doesn't exist
func createIndexIfNotExists(app *pocketbase.PocketBase) error {
	// Check if index exists
	var count int
	err := app.DB().NewQuery(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type='index' AND name='idx_urls_short_code'
	`).Row(&count)

	if err != nil {
		return err
	}

	if count == 0 {
		log.Println("Creating index on short_code...")
		_, err = app.DB().NewQuery(`
			CREATE INDEX idx_urls_short_code
			ON urls(short_code)
		`).Execute()
		return err
	}

	return nil
}

// cleanupOldURLs removes URLs older than 1 hour from the database
func cleanupOldURLs(app *pocketbase.PocketBase) error {
	// Calculate the cutoff time (1 hour ago)
	cutoffTime := time.Now().Add(-1 * time.Hour)

	result, err := app.DB().NewQuery(`
		DELETE FROM urls
		WHERE created < {:cutoffTime}
	`).Bind(map[string]interface{}{
		"cutoffTime": cutoffTime,
	}).Execute()

	if err != nil {
		return err
	}

	// Log how many rows were deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d old URLs (older than 1 hour)", rowsAffected)
	}

	return nil
}

// startCleanupScheduler starts a goroutine that cleans up old URLs every 10 minutes
func startCleanupScheduler(app *pocketbase.PocketBase) {
	// Schedule cleanup every 10 minutes (don't run immediately to avoid DB issues)
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			if err := cleanupOldURLs(app); err != nil {
				log.Printf("Error during scheduled cleanup: %v", err)
			}
		}
	}()

	log.Println("URL cleanup scheduler started (runs every 10 minutes)")
}

// getBaseURL returns the base URL for the application
func getBaseURL() string {
	// Check for environment variable first
	if baseURL := os.Getenv(constants.EnvBaseURL); baseURL != "" {
		return strings.TrimSuffix(baseURL, "/")
	}

	// Default to localhost for development
	return constants.DefaultBaseURL
}

// getAPIBaseURL returns the API base URL for the application
func getAPIBaseURL() string {
	baseURL := getBaseURL()
	return baseURL + constants.APIBasePath
}

// getWebSocketURL returns the WebSocket URL for the application
func getWebSocketURL() string {
	baseURL := getBaseURL()

	// Convert HTTP to WS and HTTPS to WSS
	if strings.HasPrefix(baseURL, constants.HTTPSProtocol) {
		return constants.WSSProtocol + strings.TrimPrefix(baseURL, constants.HTTPSProtocol) + constants.WebSocketPath
	}
	return constants.WSProtocol + strings.TrimPrefix(baseURL, constants.HTTPProtocol) + constants.WebSocketPath
}

func main() {
	app := pocketbase.New()

	// Start the cleanup scheduler for old URLs
	startCleanupScheduler(app)

	// Start the WebSocket hub
	go hub.Run()

	// Add custom routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Initialize database table
		if err := initDatabase(app); err != nil {
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
			e.Response.Header().Set(constants.Expires, "0")

			return e.JSON(http.StatusOK, map[string]interface{}{
				constants.JSONMessage:       constants.CleanupCompletedMsg,
				constants.JSONRowsDeleted:   rowsAffected,
				constants.JSONCutoffTime:    cutoffTime.Format(constants.TimeFormat),
				constants.JSONCleanupReason: constants.ManualCleanup,
			})
		})

		// Test endpoint to verify click counting is working
		e.Router.GET("/api/test-click/{shortCode}", func(e *core.RequestEvent) error {
			shortCode := e.Request.PathValue("shortCode")

			if shortCode == "" {
				return e.JSON(http.StatusNotFound, map[string]string{
					constants.JSONError: constants.ShortCodeRequired,
				})
			}

			// Get current click count
			urlData, err := getURLFromDB(app, shortCode)
			if err != nil || urlData == nil {
				return e.JSON(http.StatusNotFound, map[string]string{
					constants.JSONError: constants.URLNotFound,
				})
			}

			currentClicks := urlData["clicks"].(int)

			// Increment click count
			if err := incrementClickCount(app, shortCode); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]string{
					constants.JSONError: constants.IncrementError,
				})
			}

			// Get updated click count
			updatedData, err := getURLFromDB(app, shortCode)
			if err != nil || updatedData == nil {
				return e.JSON(http.StatusInternalServerError, map[string]string{
					constants.JSONError: constants.DatabaseError,
				})
			}

			newClicks := updatedData["clicks"].(int)

			// Set aggressive cache-busting headers
			e.Response.Header().Set(constants.CacheControl, constants.NoCacheMaxAge)
			e.Response.Header().Set(constants.Pragma, constants.NoCache)
			e.Response.Header().Set(constants.Expires, "0")
			e.Response.Header().Set(constants.LastModified, time.Now().Format(http.TimeFormat))
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
		e.Router.GET("/api/stats/{shortCode}", func(e *core.RequestEvent) error {
			shortCode := e.Request.PathValue("shortCode")

			if shortCode == "" {
				return e.JSON(http.StatusNotFound, map[string]string{
					constants.JSONError: constants.ShortCodeRequired,
				})
			}

			// Try database first
			urlData, err := getURLFromDB(app, shortCode)
			if err == nil && urlData != nil {
				// Set cache-busting headers
				e.Response.Header().Set(constants.CacheControl, constants.NoCache)
				e.Response.Header().Set(constants.Pragma, constants.NoCache)
				e.Response.Header().Set(constants.Expires, "0")

				return e.JSON(http.StatusOK, map[string]interface{}{
					constants.JSONShortCode:   shortCode,
					constants.JSONOriginalURL: urlData["original_url"].(string),
					constants.JSONClicks:      urlData["clicks"].(int),
					constants.JSONCreated:     urlData["created"].(time.Time).Format(constants.TimeFormat),
				})
			}

			return e.JSON(http.StatusNotFound, map[string]string{
				constants.JSONError: constants.URLNotFound,
			})
		})

		// GET endpoint to get recent URLs
		e.Router.GET("/api/recent", func(e *core.RequestEvent) error {
			// Fetch last 5 URLs from database
			recentURLs, err := getRecentURLsFromDB(app, 5)
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]string{
					constants.JSONError: constants.FetchURLError,
				})
			}

			// Set cache-busting headers
			e.Response.Header().Set(constants.CacheControl, constants.NoCache)
			e.Response.Header().Set(constants.Pragma, constants.NoCache)
			e.Response.Header().Set(constants.Expires, "0")

			return e.JSON(http.StatusOK, map[string]interface{}{
				"urls": recentURLs,
			})
		})

		// POST endpoint to create short URL
		e.Router.POST("/api/shorten", func(e *core.RequestEvent) error {
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
			} else if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
				return e.JSON(http.StatusBadRequest, map[string]string{
					constants.JSONError: constants.HTTPSOnly,
				})
			}

			// Generate short code
			shortCode := generateShortCode()

			// Store URL in PocketBase database
			err = storeURLInDB(app, shortCode, data.URL)
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]string{
					constants.JSONError: constants.StoreURLError,
				})
			}

			// NOTE: To make data visible in PocketBase dashboard, you need to:
			// 1. Create a proper "urls" collection in PocketBase admin UI
			// 2. Replace the in-memory storage with actual database operations
			// 3. Use PocketBase's collection and record APIs

			baseURL := getBaseURL()
			return e.JSON(http.StatusCreated, map[string]interface{}{
				constants.JSONOriginalURL: data.URL,
				constants.JSONShortCode:   shortCode,
				constants.JSONShortURL:    baseURL + "/" + shortCode,
			})
		})

		// WebSocket endpoint for real-time stats updates
		e.Router.GET("/ws", func(e *core.RequestEvent) error {
			handleWebSocket(e.Response, e.Request)
			return nil
		})

		// GET endpoint to redirect short URLs
		e.Router.GET("/{shortCode}", func(e *core.RequestEvent) error {
			shortCode := e.Request.PathValue("shortCode")

			if shortCode == "" {
				return e.JSON(http.StatusNotFound, map[string]string{
					"error": "Short code not provided",
				})
			}

			var originalURL string
			var found bool

			// Try database first
			urlData, err := getURLFromDB(app, shortCode)
			if err == nil && urlData != nil {
				originalURL = urlData["original_url"].(string)
				found = true

				// Increment click count in database
				if err := incrementClickCount(app, shortCode); err != nil {
					// Log error but don't fail the request
					log.Printf("Failed to increment click count in database: %v", err)
				} else {
					log.Printf("Incremented click count for short code: %s", shortCode)
					// Broadcast the click update via WebSocket
					if urlData["clicks"] != nil {
						clicks := urlData["clicks"].(int) + 1
						hub.BroadcastStatsUpdate(shortCode, clicks)
					}
				}
			}

			if !found {
				return e.JSON(http.StatusNotFound, map[string]string{
					"error": "URL not found",
				})
			}

			// Validate URL before redirecting
			if originalURL == "" {
				return e.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Invalid URL stored",
				})
			}

			// Set cache-busting headers to prevent caching
			e.Response.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			e.Response.Header().Set("Pragma", "no-cache")
			e.Response.Header().Set("Expires", "0")

			// Redirect to original URL
			return e.Redirect(http.StatusMovedPermanently, originalURL)
		})

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// storeURLInDB stores URL data in PocketBase database
func storeURLInDB(app *pocketbase.PocketBase, shortCode, originalURL string) error {
	// Use the table structure created by initDatabase
	_, err := app.DB().NewQuery(`
		INSERT OR REPLACE INTO urls (short_code, original_url, clicks, created, updated)
		VALUES ({:shortCode}, {:originalURL}, {:clicks}, {:createdAt}, {:updatedAt})
	`).Bind(map[string]interface{}{
		"shortCode":   shortCode,
		"originalURL": originalURL,
		"clicks":      0,
		"createdAt":   time.Now(),
		"updatedAt":   time.Now(),
	}).Execute()

	return err
}

// getURLFromDB retrieves URL data from PocketBase database
func getURLFromDB(app *pocketbase.PocketBase, shortCode string) (map[string]interface{}, error) {
	var id string
	var originalURL string
	var clicks int
	var created string

	err := app.DB().NewQuery(`
		SELECT id, original_url, clicks, created
		FROM urls
		WHERE short_code = {:shortCode}
	`).Bind(map[string]interface{}{
		"shortCode": shortCode,
	}).Row(&id, &originalURL, &clicks, &created)

	if err != nil {
		return nil, err
	}

	// Parse the created timestamp
	createdAt, err := time.Parse("2006-01-02 15:04:05 -0700 MST", created)
	if err != nil {
		// If parsing fails, use current time
		createdAt = time.Now()
	}

	return map[string]interface{}{
		"id":           id,
		"original_url": originalURL,
		"short_code":   shortCode,
		"clicks":       clicks,
		"created":      createdAt,
	}, nil
}

// incrementClickCount increments the click count for a short code
func incrementClickCount(app *pocketbase.PocketBase, shortCode string) error {
	_, err := app.DB().NewQuery(`
		UPDATE urls
		SET clicks = clicks + 1, updated = CURRENT_TIMESTAMP
		WHERE short_code = {:shortCode}
	`).Bind(map[string]interface{}{
		"shortCode": shortCode,
	}).Execute()

	return err
}

// getRecentURLsFromDB retrieves the last N URLs from the database
func getRecentURLsFromDB(app *pocketbase.PocketBase, limit int) ([]map[string]interface{}, error) {
	rows := []struct {
		ID          string `db:"id" json:"id"`
		ShortCode   string `db:"short_code" json:"short_code"`
		OriginalURL string `db:"original_url" json:"original_url"`
		Clicks      int    `db:"clicks" json:"clicks"`
		Created     string `db:"created" json:"created"`
	}{}

	err := app.DB().NewQuery(`
		SELECT id, short_code, original_url, clicks, created
		FROM urls
		ORDER BY created DESC
		LIMIT {:limit}
	`).Bind(map[string]interface{}{
		"limit": limit,
	}).All(&rows)

	if err != nil {
		log.Printf("Database error in getRecentURLsFromDB: %v", err)
		return nil, err
	}

	// Convert to the format expected by the frontend
	var urls []map[string]interface{}
	baseURL := getBaseURL()
	for _, row := range rows {
		// Parse the created timestamp
		createdAt, err := time.Parse("2006-01-02 15:04:05 -0700 MST", row.Created)
		if err != nil {
			log.Printf("Error parsing created timestamp %s: %v", row.Created, err)
			createdAt = time.Now()
		}

		urlData := map[string]interface{}{
			"id":           row.ID,
			"short_code":   row.ShortCode,
			"original_url": row.OriginalURL,
			"clicks":       row.Clicks,
			"created":      createdAt.Format("2006-01-02 15:04:05"),
			"short_url":    baseURL + "/" + row.ShortCode,
		}
		urls = append(urls, urlData)
	}

	return urls, nil
}

// Simple in-memory storage for demo purposes (keeping as fallback)
var urlStore = make(map[string]map[string]interface{})

// storeURL stores URL data in memory
func storeURL(shortCode string, data map[string]interface{}) {
	urlStore[shortCode] = data
}

// getURL retrieves URL data from memory
func getURL(shortCode string) map[string]interface{} {
	return urlStore[shortCode]
}

// generateShortCode generates a random 6-character base62 encoded string
func generateShortCode() string {
	const length = 6

	// Generate 4 random bytes (32 bits) which gives us enough entropy
	// 4 bytes = 32 bits = 2^32 possibilities
	// When base62 encoded, this gives us a 6-character string
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based code if random fails
		return fmt.Sprintf("%x", time.Now().UnixNano())[:length]
	}

	// Convert bytes to uint32
	num := uint32(bytes[0])<<24 + uint32(bytes[1])<<16 + uint32(bytes[2])<<8 + uint32(bytes[3])

	// Base62 encode the number
	encoded := encodeBase62(num)

	// Ensure we have at least 6 characters by padding with '0' if necessary
	if len(encoded) < length {
		// Pad with leading zeros
		padding := length - len(encoded)
		padded := strings.Repeat("0", padding) + encoded
		return padded
	}

	// Return first 6 characters if longer
	return encoded[:length]
}

// encodeBase62 encodes a number to base62 string
func encodeBase62(n uint32) string {
	const charset = constants.Base62Charset
	const base = uint32(len(charset))

	if n == 0 {
		return string(charset[0])
	}

	var result []byte
	for n > 0 {
		remainder := n % base
		result = append(result, charset[remainder])
		n = n / base
	}

	// Reverse the result since we built it backwards
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// generatePocketBaseID generates a PocketBase-style ID
func generatePocketBaseID() string {
	bytes := make([]byte, 7)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return constants.PBIDPrefix + fmt.Sprintf("%07x", time.Now().UnixNano()%0x10000000)
	}

	// Convert to hex and lowercase
	hexStr := fmt.Sprintf("%014x", bytes)
	return constants.PBIDPrefix + strings.ToLower(hexStr)
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// Client represents a WebSocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mutex      sync.Mutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected: %s", client.ID)

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("Client disconnected: %s", client.ID)

		case message := <-h.Broadcast:
			h.mutex.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// BroadcastStatsUpdate broadcasts a stats update to all connected clients
func (h *Hub) BroadcastStatsUpdate(shortCode string, clicks int) {
	update := map[string]interface{}{
		constants.WSMessageType: constants.WSStatsUpdate,
		constants.WSShortCode:   shortCode,
		constants.WSClicks:      clicks,
		constants.WSTimestamp:   time.Now().Unix(),
	}

	message, err := json.Marshal(update)
	if err != nil {
		log.Printf(constants.WSMarshalError+" %v", err)
		return
	}

	select {
	case h.Broadcast <- message:
	default:
		log.Println(constants.WSBroadcastFull)
	}
}

// Global hub instance
var hub = NewHub()

// WebSocket message types
type WSMessage struct {
	Type      string      `json:"type"`
	ShortCode string      `json:"short_code,omitempty"`
	Clicks    int         `json:"clicks,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(constants.WSUpgradeError+" %v", err)
		return
	}

	clientID := generatePocketBaseID()
	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	hub.Register <- client

	// Start goroutine to write messages to WebSocket
	go client.writePump()

	// Start goroutine to read messages from WebSocket
	go client.readPump()
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
		hub.Unregister <- c
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf(constants.WSWriteError+" %v", err)
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Conn.Close()
		hub.Unregister <- c
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf(constants.WSReadError+" %v", err)
			}
			break
		}
		// We don't need to handle incoming messages for now
		// This is just to keep the connection alive
	}
}
