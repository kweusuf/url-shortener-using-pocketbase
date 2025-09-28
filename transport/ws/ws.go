package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/generator"
	wsutil "github.com/kweusuf/pocketbase-demo/pkg/utils/ws"
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
	// The GlobalHub is already initialized as a global variable
	// We just need to make sure it's running
	go func() {
		wsutil.GlobalHub.Run()
	}()
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Ensure WebSocket hub is initialized
	InitializeWebSocket()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(constants.WSUpgradeError+" %v", err)
		return
	}

	clientID := generator.GeneratePocketBaseID()
	client := &wsutil.Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, constants.LargeBatchSize),
	}

	wsutil.GlobalHub.Register <- client

	// Start goroutine to write messages to WebSocket
	go client.WritePump(wsutil.GlobalHub)

	// Start goroutine to read messages from WebSocket
	go client.ReadPump(wsutil.GlobalHub)
}
