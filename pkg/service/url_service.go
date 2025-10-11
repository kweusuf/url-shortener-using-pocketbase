package service

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/db"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/generator"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"
	urlutil "github.com/kweusuf/pocketbase-demo/pkg/utils/url"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// URLData represents URL data structure
type URLData struct {
	ShortCode   string
	OriginalURL string
	Clicks      int
	Created     time.Time
}

// ShortenResponse represents the response for URL shortening
type ShortenResponse struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
}

// StatsResponse represents the response for URL stats
type StatsResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Clicks      int    `json:"clicks"`
	Created     string `json:"created"`
}

// TestClickResponse represents the response for test click
type TestClickResponse struct {
	ShortCode  string `json:"short_code"`
	PrevClicks int    `json:"prev_clicks"`
	CurrClicks int    `json:"curr_clicks"`
	ClicksInc  int    `json:"clicks_inc"`
	Test       string `json:"test"`
}

// RecentURLsResponse represents the response for recent URLs
type RecentURLsResponse struct {
	URLs []map[string]interface{} `json:"urls"`
}

// CleanupResponse represents the response for cleanup
type CleanupResponse struct {
	Message       string `json:"message"`
	RowsDeleted   int64  `json:"rows_deleted"`
	CutoffTime    string `json:"cutoff_time"`
	CleanupReason string `json:"cleanup_reason"`
}

// HelloResponse represents the response for hello endpoint
type HelloResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// URLService handles URL shortening business logic
type URLService struct {
	app       *pocketbase.PocketBase
	wsService *WSService
}

// NewURLService creates a new URLService instance
func NewURLService(app *pocketbase.PocketBase, wsService *WSService) *URLService {
	return &URLService{
		app:       app,
		wsService: wsService,
	}
}

// ShortenURL processes URL shortening business logic
func (s *URLService) ShortenURL(originalURL, userID string) (*ShortenResponse, error) {
	// Validate and normalize URL format
	if originalURL == "" {
		return nil, fmt.Errorf(constants.URLRequired)
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return nil, fmt.Errorf(constants.InvalidURLFormat)
	}

	// Add https:// protocol if missing
	if parsedURL.Scheme == "" {
		// Check if it looks like it has a protocol but is missing
		if strings.HasPrefix(originalURL, constants.HTTPProtocol) || strings.HasPrefix(originalURL, constants.HTTPSProtocol) {
			return nil, fmt.Errorf(constants.InvalidURLFormat)
		}
		// Add https:// protocol for bare URLs
		originalURL = constants.DefaultProtocol + originalURL
	} else if parsedURL.Scheme != constants.HTTPScheme && parsedURL.Scheme != constants.HTTPSScheme {
		return nil, fmt.Errorf(constants.HTTPSOnly)
	}

	// Generate short code
	shortCode := generator.GenerateShortCode()
	log.Info("Generated short code: %s for URL: %s", shortCode, originalURL)

	// Store URL in PocketBase database with proper user association
	err = db.StoreURLInDB(s.app, shortCode, originalURL, userID)
	if err != nil {
		log.Info("Failed to store URL with userID %s: %v", userID, err)
		return nil, fmt.Errorf(constants.StoreURLError)
	}

	log.Info("Successfully stored URL: %s -> %s", shortCode, originalURL)
	baseURL := urlutil.GetBaseURL()
	shortURL := baseURL + "/" + shortCode

	return &ShortenResponse{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		ShortURL:    shortURL,
	}, nil
}

// GetURLStats retrieves URL statistics
func (s *URLService) GetURLStats(shortCode string) (*StatsResponse, error) {
	if shortCode == "" {
		return nil, fmt.Errorf(constants.ShortCodeRequired)
	}

	// Try database first
	urlData, err := db.GetURLFromDB(s.app, shortCode)
	if err != nil || urlData == nil {
		return nil, fmt.Errorf(constants.URLNotFound)
	}

	return &StatsResponse{
		ShortCode:   shortCode,
		OriginalURL: urlData[constants.ColumnOriginalURL].(string),
		Clicks:      urlData[constants.ColumnClicks].(int),
		Created:     urlData[constants.ColumnCreated].(time.Time).Format(constants.TimeFormat),
	}, nil
}

