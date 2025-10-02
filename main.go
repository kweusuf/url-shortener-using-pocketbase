package main

import (
	"log"

	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/db"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/monitoring"
	wsutil "github.com/kweusuf/pocketbase-demo/pkg/utils/ws"
	httproutes "github.com/kweusuf/pocketbase-demo/transport/http"
	wstransport "github.com/kweusuf/pocketbase-demo/transport/ws"
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

		// Register HTTP routes
		if err := httproutes.RegisterHTTPRoutes(app, e); err != nil {
			return err
		}

		// Register WebSocket routes
		if err := wstransport.RegisterWSRoutes(e); err != nil {
			return err
		}

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
