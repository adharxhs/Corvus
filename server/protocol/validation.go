package protocol

import (
	"encoding/json"
	"fmt"
)

// ValidateEnvelope checks that an envelope has a supported version,
// recognized message type, and a non-empty JSON payload.
func ValidateEnvelope(e *Envelope) error {
	if e.Version != CurrentVersion {
		return NewError(ErrInvalidVersion, fmt.Sprintf("unsupported version %d", e.Version))
	}
	if !supportedTypes[e.Type] {
		return NewError(ErrInvalidType, fmt.Sprintf("unknown type %q", e.Type))
	}
	if len(e.Payload) == 0 || !json.Valid(e.Payload) {
		return NewError(ErrInvalidPayload, "missing or invalid payload")
	}
	return nil
}
