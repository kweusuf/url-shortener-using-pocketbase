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
	return app.HasTable(constants.TableName), nil
}

// validateTableSchema checks if the table has the correct structure
func validateTableSchema(app *pocketbase.PocketBase) (bool, error) {
	requiredColumns := []string{constants.ColumnID, constants.ColumnShortCode, constants.ColumnOriginalURL, constants.ColumnClicks, constants.ColumnCreated, constants.ColumnUpdated}

	// Get the collection to check its fields programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return false, err
	}

	// Check if all required fields exist in the collection
	for _, requiredCol := range requiredColumns {
		found := false
		for _, field := range collection.Fields {
			if field.GetName() == requiredCol {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

// createURLsTable creates the URLs table with correct structure
func createURLsTable(app *pocketbase.PocketBase) error {
	collection := core.NewBaseCollection(constants.TableName)
	// add text field
	collection.Fields.Add(
		&core.TextField{
			Name:       constants.ColumnID,
			Required:   true,
			PrimaryKey: true,
		},
		&core.TextField{
			Name:     constants.ColumnShortCode,
			Required: true,
		},
		&core.TextField{
			Name:     constants.ColumnOriginalURL,
			Required: true,
		},
		&core.NumberField{
			Name:     constants.ColumnClicks,
			Required: true,
		},
		&core.AutodateField{
			Name:     constants.ColumnCreated,
			OnCreate: true,
		},
		&core.AutodateField{
			Name:     constants.ColumnUpdated,
			OnCreate: true,
		},
	)
	collection.AddIndex(constants.IndexName, true, constants.ColumnShortCode, "")

	err := app.Save(collection)

	return err
}

// dropAndRecreateTable drops and recreates the table
func dropAndRecreateTable(app *pocketbase.PocketBase) error {
	// Drop the table
	_, err := app.DB().NewQuery(`DROP TABLE IF EXISTS ` + constants.TableName).Execute()
	if err != nil {
		return err
	}

	// Recreate it
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
	// Calculate the cutoff time (1 hour ago)
	cutoffTime := time.Now().UTC().Add(-1 * time.Hour)

	// Get the collection to work with records programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return err
	}

	// Find old records using PocketBase's record API
	records, err := app.FindRecordsByFilter(
		collection,
		constants.UpdatedFilter,
		constants.CreatedSortDesc,
		constants.RecordBatchSize,
		constants.NoOffset,
		map[string]interface{}{
			constants.ParamCutoffTime: cutoffTime.Format(constants.TimeFormat),
		},
	)
	if err != nil {
		return err
	}

	// Delete old records programmatically
	deletedCount := 0
	for _, record := range records {
		if err := app.Delete(record); err != nil {
			log.Printf("Error deleting old URL record %s: %v", record.Id, err)
			continue
		}
		deletedCount++
	}

	// Log how many rows were deleted
	if deletedCount > 0 {
		log.Printf("Cleaned up %d old URLs (not accessed for more than 1 hour)", deletedCount)
	}

	return nil
}

// StoreURLInDB stores URL data in PocketBase database
func StoreURLInDB(app *pocketbase.PocketBase, shortCode, originalURL string) error {
	// Use the table structure created by initDatabase
	_, err := app.DB().NewQuery(`
		INSERT OR REPLACE INTO ` + constants.TableName + ` (` + constants.ColumnShortCode + `, ` + constants.ColumnOriginalURL + `, ` + constants.ColumnClicks + `, ` + constants.ColumnCreated + `, ` + constants.ColumnUpdated + `)
		VALUES ({:` + constants.ParamShortCode + `}, {:` + constants.ParamOriginalURL + `}, {:` + constants.ParamClicks + `}, {:` + constants.ParamCreatedAt + `}, {:` + constants.ParamUpdatedAt + `})
	`).Bind(map[string]interface{}{
		constants.ParamShortCode:   shortCode,
		constants.ParamOriginalURL: originalURL,
		constants.ParamClicks:      0,
		constants.ParamCreatedAt:   time.Now().UTC().Format(constants.TimeFormat),
		constants.ParamUpdatedAt:   time.Now().UTC().Format(constants.TimeFormat),
	}).Execute()

	return err
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

// GetRecentURLsFromDB retrieves the last N URLs from the database
func GetRecentURLsFromDB(app *pocketbase.PocketBase, limit int) ([]map[string]interface{}, error) {
	// Get the collection to work with records programmatically
	collection, err := app.FindCollectionByNameOrId(constants.TableName)
	if err != nil {
		return nil, err
	}

	// Find recent records using PocketBase's record API with sorting and limit
	records, err := app.FindRecordsByFilter(
		collection,
		constants.NoFilter,
		constants.CreatedSortDesc,
		limit,
		constants.NoOffset,
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
