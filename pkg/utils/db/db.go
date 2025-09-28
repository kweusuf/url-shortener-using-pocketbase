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

// CleanupOldURLs removes URLs older than 1 hour from the database
func CleanupOldURLs(app *pocketbase.PocketBase) error {
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

// StoreURLInDB stores URL data in PocketBase database
func StoreURLInDB(app *pocketbase.PocketBase, shortCode, originalURL string) error {
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

// GetURLFromDB retrieves URL data from PocketBase database
func GetURLFromDB(app *pocketbase.PocketBase, shortCode string) (map[string]interface{}, error) {
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

// IncrementClickCount increments the click count for a short code
func IncrementClickCount(app *pocketbase.PocketBase, shortCode string) error {
	_, err := app.DB().NewQuery(`
		UPDATE urls
		SET clicks = clicks + 1, updated = CURRENT_TIMESTAMP
		WHERE short_code = {:shortCode}
	`).Bind(map[string]interface{}{
		"shortCode": shortCode,
	}).Execute()

	return err
}

// GetRecentURLsFromDB retrieves the last N URLs from the database
func GetRecentURLsFromDB(app *pocketbase.PocketBase, limit int) ([]map[string]interface{}, error) {
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
	baseURL := url.GetBaseURL()
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
