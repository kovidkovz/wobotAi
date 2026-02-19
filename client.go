package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// How long we wait to send a message before giving up.
	writeWait = 10 * time.Second

	// How long we wait for a 'alive' signal from the client.
	pongWait = 60 * time.Second

	// How often we poke the client to see if they are still there. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// The biggest message we allow.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
)

// Client represents a single user connected to our server.
type Client struct {
	hub *Hub

	// A unique name tag for this user
	ID string

	// The actual connection line to the user.
	conn *websocket.Conn

	// A queue of messages waiting to be sent to this user.
	send chan []byte
}

// readPump listens for messages coming from this user and passes them to the Hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Parse JSON message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			// If the message isn't proper JSON, we'll just wrap it up and send it to everyone.
			msg = Message{
				Type:    "broadcast",
				Content: string(message),
				From:    c.ID,
			}
		} else {
			msg.From = c.ID
		}

		c.hub.broadcast <- &msg
	}
}

// writePump takes messages from the Hub and sends them out to this user.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles new people trying to connect via WebSocket.
func serveWs(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// Giving this new person a unique ID
	id := uuid.New().String()
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), ID: id}
	client.hub.register <- client

	// Starting completely new background tasks to handle reading and writing for this user.
	go client.writePump()
	go client.readPump()
}
