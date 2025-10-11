package db

import (
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/models"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/url"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// InitDatabase creates the database table for storing URLs
func InitDatabase(app *pocketbase.PocketBase) error {
	log.Info(constants.InitializingDB)

	// Initialize users collection first if it doesn't exist
	if err := initUsersCollection(app); err != nil {
		log.Info("Error initializing users collection: %v", err)
		return err
	}

	// Check if table exists
	exists, err := checkTableExists(app)
	if err != nil {
		log.Info("Error checking if table exists: %v", err)
		return err
	}

	if !exists {
		log.Info(constants.IndexCreated)
		if err := createURLsTable(app); err != nil {
			log.Info("Error creating URLs table: %v", err)
			return err
		}
		log.Info(constants.TableCreated)
	} else {
		log.Info(constants.TableExists)
		// Validate table schema
		valid, err := validateTableSchema(app)
		if err != nil {
			log.Info("Error validating table schema: %v", err)
			return err
		}

		if !valid {
			log.Info(constants.TableSchemaInvalid)
			if err := dropAndRecreateTable(app); err != nil {
				log.Info("Error recreating table: %v", err)
				return err
			}
			log.Info(constants.TableRecreated)
		} else {
			log.Info(constants.TableSchemaValid)
		}
	}

	// Create index if it doesn't exist
	if err := createIndexIfNotExists(app); err != nil {
		log.Info("Error creating index: %v", err)
		return err
	}

	log.Info(constants.DatabaseInitialized)
	log.Info(constants.DashboardNote)
	log.Info(constants.DashboardStep1)
	log.Info(constants.DashboardStep2)
	log.Info(constants.DashboardStep3)
	log.Info(constants.DashboardStep4)
	return nil
}

// initUsersCollection initializes the users collection for authentication
func initUsersCollection(app *pocketbase.PocketBase) error {
	collectionName := "users"

	// Check if users collection already exists
	_, err := app.FindCollectionByNameOrId(collectionName)
	if err == nil {
		log.Info("Users collection already exists")
		return nil
	}

	// Create users collection as an auth collection (for proper PocketBase auth)
	collection := core.NewAuthCollection(collectionName)

	// Add additional verified field
	verifiedField := &core.BoolField{
		Name: "verified",
	}
	collection.Fields.Add(verifiedField)

	// Save the collection
	err = app.Save(collection)
	if err != nil {
		log.Info("Error creating users collection: %v", err)
		return err
	}

	log.Info("Successfully created auth collection 'users'")
	return nil
}

// checkTableExists checks if the urls table exists
func checkTableExists(app *pocketbase.PocketBase) (bool, error) {
	return app.HasTable(constants.TableName), nil
}

// validateTableSchema checks if the table has the correct structure
func validateTableSchema(app *pocketbase.PocketBase) (bool, error) {
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return false, err
	}

	// Check for required fields
	requiredFields := []string{
		constants.ColumnShortCode,
		constants.ColumnOriginalURL,
		constants.ColumnUserID,
		constants.ColumnClicks,
		constants.ColumnCreated,
		constants.ColumnUpdated,
	}

	fieldMap := make(map[string]bool)
	for _, field := range collection.Fields {
		fieldMap[field.GetName()] = true
	}

	for _, field := range requiredFields {
		if !fieldMap[field] {
			log.Info("Missing field: %s", field)
			return false, nil
		}
	}

	// Schema is valid
	return true, nil
}

