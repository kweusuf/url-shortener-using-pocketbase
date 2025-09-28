package db

import (
	"log"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/pocketbase/pocketbase"
)

// StartCleanupScheduler starts a goroutine that cleans up old URLs every 10 minutes
func StartCleanupScheduler(app *pocketbase.PocketBase) {
	// Schedule cleanup every 10 minutes (don't run immediately to avoid DB issues)
	ticker := time.NewTicker(constants.CleanupInterval * time.Minute)
	go func() {
		for range ticker.C {
			if err := CleanupOldURLs(app); err != nil {
				log.Printf(constants.CleanupError+" %v", err)
			}
		}
	}()

	log.Println(constants.CleanupScheduled)
}
