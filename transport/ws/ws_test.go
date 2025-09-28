package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	wsutil "github.com/kweusuf/pocketbase-demo/pkg/utils/ws"
)

// Mock WebSocket connection for testing
type mockWSConn struct {
	sentMessages [][]byte
	closed       bool
	readError    error
}

func (m *mockWSConn) WriteMessage(messageType int, data []byte) error {
	if m.closed {
		return websocket.ErrCloseSent
	}
	m.sentMessages = append(m.sentMessages, data)
	return nil
}

func (m *mockWSConn) ReadMessage() (messageType int, data []byte, err error) {
	if m.readError != nil {
		return 0, nil, m.readError
	}
	// Block indefinitely for testing
	select {}
}

func (m *mockWSConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockWSConn) getSentMessages() [][]byte {
	return m.sentMessages
}

func TestHandleWebSocket(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Give it a moment to establish connection
	time.Sleep(100 * time.Millisecond)

	// Verify connection is established
	// The handler should have created a client and registered it
	if wsutil.GlobalHub == nil {
		t.Error("Global hub should be initialized")
	}

	// Check that a client was registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount == 0 {
		t.Error("Client should be registered after WebSocket connection")
	}
}

func TestHandleWebSocketWithInvalidUpgrade(t *testing.T) {
	// Test with invalid upgrade request
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	// This should not panic and should handle the error gracefully
	HandleWebSocket(w, req)

	// Response should indicate an error (websocket upgrade failed)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleWebSocketWithValidHeaders(t *testing.T) {
	// Create test server with WebSocket handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Convert to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Set up WebSocket dialer with valid headers
	dialer := websocket.DefaultDialer
	headers := http.Header{}
	headers.Add("Origin", "http://localhost")

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("Failed to connect with valid headers: %v", err)
	}
	defer conn.Close()

	// Verify successful connection
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify client was registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount == 0 {
		t.Error("Client should be registered after successful WebSocket connection")
	}
}

func TestHandleWebSocketMultipleConnections(t *testing.T) {
	// Test handling multiple WebSocket connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	numConnections := 5

	var conns []*websocket.Conn
	for i := 0; i < numConnections; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to create connection %d: %v", i, err)
		}
		conns = append(conns, conn)
	}

	// Give connections time to establish
	time.Sleep(200 * time.Millisecond)

	// Verify all clients were registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != numConnections {
		t.Errorf("Expected %d registered clients, got %d", numConnections, clientCount)
	}

	// Clean up connections
	for _, conn := range conns {
		conn.Close()
	}

	// Give time for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify all clients were unregistered
	wsutil.GlobalHub.Mutex.Lock()
	finalClientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if finalClientCount != 0 {
		t.Errorf("Expected 0 remaining clients after cleanup, got %d", finalClientCount)
	}
}

func TestHandleWebSocketConnectionLifecycle(t *testing.T) {
	// Test complete WebSocket connection lifecycle
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Establish connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to establish WebSocket connection: %v", err)
	}

	// Verify client was registered
	time.Sleep(100 * time.Millisecond)
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}

	// Close connection
	conn.Close()

	// Verify client was unregistered
	time.Sleep(100 * time.Millisecond)
	wsutil.GlobalHub.Mutex.Lock()
	finalClientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if finalClientCount != 0 {
		t.Errorf("Expected 0 remaining clients after close, got %d", finalClientCount)
	}
}

func TestHandleWebSocketWithSubProtocols(t *testing.T) {
	// Test WebSocket connection with subprotocols
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create dialer with subprotocols
	dialer := websocket.DefaultDialer
	dialer.Subprotocols = []string{"chat", "superchat"}

	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect with subprotocols: %v", err)
	}
	defer conn.Close()

	// Verify successful connection
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify client was registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}
}

func TestHandleWebSocketConcurrentConnections(t *testing.T) {
	// Test concurrent WebSocket connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	numConnections := 10

	var conns []*websocket.Conn
	for i := 0; i < numConnections; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to create connection %d: %v", i, err)
		}
		conns = append(conns, conn)

		// Small delay to avoid overwhelming the server
		time.Sleep(10 * time.Millisecond)
	}

	// Give connections time to establish
	time.Sleep(200 * time.Millisecond)

	// Verify all clients were registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != numConnections {
		t.Errorf("Expected %d registered clients, got %d", numConnections, clientCount)
	}

	// Clean up all connections
	for _, conn := range conns {
		conn.Close()
	}

	// Give time for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify all clients were unregistered
	wsutil.GlobalHub.Mutex.Lock()
	finalClientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if finalClientCount != 0 {
		t.Errorf("Expected 0 remaining clients after cleanup, got %d", finalClientCount)
	}
}