// createURLsTable creates the URLs table with correct structure
func createURLsTable(app *pocketbase.PocketBase) error {
	// Check if collection already exists
	existingCollection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err == nil && existingCollection != nil {
		log.Info("Collection 'urls' already exists, skipping creation")
		return nil
	}

	// Create new collection if it doesn't exist
	collection := core.NewBaseCollection(constants.TableName)

	// Add fields with proper initialization
	shortCodeField := &core.TextField{
		Name:     constants.ColumnShortCode,
		Required: true,
	}
	collection.Fields.Add(shortCodeField)

	originalURLField := &core.TextField{
		Name:     constants.ColumnOriginalURL,
		Required: true,
	}
	collection.Fields.Add(originalURLField)

	userIDField := &core.TextField{
		Name:     constants.ColumnUserID,
		Required: false, // Allow empty user ID for anonymous URLs
	}
	collection.Fields.Add(userIDField)

	clicksField := &core.NumberField{
		Name:     constants.ColumnClicks,
		Required: false, // Allow 0 value for initial clicks
	}
	collection.Fields.Add(clicksField)

	createdField := &core.AutodateField{
		Name:     constants.ColumnCreated,
		OnCreate: true,
	}
	collection.Fields.Add(createdField)

	updatedField := &core.AutodateField{
		Name:     constants.ColumnUpdated,
		OnCreate: true,
		OnUpdate: true,
	}
	collection.Fields.Add(updatedField)

	// Add index for short_code field
	collection.AddIndex(constants.IndexNameShortCode, true, constants.ColumnShortCode, "")

	err = app.Save(collection)
	if err != nil {
		log.Info("Error creating collection: %v", err)
		return err
	}

	log.Info("Successfully created 'urls' collection")
	return nil
}

// dropAndRecreateTable drops and recreates the table
func dropAndRecreateTable(app *pocketbase.PocketBase) error {
	// Find the existing collection
	existingCollection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		log.Info("Collection not found, creating new one: %v", err)
		return createURLsTable(app)
	}

	// Delete the existing collection
	if err := app.Delete(existingCollection); err != nil {
		log.Info("Error deleting existing collection: %v", err)
		return err
	}

	log.Info("Successfully deleted existing 'urls' collection")

	// Create new collection
	return createURLsTable(app)
}

// createIndexIfNotExists creates the index if it doesn't exist
func createIndexIfNotExists(app *pocketbase.PocketBase) error {
	// Get the collection to check its indexes programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return err
	}

	// Check if index already exists
	indexExists := false
	for _, index := range collection.Indexes {
		if index == constants.IndexNameShortCode {
			indexExists = true
			break
		}
	}

	if !indexExists {
		log.Info("Creating index on short_code...")
		collection.AddIndex(constants.IndexNameShortCode, true, constants.ColumnShortCode, "")
		collection.AddIndex(constants.IndexNameUserID, false, constants.ColumnUserID, "")
		err = app.Save(collection)
		return err
	}

	return nil
}

// CleanupOldURLs removes URLs that haven't been accessed for more than 1 hour from the database
func CleanupOldURLs(app *pocketbase.PocketBase) error {
	// Get the collection to work with records programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return err
	}

	// Find all records and filter manually (simplified for now)
	records, err := app.FindRecordsByFilter(
		collection,
		constants.NoFilter,
		constants.NoSort,
		constants.LargeBatchSize,
		constants.NoOffset,
		map[string]interface{}{},
	)
	if err != nil {
		return err
	}

	// Calculate cutoff time: now minus TTL duration
	cutoffTime := time.Now().UTC().Add(-constants.URLExpirationHours * time.Hour)
	deletedCount := 0

	for _, record := range records {
		// Parse the created timestamp and compare with cutoff (created < cutoff means expired)
		createdStr := record.GetString(constants.ColumnCreated)
		createdTime, err := time.Parse(constants.TimeFormat, createdStr)
		if err != nil {
			log.Info("Error parsing created for record %s: %v", record.Id, err)
			continue // Skip records with invalid timestamps
		}

		if createdTime.Before(cutoffTime) {
			if err := app.Delete(record); err != nil {
				log.Info("Error deleting expired URL record %s: %v", record.Id, err)
				continue
			}
			deletedCount++
		}
	}

	// Log how many rows were deleted
	if deletedCount > 0 {
		log.Info("Cleaned up %d expired URLs", deletedCount)
	}

	return nil
}

