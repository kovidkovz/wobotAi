package main

import (
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Mutex to protect the clients map
	mu sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client Registered: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
				log.Printf("Client Unregistered: %s", client.ID)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.handleMessage(msg)
		}
	}
}

func (h *Hub) handleMessage(msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If a specific target is set, try to send to that client
	if msg.Target != "" {
		if client, ok := h.clients[msg.Target]; ok {
			select {
			case client.send <- []byte(msg.Content):
			default:
				close(client.send)
				delete(h.clients, msg.Target)
			}
		}
		return
	}

	// Otherwise, broadcast to all
	for id, client := range h.clients {
		// Don't send back to sender for broadcast (optional, but good practice usually)
		// But here we might want echo. Let's send to all for now.
		select {
		case client.send <- []byte(msg.Content):
		default:
			close(client.send)
			delete(h.clients, id)
		}
	}
}

// getClients returns a list of connected client IDs.
func (h *Hub) getClients() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clientIDs []string
	for id := range h.clients {
		clientIDs = append(clientIDs, id)
	}
	return clientIDs
}
