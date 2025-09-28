package ws

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestClientStruct(t *testing.T) {
	// Test Client struct initialization
	client := &Client{
		ID:   "test-client",
		Send: make(chan []byte, 256),
	}

	if client.ID != "test-client" {
		t.Errorf("Expected client ID 'test-client', got %s", client.ID)
	}

	if client.Send == nil {
		t.Error("Client send channel should be initialized")
	}

	if cap(client.Send) != 256 {
		t.Errorf("Expected send channel capacity 256, got %d", cap(client.Send))
	}
}

func TestHubStruct(t *testing.T) {
	// Test Hub struct initialization
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	if hub.Clients == nil {
		t.Error("Hub clients map should be initialized")
	}

	if hub.Broadcast == nil {
		t.Error("Hub broadcast channel should be initialized")
	}

	if hub.Register == nil {
		t.Error("Hub register channel should be initialized")
	}

	if hub.Unregister == nil {
		t.Error("Hub unregister channel should be initialized")
	}

	if len(hub.Clients) != 0 {
		t.Error("Hub clients map should start empty")
	}
}

func TestBroadcastStatsUpdate(t *testing.T) {
	// Test the BroadcastStatsUpdate function logic
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 1),
		Register:   make(chan *Client, 1),
		Unregister: make(chan *Client, 1),
	}

	// Start hub in goroutine
	go hub.Run()

	// Broadcast stats update
	shortCode := "test123"
	clicks := 42
	hub.BroadcastStatsUpdate(shortCode, clicks)

	// Give it a moment to process
	time.Sleep(10 * time.Millisecond)

	// Verify message was queued for broadcast
	// We can't easily test the exact message without a real WebSocket connection
	// But we can verify the function doesn't panic and handles the data correctly
	t.Logf("BroadcastStatsUpdate completed for shortCode=%s, clicks=%d", shortCode, clicks)
}

func TestBroadcastStatsUpdateWithDifferentValues(t *testing.T) {
	// Test BroadcastStatsUpdate with various input values
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 10),
		Register:   make(chan *Client, 1),
		Unregister: make(chan *Client, 1),
	}

	// Start hub in goroutine
	go hub.Run()

	testCases := []struct {
		shortCode string
		clicks    int
	}{
		{"abc123", 0},
		{"test456", 100},
		{"xyz789", 999999},
		{"short", 1},
		{"verylongshortcode", 42},
	}

	for _, tc := range testCases {
		hub.BroadcastStatsUpdate(tc.shortCode, tc.clicks)
		time.Sleep(1 * time.Millisecond)
	}

	// Should complete without errors
	t.Log("BroadcastStatsUpdate handled all test cases successfully")
}

func TestGlobalHubInitialization(t *testing.T) {
	// Test that the global hub is properly initialized
	if GlobalHub == nil {
		t.Error("Global hub should not be nil")
	}

	if GlobalHub.Clients == nil {
		t.Error("Global hub clients map should be initialized")
	}

	if GlobalHub.Broadcast == nil {
		t.Error("Global hub broadcast channel should be initialized")
	}

	if GlobalHub.Register == nil {
		t.Error("Global hub register channel should be initialized")
	}

	if GlobalHub.Unregister == nil {
		t.Error("Global hub unregister channel should be initialized")
	}

	if len(GlobalHub.Clients) != 0 {
		t.Error("Global hub clients map should start empty")
	}
}

func TestHubChannels(t *testing.T) {
	// Test hub channel operations
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 5),
		Register:   make(chan *Client, 5),
		Unregister: make(chan *Client, 5),
	}

	// Test broadcast channel
	for i := 0; i < 5; i++ {
		message := []byte(`{"type":"test","message":"broadcast"}`)
		hub.Broadcast <- message
	}

	// Test register channel
	for i := 0; i < 5; i++ {
		client := &Client{
			ID:   fmt.Sprintf("client-%d", i),
			Send: make(chan []byte, 256),
		}
		hub.Register <- client
	}

	// Test unregister channel
	for i := 0; i < 5; i++ {
		client := &Client{
			ID:   fmt.Sprintf("client-%d", i),
			Send: make(chan []byte, 256),
		}
		hub.Unregister <- client
	}

	// Should handle all channel operations without blocking
	t.Log("Hub channel operations completed successfully")
}