// StoreURLInDB stores URL data in PocketBase database
func StoreURLInDB(app *pocketbase.PocketBase, shortCode, originalURL, userID string) error {
	log.Info("StoreURLInDB: Attempting to store URL with shortCode=%s, userID=%s", shortCode, userID)

	// Get the URLs collection
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		log.Info("StoreURLInDB: Failed to get collection '%s': %v", constants.TableName, err)
		return err
	}
	log.Info("StoreURLInDB: Got collection: %v", collection.Name)

	// Create record (PocketBase auto-generates ID)
	record := core.NewRecord(collection)

	// Set the fields
	record.Set(constants.ColumnShortCode, shortCode)
	record.Set(constants.ColumnOriginalURL, originalURL)
	record.Set(constants.ColumnUserID, userID)

	// Don't set clicks at all - let it use the database default
	// Don't set id - it will be auto-generated

	log.Info("StoreURLInDB: Created record with fields: shortCode=%s, originalURL=%s, userID=%s",
		record.GetString(constants.ColumnShortCode),
		record.GetString(constants.ColumnOriginalURL),
		record.GetString(constants.ColumnUserID))

	// Save without the id field set
	if err := app.Save(record); err != nil {
		log.Info("StoreURLInDB: Failed to save record: %v", err)
		return err
	}

	log.Info("StoreURLInDB: Successfully created URL with AutoID: %s", record.Id)
	return nil
}

// GetURLFromDB retrieves URL data from PocketBase database
func GetURLFromDB(app *pocketbase.PocketBase, shortCode string) (map[string]interface{}, error) {
	// Get the collection to work with records programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return nil, err
	}

	// Find the record using PocketBase's record API
	record, err := app.FindFirstRecordByFilter(
		collection,
		constants.ShortCodeFilter,
		map[string]interface{}{
			constants.ParamShortCode: shortCode,
		},
	)
	if err != nil {
		return nil, err
	}

	// Parse the created timestamp
	createdAt, err := time.Parse(constants.TimeFormat, record.GetString(constants.ColumnCreated))
	if err != nil {
		// If parsing fails, use current time
		createdAt = time.Now().UTC()
	} else {
		createdAt, _ = time.Parse(constants.TimeFormat, record.GetString(constants.ColumnCreated))
	}

	return map[string]interface{}{
		constants.ResponseID:          record.Id,
		constants.ResponseOriginalURL: record.GetString(constants.ColumnOriginalURL),
		constants.ResponseShortCode:   shortCode,
		constants.ResponseClicks:      record.GetInt(constants.ColumnClicks),
		constants.ResponseCreated:     createdAt,
	}, nil
}

// IncrementClickCount increments the click count for a short code
func IncrementClickCount(app *pocketbase.PocketBase, shortCode string) error {
	// Get the collection to work with records programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return err
	}

	// Find the record using PocketBase's record API
	record, err := app.FindFirstRecordByFilter(
		collection,
		constants.ShortCodeFilter,
		map[string]interface{}{
			constants.ParamShortCode: shortCode,
		},
	)
	if err != nil {
		return err
	}

	// Increment the click count and update the timestamp programmatically
	record.Set(constants.ColumnClicks, record.GetInt(constants.ColumnClicks)+1)
	record.Set(constants.ColumnUpdated, time.Now().UTC().Format(constants.TimeFormat))

	// Save the updated record
	err = app.Save(record)
	return err
}

