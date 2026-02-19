package main

// Message is just a structure that tells us what a message looks like
type Message struct {
	Type    string `json:"type"`             // What kind of message is this? Broadcast, direct, register
	Target  string `json:"target,omitempty"` // Who is this message for? (If it's a secret)
	Content string `json:"content"`          // The actual text or data being sent
	From    string `json:"from,omitempty"`   // Who sent this? (Server will figure this out)
}
