package protocol

import "encoding/json"

// Envelope is the top-level protocol message structure. Every message
// exchanged between client and server must be wrapped in this envelope.
type Envelope struct {
	Version int             `json:"version"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