// GetRecentURLs retrieves recent URLs for authenticated user
func (s *URLService) GetRecentURLs(e *core.RequestEvent) (*RecentURLsResponse, error) {
	// Get authenticated user
	user, err := auth.GetUserFromRequest(e)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}

	userID := user.ID

	// Fetch last 5 URLs for the authenticated user
	recentURLs, err := db.GetRecentURLsFromDB(s.app, 5, userID)
	if err != nil {
		return nil, fmt.Errorf(constants.FetchURLError)
	}

	return &RecentURLsResponse{
		URLs: recentURLs,
	}, nil
}

// TestClick increments click count and returns before/after values
func (s *URLService) TestClick(shortCode string) (*TestClickResponse, error) {
	if shortCode == "" {
		return nil, fmt.Errorf(constants.ShortCodeRequired)
	}

	// Get current click count
	urlData, err := db.GetURLFromDB(s.app, shortCode)
	if err != nil || urlData == nil {
		return nil, fmt.Errorf(constants.URLNotFound)
	}

	currentClicks := urlData[constants.ColumnClicks].(int)

	// Increment click count
	if err := db.IncrementClickCount(s.app, shortCode); err != nil {
		return nil, fmt.Errorf(constants.IncrementError)
	}

	// Get updated click count
	updatedData, err := db.GetURLFromDB(s.app, shortCode)
	if err != nil || updatedData == nil {
		return nil, fmt.Errorf(constants.DatabaseError)
	}

	newClicks := updatedData[constants.ColumnClicks].(int)

	return &TestClickResponse{
		ShortCode:  shortCode,
		PrevClicks: currentClicks,
		CurrClicks: newClicks,
		ClicksInc:  newClicks - currentClicks,
		Test:       constants.TestMessage,
	}, nil
}

// RedirectURL handles URL redirection business logic
func (s *URLService) RedirectURL(shortCode string) (string, error) {
	if shortCode == "" {
		return "", fmt.Errorf(constants.ErrorShortCodeNotProvided)
	}

	// Try database first
	urlData, err := db.GetURLFromDB(s.app, shortCode)
	if err != nil || urlData == nil {
		return "", fmt.Errorf(constants.ErrorURLNotFound)
	}

	originalURL := urlData[constants.ColumnOriginalURL].(string)

	// Validate URL before redirecting
	if originalURL == "" {
		return "", fmt.Errorf(constants.ErrorInvalidURLStored)
	}

	// Increment click count in database
	if err := db.IncrementClickCount(s.app, shortCode); err != nil {
		// Log error but don't fail the request
		log.Info("Failed to increment click count in database: %v", err)
	} else {
		log.Info("Incremented click count for short code: %s", shortCode)
		// Broadcast the click update via WebSocket
		s.handleClickBroadcast(shortCode, urlData)
	}

	return originalURL, nil
}

// CleanupURLs performs manual cleanup of old URLs
func (s *URLService) CleanupURLs() (*CleanupResponse, error) {
	cutoffTime := time.Now().Add(-1 * time.Hour)

	result, err := s.app.DB().NewQuery(`
		DELETE FROM urls
		WHERE created < {:cutoffTime}
	`).Bind(map[string]interface{}{
		"cutoffTime": cutoffTime,
	}).Execute()

	if err != nil {
		return nil, fmt.Errorf(constants.CleanupErrorMsg)
	}

	rowsAffected, _ := result.RowsAffected()

	return &CleanupResponse{
		Message:       constants.CleanupCompletedMsg,
		RowsDeleted:   rowsAffected,
		CutoffTime:    cutoffTime.Format(constants.TimeFormat),
		CleanupReason: constants.ManualCleanup,
	}, nil
}

// Hello returns a hello message
func (s *URLService) Hello() *HelloResponse {
	return &HelloResponse{
		Message: constants.HelloMessage,
		Status:  constants.SuccessStatus,
	}
}

// handleClickBroadcast broadcasts URL click updates via WebSocket
func (s *URLService) handleClickBroadcast(shortCode string, urlData map[string]interface{}) {
	if urlData[constants.ColumnClicks] != nil {
		clicks := urlData[constants.ColumnClicks].(int) + 1
		s.wsService.BroadcastStatsUpdate(shortCode, clicks)
	}
}