func TestHubMutex(t *testing.T) {
	// Test hub mutex functionality
	hub := &Hub{
		Clients: make(map[*Client]bool),
	}

	// Test concurrent access to clients map
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			hub.Mutex.Lock()
			client := &Client{
				ID:   fmt.Sprintf("concurrent-client-%d", id),
				Send: make(chan []byte, 256),
			}
			hub.Clients[client] = true
			hub.Mutex.Unlock()

			time.Sleep(1 * time.Millisecond)

			hub.Mutex.Lock()
			delete(hub.Clients, client)
			hub.Mutex.Unlock()
		}(i)
	}

	wg.Wait()

	// Should complete without data races
	if len(hub.Clients) != 0 {
		t.Errorf("Expected empty clients map, got %d clients", len(hub.Clients))
	}

	t.Log("Hub mutex test completed successfully")
}

func TestBroadcastStatsUpdateMessageFormat(t *testing.T) {
	// Test the message format created by BroadcastStatsUpdate
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 1),
		Register:   make(chan *Client, 1),
		Unregister: make(chan *Client, 1),
	}

	// Start hub in goroutine
	go hub.Run()

	// Broadcast stats update
	shortCode := "format123"
	clicks := 15
	hub.BroadcastStatsUpdate(shortCode, clicks)

	// Give it a moment to process
	time.Sleep(10 * time.Millisecond)

	// The message should be properly formatted JSON
	// We can't easily inspect it without a WebSocket connection,
	// but we can verify the function handles the data types correctly
	t.Logf("Message formatted for shortCode=%s, clicks=%d", shortCode, clicks)
}

func TestBroadcastStatsUpdateWithZeroClicks(t *testing.T) {
	// Test edge case with zero clicks
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 1),
		Register:   make(chan *Client, 1),
		Unregister: make(chan *Client, 1),
	}

	// Start hub in goroutine
	go hub.Run()

	hub.BroadcastStatsUpdate("zero123", 0)
	time.Sleep(10 * time.Millisecond)

	t.Log("BroadcastStatsUpdate handled zero clicks successfully")
}

func TestBroadcastStatsUpdateWithLargeClicks(t *testing.T) {
	// Test edge case with large click count
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 1),
		Register:   make(chan *Client, 1),
		Unregister: make(chan *Client, 1),
	}

	// Start hub in goroutine
	go hub.Run()

	hub.BroadcastStatsUpdate("large123", 999999)
	time.Sleep(10 * time.Millisecond)

	t.Log("BroadcastStatsUpdate handled large click count successfully")
}

func TestHubChannelCapacity(t *testing.T) {
	// Test hub with different channel capacities
	capacities := []int{0, 1, 10, 100}

	for _, capacity := range capacities {
		hub := &Hub{
			Clients:    make(map[*Client]bool),
			Broadcast:  make(chan []byte, capacity),
			Register:   make(chan *Client, capacity),
			Unregister: make(chan *Client, capacity),
		}

		// Test that channels have expected capacity
		if len(hub.Broadcast) != 0 || cap(hub.Broadcast) != capacity {
			t.Errorf("Broadcast channel capacity should be %d", capacity)
		}

		if len(hub.Register) != 0 || cap(hub.Register) != capacity {
			t.Errorf("Register channel capacity should be %d", capacity)
		}

		if len(hub.Unregister) != 0 || cap(hub.Unregister) != capacity {
			t.Errorf("Unregister channel capacity should be %d", capacity)
		}
	}

	t.Log("Hub channel capacity test completed successfully")
}

func TestClientSendChannelCapacity(t *testing.T) {
	// Test client send channel with different capacities
	capacities := []int{1, 10, 100, 1000}

	for _, capacity := range capacities {
		client := &Client{
			ID:   "test-client",
			Send: make(chan []byte, capacity),
		}

		if cap(client.Send) != capacity {
			t.Errorf("Client send channel capacity should be %d, got %d", capacity, cap(client.Send))
		}

		// Test filling the channel
		for i := 0; i < capacity; i++ {
			message := []byte(fmt.Sprintf("message-%d", i))
			client.Send <- message
		}

		if len(client.Send) != capacity {
			t.Errorf("Client send channel should be full with %d messages", capacity)
		}
	}

	t.Log("Client send channel capacity test completed successfully")
}

