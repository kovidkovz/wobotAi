package main

// Message defines the JSON structure for WebSocket communication
type Message struct {
	Type    string `json:"type"`             // "broadcast", "direct", "register"
	Target  string `json:"target,omitempty"` // Target Client ID (for direct messages)
	Content string `json:"content"`          // The actual message content
	From    string `json:"from,omitempty"`   // Sender ID (server fills this)
}
