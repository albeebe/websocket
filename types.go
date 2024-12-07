// Copyright (c) 2024 Alan Beebe [www.alanbeebe.com]
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// Created: October 18, 2024

package websocket

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

// Define constants for different message types
type MessageType int

const (
	TextMessage   MessageType = 1  // Text message (e.g., JSON, plain text)
	BinaryMessage MessageType = 2  // Binary message (e.g., images, files)
	CloseMessage  MessageType = 8  // Close connection request
	PingMessage   MessageType = 9  // Ping to check connection
	PongMessage   MessageType = 10 // Pong response to a ping
)

// WebSocket manages the lifecycle and communication of a WebSocket connection
type WebSocket struct {
	Context          context.Context    // Context for managing lifecycle
	Cancel           context.CancelFunc // Function to cancel context
	Connected        chan bool          // Channel for connection status
	Disconnected     chan error         // Channel for disconnect notifications
	IncomingMessages chan Message       // Channel for received messages
	outgoingMessages chan Message       // Channel for outgoing messages
	closeOnce        sync.Once          // Ensures close logic runs only once
	writeMutex       sync.Mutex         // Synchronizes writes to the WebSocket connection.
	conn             *websocket.Conn    // Underlying WebSocket connection
}

// Message represents a message sent or received over a WebSocket connection
type Message struct {
	Type MessageType // Type of the message (e.g., Text, Binary)
	Data []byte      // Message payload data
}
