package httproutes

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterStatic registers all HTTP endpoints for the URL shortener service
func RegisterStatic(app *pocketbase.PocketBase, e *core.ServeEvent) error {
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
	return nil
}
