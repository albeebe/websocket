# WebSocket Package

This package provides a wrapper around the gorilla/websocket package to simplify WebSocket client usage. It offers a more intuitive interface for managing WebSocket connections by leveraging channels for sending and receiving messages, as well as context for handling lifecycle and cancellation.

## Features

- Channel-based message handling for sending and receiving messages, providing a more intuitive and thread-safe interface
- Graceful shutdown and resource cleanup through context-based cancellation

## Installation

```bash
go get github.com/albeebe/websocket
```

## Example Usage

```go
package main

import (
   "fmt"
   "time"

   "github.com/albeebe/websocket"
)

func main() {
   // Create a new WebSocket client
   url := "wss://example.com/socket"
   client := websocket.NewWebSocketClient(url)

   // Monitor WebSocket client state changes and process incoming messages
   go func() {
      for {
         select {
         case <-client.Connected:
            fmt.Println("WebSocket client connected successfully!")

         case err := <-client.Disconnected:
            if err != nil {
               fmt.Println("WebSocket client disconnected with error:", err)
            } else {
               fmt.Println("WebSocket client disconnected")
            }
            return // Exit the loop and terminate the Goroutine after disconnection

         case msg := <-client.IncomingMessages:
            switch msg.Type {
            case websocket.TextMessage:
               fmt.Println("Received text message:", string(msg.Data))
            case websocket.BinaryMessage:
               fmt.Println("Received binary message of length:", len(msg.Data))
            case websocket.PingMessage:
               fmt.Println("Received ping message")
            case websocket.PongMessage:
               fmt.Println("Received pong message")
            }
         }
      }
   }()

   // Send a message to the server
   messageData := []byte("Hello WebSocket Server!")
   client.SendMessage(websocket.TextMessage, messageData)

   // Keep the client running to receive messages
   time.Sleep(10 * time.Second)

   // Close the WebSocket connection
   client.Close()
}
```
