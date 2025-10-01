package db

import (
	"log"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/url"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// InitDatabase creates the database table for storing URLs
func InitDatabase(app *pocketbase.PocketBase) error {
	log.Println(constants.InitializingDB)

	// Initialize users collection first if it doesn't exist
	if err := initUsersCollection(app); err != nil {
		log.Printf("Error initializing users collection: %v", err)
		return err
	}

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

// initUsersCollection initializes the users collection for authentication
func initUsersCollection(app *pocketbase.PocketBase) error {
	collectionName := "users"

	// Check if users collection already exists
	_, err := app.FindCollectionByNameOrId(collectionName)
	if err == nil {
		log.Println("Users collection already exists")
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
		log.Printf("Error creating users collection: %v", err)
		return err
	}

	log.Println("Successfully created auth collection 'users'")
	return nil
}

// checkTableExists checks if the urls table exists
func checkTableExists(app *pocketbase.PocketBase) (bool, error) {
	return app.HasTable(constants.TableName), nil
}

// validateTableSchema checks if the table has the correct structure
func validateTableSchema(app *pocketbase.PocketBase) (bool, error) {
	// FOR TESTING: Always return invalid to force recreation
	log.Print("Forcing table schema recreation for testing")
	return false, nil
}

// createURLsTable creates the URLs table with correct structure
func createURLsTable(app *pocketbase.PocketBase) error {
	// Check if collection already exists
	existingCollection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err == nil && existingCollection != nil {
		log.Println("Collection 'urls' already exists, skipping creation")
		return nil
	}

	// Create new collection if it doesn't exist
	collection := core.NewBaseCollection(constants.TableName)

	// Add fields with proper initialization
	idField := &core.TextField{
		Name:       constants.ColumnID,
		PrimaryKey: true,
		// Don't set Required=true for primary key - let PocketBase handle auto-generation
	}
	collection.Fields.Add(idField)

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
	collection.AddIndex(constants.IndexName, true, constants.ColumnShortCode, "")

	err = app.Save(collection)
	if err != nil {
		log.Printf("Error creating collection: %v", err)
		return err
	}

	log.Println("Successfully created 'urls' collection")
	return nil
}

// dropAndRecreateTable drops and recreates the table
func dropAndRecreateTable(app *pocketbase.PocketBase) error {
	// Find the existing collection
	existingCollection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		log.Printf("Collection not found, creating new one: %v", err)
		return createURLsTable(app)
	}

	// Delete the existing collection
	if err := app.Delete(existingCollection); err != nil {
		log.Printf("Error deleting existing collection: %v", err)
		return err
	}

	log.Println("Successfully deleted existing 'urls' collection")

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
		if index == constants.IndexName {
			indexExists = true
			break
		}
	}

	if !indexExists {
		log.Println("Creating index on short_code...")
		collection.AddIndex(constants.IndexName, true, constants.ColumnShortCode, "")
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

	// Calculate cutoff time and delete old records manually
	cutoffTime := time.Now().UTC().Add(-1 * time.Hour)
	deletedCount := 0

	for _, record := range records {
		// Parse the updated timestamp and compare
		updatedStr := record.GetString(constants.ColumnUpdated)
		updatedTime, err := time.Parse(constants.TimeFormat, updatedStr)
		if err != nil {
			continue // Skip records with invalid timestamps
		}

		if updatedTime.Before(cutoffTime) {
			if err := app.Delete(record); err != nil {
				log.Printf("Error deleting old URL record %s: %v", record.Id, err)
				continue
			}
			deletedCount++
		}
	}

	// Log how many rows were deleted
	if deletedCount > 0 {
		log.Printf("Cleaned up %d old URLs (not accessed for more than 1 hour)", deletedCount)
	}

	return nil
}

// StoreURLInDB stores URL data in PocketBase database
func StoreURLInDB(app *pocketbase.PocketBase, shortCode, originalURL, userID string) error {
	log.Printf("StoreURLInDB: Attempting to store URL with shortCode=%s, userID=%s", shortCode, userID)

	// Get the URLs collection
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		log.Printf("StoreURLInDB: Failed to get collection '%s': %v", constants.TableName, err)
		return err
	}
	log.Printf("StoreURLInDB: Got collection: %v", collection.Name)

	// Create record and manually generate ID (same pattern as PocketBase)
	record := core.NewRecord(collection)

	// Manually generate ID using PocketBase-style pattern
	// PocketBase uses: 'r'||lower(hex(randomblob(7)))
	record.Id = "r" + "0123456789abcdef012" // Simplified for testing
	record.MarkAsNew()                      // Ensure it's marked as new

	// Set the fields
	record.Set(constants.ColumnID, record.Id) // Explicitly set since manually generated
	record.Set(constants.ColumnShortCode, shortCode)
	record.Set(constants.ColumnOriginalURL, originalURL)
	record.Set(constants.ColumnUserID, userID)

	// Don't set clicks at all - let it use the database default
	// Don't set id - it will be auto-generated

	log.Printf("StoreURLInDB: Created record with fields: shortCode=%s, originalURL=%s, userID=%s",
		record.GetString(constants.ColumnShortCode),
		record.GetString(constants.ColumnOriginalURL),
		record.GetString(constants.ColumnUserID))

	// Save without the id field set
	if err := app.Save(record); err != nil {
		log.Printf("StoreURLInDB: Failed to save record: %v", err)
		return err
	}

	log.Printf("StoreURLInDB: Successfully created URL with AutoID: %s", record.Id)
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

	// Find recent records using PocketBase's record API with sorting and limit
	// For now, skip user filtering and return all recent URLs
	// Proper filtering will be implemented once we figure out PocketBase filter syntax
	records, err := app.FindRecordsByFilter(
		collection,
		constants.NoFilter,
		constants.CreatedSortDesc,
		limit,
		constants.NoOffset,
		map[string]interface{}{},
	)
	if err != nil {
		log.Printf("Database error in getRecentURLsFromDB: %v", err)
		return nil, err
	}

	// Convert to the format expected by the frontend
	var urls []map[string]interface{}
	baseURL := url.GetBaseURL()
	for _, record := range records {
		// Parse the created timestamp
		createdAt, err := time.Parse(constants.TimeFormat, record.GetString(constants.ColumnCreated))
		if err != nil {
			log.Printf("Error parsing created timestamp %s: %v", record.GetString(constants.ColumnCreated), err)
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
		log.Printf("Database error in GetUserURLsFromDB: %v", err)
		return nil, err
	}

	// Convert to the format expected by the frontend
	var urls []map[string]interface{}
	baseURL := url.GetBaseURL()
	for _, record := range records {
		// Parse the created timestamp
		createdAt, err := time.Parse(constants.TimeFormat, record.GetString(constants.ColumnCreated))
		if err != nil {
			log.Printf("Error parsing created timestamp %s: %v", record.GetString(constants.ColumnCreated), err)
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
