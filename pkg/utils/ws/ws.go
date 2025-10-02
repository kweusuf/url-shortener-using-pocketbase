package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/utils/log"

	"github.com/gorilla/websocket"
	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

// Client represents a WebSocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.Mutex
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.Mutex.Unlock()
			log.Info("Client connected: %s", client.ID)

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.Mutex.Unlock()
			log.Info("Client disconnected: %s", client.ID)

		case message := <-h.Broadcast:
			h.Mutex.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.Mutex.Unlock()
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump(hub *Hub) {
	defer func() {
		c.Conn.Close()
		hub.Unregister <- c
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Info(constants.WSWriteError+" %v", err)
			return
		}
	}

	// Channel closed, send close message
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		c.Conn.Close()
		hub.Unregister <- c
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Info(constants.WSReadError+" %v", err)
			}
			break
		}
		// We don't need to handle incoming messages for now
		// This is just to keep the connection alive
	}
}

// BroadcastStatsUpdate broadcasts a stats update to all connected clients
func (h *Hub) BroadcastStatsUpdate(shortCode string, clicks int) {
	update := map[string]interface{}{
		constants.WSMessageType: constants.WSStatsUpdate,
		constants.WSShortCode:   shortCode,
		constants.WSClicks:      clicks,
		constants.WSTimestamp:   time.Now().Unix(),
	}

	message, err := json.Marshal(update)
	if err != nil {
		log.Info(constants.WSMarshalError+" %v", err)
		return
	}

	select {
	case h.Broadcast <- message:
	default:
		log.Info(constants.WSBroadcastFull)
	}
}

// Global hub instance shared across the application
var GlobalHub = &Hub{
	Clients:    make(map[*Client]bool),
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}