func TestHandleWebSocketWithInvalidOrigin(t *testing.T) {
	// Test with invalid origin (though our upgrader allows all origins)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create connection with suspicious origin
	headers := http.Header{}
	headers.Add("Origin", "http://malicious-site.com")

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("Failed to connect with suspicious origin: %v", err)
	}
	defer conn.Close()

	// Should still succeed since we allow all origins
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify client was registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}
}

func TestHandleWebSocketConnectionError(t *testing.T) {
	// Test handling of connection errors
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Sec-WebSocket-Key", "invalid-key")
	req.Header.Set("Sec-WebSocket-Version", "13")

	w := httptest.NewRecorder()

	// This should handle the error gracefully
	HandleWebSocket(w, req)

	// Should return bad request due to invalid headers
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleWebSocketWithQueryParameters(t *testing.T) {
	// Test WebSocket connection with query parameters
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?param1=value1&param2=value2"

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect with query parameters: %v", err)
	}
	defer conn.Close()

	// Verify successful connection
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify client was registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}
}

func TestHandleWebSocketStressTest(t *testing.T) {
	// Stress test with many rapid connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	numConnections := 50

	var conns []*websocket.Conn
	for i := 0; i < numConnections; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to create connection %d: %v", i, err)
		}
		conns = append(conns, conn)
	}

	// Give connections time to establish
	time.Sleep(500 * time.Millisecond)

	// Check client count
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != numConnections {
		t.Errorf("Expected %d registered clients, got %d", numConnections, clientCount)
	}

	// Close all connections rapidly
	for _, conn := range conns {
		conn.Close()
	}

	// Give time for cleanup
	time.Sleep(500 * time.Millisecond)

	// Verify all clients were unregistered
	wsutil.GlobalHub.Mutex.Lock()
	finalClientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if finalClientCount != 0 {
		t.Errorf("Expected 0 remaining clients after stress test, got %d", finalClientCount)
	}
}

func TestHandleWebSocketWithCustomHeaders(t *testing.T) {
	// Test WebSocket connection with custom headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create dialer with custom headers
	dialer := websocket.DefaultDialer
	headers := http.Header{}
	headers.Add("Authorization", "Bearer token123")
	headers.Add("X-Custom-Header", "custom-value")
	headers.Add("User-Agent", "Go-Test-Client")

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("Failed to connect with custom headers: %v", err)
	}
	defer conn.Close()

	// Verify successful connection
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify client was registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}
}

func TestHandleWebSocketConnectionTimeout(t *testing.T) {
	// Test WebSocket connection behavior under timeout conditions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add delay to simulate slow connection
		time.Sleep(50 * time.Millisecond)
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Set short timeout
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Millisecond

	_, _, err := dialer.Dial(wsURL, nil)
	// This might succeed or fail depending on system speed
	// The important thing is it doesn't panic
	t.Logf("Connection with timeout completed with error: %v", err)
}

func TestHandleWebSocketWithLargeHeaders(t *testing.T) {
	// Test WebSocket connection with large headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create large header value
	largeValue := strings.Repeat("x", 10000)
	headers := http.Header{}
	headers.Add("X-Large-Header", largeValue)

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("Failed to connect with large headers: %v", err)
	}
	defer conn.Close()

	// Verify successful connection
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify client was registered
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}
}

func TestHandleWebSocketConnectionPersistence(t *testing.T) {
	// Test that WebSocket connections persist correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to establish WebSocket connection: %v", err)
	}
	defer conn.Close()

	// Verify client was registered
	time.Sleep(100 * time.Millisecond)
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}

	// Keep connection alive for a short period
	time.Sleep(200 * time.Millisecond)

	// Verify client is still registered
	wsutil.GlobalHub.Mutex.Lock()
	stillClientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if stillClientCount != 1 {
		t.Errorf("Expected 1 persistent client, got %d", stillClientCount)
	}
}

func TestHandleWebSocketErrorRecovery(t *testing.T) {
	// Test error recovery when WebSocket operations fail
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Establish connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to establish WebSocket connection: %v", err)
	}

	// Verify client was registered
	time.Sleep(100 * time.Millisecond)
	wsutil.GlobalHub.Mutex.Lock()
	clientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 registered client, got %d", clientCount)
	}

	// Close connection abruptly (simulating network error)
	conn.Close()

	// Give time for error handling
	time.Sleep(200 * time.Millisecond)

	// Verify client was unregistered after error
	wsutil.GlobalHub.Mutex.Lock()
	finalClientCount := len(wsutil.GlobalHub.Clients)
	wsutil.GlobalHub.Mutex.Unlock()

	if finalClientCount != 0 {
		t.Errorf("Expected 0 remaining clients after error, got %d", finalClientCount)
	}
}

func BenchmarkHandleWebSocket(b *testing.B) {
	// Benchmark WebSocket connection handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			b.Fatalf("Failed to connect: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkHandleWebSocketConcurrent(b *testing.B) {
	// Benchmark concurrent WebSocket connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				b.Fatalf("Failed to connect: %v", err)
			}
			conn.Close()
		}
	})
}