// GetRecentURLsFromDB retrieves the last N URLs from the database for a specific user
func GetRecentURLsFromDB(app *pocketbase.PocketBase, limit int, userID string) ([]map[string]interface{}, error) {
	// Get the collection to work with records programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return nil, err
	}

	// If no userID provided, return empty results for security
	if userID == "" {
		return []map[string]interface{}{}, nil
	}

	// Find recent records for the user using PocketBase's record API with sorting and limit
	records, err := app.FindRecordsByFilter(
		collection,
		constants.ColumnUserID+" = {:userId}",
		constants.CreatedSortDesc,
		limit,
		constants.NoOffset,
		map[string]interface{}{constants.ParamUserID: userID},
	)
	if err != nil {
		log.Info("Database error in GetRecentURLsFromDB: %v", err)
		return nil, err
	}

	// Convert to the format expected by the frontend
	var urls []map[string]interface{}
	baseURL := url.GetBaseURL()
	for _, record := range records {
		// Parse the created timestamp
		createdAt, err := time.Parse(constants.TimeFormat, record.GetString(constants.ColumnCreated))
		if err != nil {
			log.Info("Error parsing created timestamp %s: %v", record.GetString(constants.ColumnCreated), err)
			createdAt = time.Now().UTC()
		}

		urlData := map[string]interface{}{
			constants.ResponseID:          record.Id,
			constants.ResponseShortCode:   record.GetString(constants.ColumnShortCode),
			constants.ResponseOriginalURL: record.GetString(constants.ColumnOriginalURL),
			constants.ResponseClicks:      record.GetInt(constants.ColumnClicks),
			constants.ResponseCreated:     createdAt.Format(constants.TimeFormat),
			constants.ResponseShortURL:    baseURL + "/" + record.GetString(constants.ColumnShortCode),
		}
		urls = append(urls, urlData)
	}

	return urls, nil
}

// GetDatabaseInfo returns database information for health checks
func GetDatabaseInfo(app *pocketbase.PocketBase) models.DBInfo {
	info := models.DBInfo{
		Type:             "sqlite",
		ConnectionStatus: "unknown",
	}

	if app.DB() == nil {
		info.ConnectionStatus = "disconnected"
		return info
	}

	start := time.Now()
	_, err := app.DB().NewQuery("SELECT 1").Execute()
	responseTime := time.Since(start)

	if err != nil {
		info.ConnectionStatus = "error"
		info.ResponseTime = responseTime
	} else {
		info.ConnectionStatus = "connected"
		info.ResponseTime = responseTime
	}

	return info
}

// GetUserURLsFromDB retrieves all URLs for a specific user
func GetUserURLsFromDB(app *pocketbase.PocketBase, userID string) ([]map[string]interface{}, error) {
	// Get the collection to work with records programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return nil, err
	}

	// Find all records for the user using PocketBase's record API
	records, err := app.FindRecordsByFilter(
		collection,
		constants.ColumnUserID+" = {:userID}",
		constants.CreatedSortDesc,
		constants.LargeBatchSize,
		constants.NoOffset,
		map[string]interface{}{constants.ParamUserID: userID},
	)
	if err != nil {
		log.Info("Database error in GetUserURLsFromDB: %v", err)
		return nil, err
	}

	// Convert to the format expected by the frontend
	var urls []map[string]interface{}
	baseURL := url.GetBaseURL()
	for _, record := range records {
		// Parse the created timestamp
		createdAt, err := time.Parse(constants.TimeFormat, record.GetString(constants.ColumnCreated))
		if err != nil {
			log.Info("Error parsing created timestamp %s: %v", record.GetString(constants.ColumnCreated), err)
			createdAt = time.Now().UTC()
		}

		urlData := map[string]interface{}{
			constants.ResponseID:          record.Id,
			constants.ResponseShortCode:   record.GetString(constants.ColumnShortCode),
			constants.ResponseOriginalURL: record.GetString(constants.ColumnOriginalURL),
			constants.ResponseClicks:      record.GetInt(constants.ColumnClicks),
			constants.ResponseCreated:     createdAt.Format(constants.TimeFormat),
			constants.ResponseShortURL:    baseURL + "/" + record.GetString(constants.ColumnShortCode),
		}
		urls = append(urls, urlData)
	}

	return urls, nil
}
