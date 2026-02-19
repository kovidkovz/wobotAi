package main

import (
	"log"
	"sync"
)

// Hub is like a chat room manager. It keeps track of everyone and sends messages around.
type Hub struct {
	// All the people currently connected
	clients map[string]*Client

	// Messages waiting to be sent out
	broadcast chan *Message

	// New people trying to join
	register chan *Client

	// People trying to leave
	unregister chan *Client

	// A lock to make sure we don't mess up the client list when multiple things happen at once
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

	// If this message is meant for a specific person, let's try to send it to them
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

	// Otherwise, shout it out to everyone!
	for id, client := range h.clients {
		// Usually we wouldn't echo back to the sender, but let's just send it to everyone for now.
		select {
		case client.send <- []byte(msg.Content):
		default:
			close(client.send)
			delete(h.clients, id)
		}
	}
}

// getClients helps us see a list of everyone who is online.
func (h *Hub) getClients() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clientIDs []string
	for id := range h.clients {
		clientIDs = append(clientIDs, id)
	}
	return clientIDs
}
