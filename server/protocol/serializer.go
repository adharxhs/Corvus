package protocol

import (
	"encoding/json"
	"fmt"
)

// ParseEnvelope decodes raw JSON bytes into an Envelope and validates it.
func ParseEnvelope(data []byte) (*Envelope, error) {
	var e Envelope
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, NewError(ErrMalformedJSON, fmt.Sprintf("json decode: %v", err))
	}
	if err := ValidateEnvelope(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Encode serializes any value to JSON bytes.
func Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

// EncodeError creates a complete error envelope ready to send over WebSocket.
func EncodeError(code ErrorCode, msg string) []byte {
	env := Envelope{
		Version: CurrentVersion,
		Type:    TypeError,
	}
	env.Payload, _ = json.Marshal(&Error{Code: code, Message: msg})
	data, _ := json.Marshal(env)
	return data
}
