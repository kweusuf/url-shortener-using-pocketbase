package db

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
)

// StartCleanupScheduler starts a goroutine that cleans up old URLs every 10 minutes
func StartCleanupScheduler(app *pocketbase.PocketBase) {
	// Schedule cleanup every 10 minutes (don't run immediately to avoid DB issues)
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			if err := CleanupOldURLs(app); err != nil {
				log.Printf("Error during scheduled cleanup: %v", err)
			}
		}
	}()

	log.Println("URL cleanup scheduler started (runs every 10 minutes)")
}
