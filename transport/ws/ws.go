package ws

import (
	"net/http"

	"github.com/kweusuf/pocketbase-demo/pkg/service"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"

	"github.com/gorilla/websocket"
	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/generator"
	"github.com/pocketbase/pocketbase/core"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// InitializeWebSocket initializes the WebSocket hub
func InitializeWebSocket() {
	// Start the global hub if it's not already running
	// The GlobalWSService is already initialized as a global variable
	// We just need to make sure it's running
	service.GlobalWSService.Start()
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Ensure WebSocket hub is initialized
	InitializeWebSocket()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Info(constants.WSUpgradeError+" %v", err)
		return
	}

	clientID := generator.GeneratePocketBaseID()
	client := &service.WSClient{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, constants.LargeBatchSize),
	}

	service.GlobalWSService.RegisterClient(client)

	// Start goroutine to write messages to WebSocket
	go client.WritePump(service.GlobalWSService.GetHub())

	// Start goroutine to read messages from WebSocket
	go client.ReadPump(service.GlobalWSService.GetHub())
}

// RegisterWSRoutes registers all WebSocket endpoints
func RegisterWSRoutes(e *core.ServeEvent) error {
	// WebSocket endpoint for real-time stats updates
	e.Router.GET("/ws", func(e *core.RequestEvent) error {
		HandleWebSocket(e.Response, e.Request)
		return nil
	})

	return nil
}
