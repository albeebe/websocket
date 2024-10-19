package websocket_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/albeebe/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestWebSocketClientConnection tests if the WebSocket client successfully connects to a server
func TestWebSocketClientConnection(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gorilla.Upgrader{}
		_, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal("Failed to upgrade connection", err)
		}
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:]

	// Initialize WebSocket client
	client := websocket.NewWebSocketClient(wsURL)

	// Wait until connection state is determined
	select {
	case <-client.Connected:
		// Connected successfully
		assert.True(t, true)
	case err := <-client.Disconnected:
		t.Errorf("Failed to connect: %v", err)
	case <-time.After(time.Second * 2):
		t.Error("Connection timed out")
	}
}

// TestWebSocketClientSendMessage tests if messages can be sent from the client to server
func TestWebSocketClientSendMessage(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gorilla.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal("Failed to upgrade connection", err)
		}
		// Read message
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
		}
		assert.Equal(t, int(websocket.TextMessage), int(messageType))
		assert.Equal(t, []byte("Hello"), message)
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:]

	// Initialize WebSocket client
	client := websocket.NewWebSocketClient(wsURL)

	select {
	case <-client.Connected:
		// Send message
		client.SendMessage(websocket.TextMessage, []byte("Hello"))
	case err := <-client.Disconnected:
		t.Errorf("Failed to connect: %v", err)
	case <-time.After(time.Second * 2):
		t.Error("Connection timed out")
	}
}

// TestWebSocketClientReceiveMessage tests if messages can be received from server to client
func TestWebSocketClientReceiveMessage(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gorilla.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal("Failed to upgrade connection", err)
		}
		// Send message to client
		err = conn.WriteMessage(gorilla.TextMessage, []byte("Hello from server"))
		if err != nil {
			t.Errorf("Failed to write message: %v", err)
		}
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:]

	// Initialize WebSocket client
	client := websocket.NewWebSocketClient(wsURL)

	select {
	case <-client.Connected:
		// Wait to receive a message
		select {
		case msg := <-client.IncomingMessages:
			assert.Equal(t, websocket.TextMessage, msg.Type)
			assert.Equal(t, []byte("Hello from server"), msg.Data)
		case <-time.After(time.Second * 2):
			t.Error("Message receiving timed out")
		}
	case err := <-client.Disconnected:
		t.Errorf("Failed to connect: %v", err)
	case <-time.After(time.Second * 2):
		t.Error("Connection timed out")
	}
}

// TestWebSocketClientDisconnect tests if the WebSocket client properly handles disconnection
func TestWebSocketClientDisconnect(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gorilla.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal("Failed to upgrade connection", err)
		}
		// Close connection after some time
		time.Sleep(time.Millisecond * 500)
		conn.Close()
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:]

	// Initialize WebSocket client
	client := websocket.NewWebSocketClient(wsURL)

	select {
	case <-client.Connected:
		// Wait for disconnection
		select {
		case <-client.Disconnected:

		case <-time.After(time.Second * 2):
			t.Error("Disconnection timed out")
		}
	case <-time.After(time.Second * 2):
		t.Error("Connection timed out")
	}
}