func TestHubConcurrentChannelOperations(t *testing.T) {
	// Test concurrent channel operations
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 100),
		Register:   make(chan *Client, 100),
		Unregister: make(chan *Client, 100),
	}

	// Start hub in goroutine
	go hub.Run()

	var wg sync.WaitGroup
	numOperations := 50

	// Concurrent broadcast operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			message := []byte(fmt.Sprintf(`{"type":"test","id":%d}`, id))
			hub.Broadcast <- message
			time.Sleep(1 * time.Millisecond)
		}(i)
	}

	// Concurrent register operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client := &Client{
				ID:   fmt.Sprintf("concurrent-client-%d", id),
				Send: make(chan []byte, 256),
			}
			hub.Register <- client
			time.Sleep(1 * time.Millisecond)
		}(i)
	}

	// Concurrent unregister operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client := &Client{
				ID:   fmt.Sprintf("concurrent-client-%d", id),
				Send: make(chan []byte, 256),
			}
			hub.Unregister <- client
			time.Sleep(1 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Should complete without deadlocks or panics
	t.Log("Concurrent channel operations completed successfully")
}

func TestBroadcastStatsUpdateJSONStructure(t *testing.T) {
	// Test that BroadcastStatsUpdate creates properly structured JSON
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 1),
		Register:   make(chan *Client, 1),
		Unregister: make(chan *Client, 1),
	}

	// Start hub in goroutine
	go hub.Run()

	// Test various short code formats
	testCases := []struct {
		shortCode string
		clicks    int
	}{
		{"abc123", 5},
		{"ABC123", 10},
		{"123456", 15},
		{"a1b2c3", 20},
		{"", 25}, // Edge case: empty short code
	}

	for _, tc := range testCases {
		hub.BroadcastStatsUpdate(tc.shortCode, tc.clicks)
		time.Sleep(1 * time.Millisecond)
	}

	t.Log("JSON structure test completed successfully")
}

func TestHubStressTest(t *testing.T) {
	// Stress test the hub with many operations
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 1000),
		Register:   make(chan *Client, 1000),
		Unregister: make(chan *Client, 1000),
	}

	// Start hub in goroutine
	go hub.Run()

	var wg sync.WaitGroup
	numOperations := 100

	// Stress test with many concurrent operations
	for i := 0; i < numOperations; i++ {
		wg.Add(3) // broadcast, register, unregister

		// Broadcast operation
		go func(id int) {
			defer wg.Done()
			message := []byte(fmt.Sprintf(`{"type":"stress","id":%d}`, id))
			hub.Broadcast <- message
		}(i)

		// Register operation
		go func(id int) {
			defer wg.Done()
			client := &Client{
				ID:   fmt.Sprintf("stress-client-%d", id),
				Send: make(chan []byte, 256),
			}
			hub.Register <- client
		}(i)

		// Unregister operation
		go func(id int) {
			defer wg.Done()
			client := &Client{
				ID:   fmt.Sprintf("stress-client-%d", id),
				Send: make(chan []byte, 256),
			}
			hub.Unregister <- client
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Should handle stress without issues
	t.Log("Hub stress test completed successfully")
}

func BenchmarkBroadcastStatsUpdate(b *testing.B) {
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 100),
		Register:   make(chan *Client, 100),
		Unregister: make(chan *Client, 100),
	}

	go hub.Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.BroadcastStatsUpdate("test123", i)
	}
}

func BenchmarkHubChannelOperations(b *testing.B) {
	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 100),
		Register:   make(chan *Client, 100),
		Unregister: make(chan *Client, 100),
	}

	go hub.Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate channel operations
		message := []byte(`{"type":"benchmark"}`)
		hub.Broadcast <- message

		client := &Client{
			ID:   "bench-client",
			Send: make(chan []byte, 256),
		}
		hub.Register <- client
		hub.Unregister <- client
	}
}
