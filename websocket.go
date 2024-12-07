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
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

// NewWebSocketClient initializes a new WebSocket client and connects to the provided URL.
// It returns an instance of the WebSocket client.
func NewWebSocketClient(url string) *WebSocket {

	ctx, cancel := context.WithCancel(context.Background())

	ws := &WebSocket{
		Context:          ctx,
		Cancel:           cancel,
		Connected:        make(chan bool),
		Disconnected:     make(chan error),
		IncomingMessages: make(chan Message, 50),
		outgoingMessages: make(chan Message, 50),
	}

	// Goroutine to establish WebSocket connection
	go func() {
		var err error
		ws.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			ws.closeConnection(fmt.Errorf("failed to establish connection: %w", err))
			return
		}
		ws.Connected <- true

		// Start reading and writing goroutines once connected
		go ws.readMessages()
		go ws.writeMessages()
	}()

	return ws
}

// readMessages continuously reads messages from the WebSocket connection and sends
// them to the IncomingMessages channel.
func (ws *WebSocket) readMessages() {
	defer func() {
		if r := recover(); r != nil {
			ws.closeConnection(fmt.Errorf("panic during read message: %v", r))
		}
	}()

	for {
		select {
		case <-ws.Context.Done(): // Context cancellation detected
			ws.closeConnection(nil)
			return
		default:
			// Read message from WebSocket connection
			messageType, message, err := ws.conn.ReadMessage()
			if err != nil {
				// Properly handle both normal and unexpected close events
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					ws.closeConnection(nil)
				} else {
					ws.closeConnection(err)
				}
				return
			}
			// Send the received message to the IncomingMessages channel
			ws.IncomingMessages <- Message{
				Type: MessageType(messageType),
				Data: message,
			}
		}
	}
}

// writeMessages continuously writes messages from the outgoingMessages channel to the
// WebSocket connection. Stops writing if the context is canceled or the connection closes.
func (ws *WebSocket) writeMessages() {
	for {
		select {
		case <-ws.Context.Done(): // Context cancellation detected
			ws.closeConnection(nil)
			return
		case message, ok := <-ws.outgoingMessages:
			if !ok {
				// Channel is closed; exit the loop.
				return
			}
			ws.writeMutex.Lock()
			err := ws.conn.WriteMessage(int(message.Type), message.Data)
			ws.writeMutex.Unlock()
			if err != nil {
				ws.closeConnection(err)
				return
			}
		}
	}
}

// SendMessage queues a message for sending through the WebSocket connection.
// Panics due to channel closure are gracefully handled.
func (ws *WebSocket) SendMessage(messageType MessageType, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			ws.closeConnection(fmt.Errorf("panic during message send: %v", r))
		}
	}()

	select {
	case <-ws.Context.Done():
		// Connection context is canceled; message is not sent.
	case ws.outgoingMessages <- Message{Type: messageType, Data: data}:
		// Message successfully enqueued.
	default:
		// Buffer is full; close the connection with an error.
		ws.closeConnection(errors.New("outgoing message buffer is full"))
	}
}

// closeConnection safely closes the WebSocket connection and releases all resources.
// It ensures the closure happens only once, and all channels are properly closed.
// The provided error is optional. If an error is passed, it will be sent to the Disconnected channel.
func (ws *WebSocket) closeConnection(err error) {
	ws.closeOnce.Do(func() {
		// Cancel the context to release related resources
		ws.Cancel()

		// Properly close the WebSocket connection if it is still open
		if ws.conn != nil {
			closeErr := ws.conn.Close()
			if closeErr != nil {
				fmt.Println("[websocket] error closing WebSocket connection:", closeErr)
			}
		}

		// Notify any listener that the connection has been closed
		ws.Disconnected <- err

		// Close all channels to free resources
		close(ws.Connected)
		close(ws.Disconnected)
		close(ws.IncomingMessages)
		close(ws.outgoingMessages)
	})
}
