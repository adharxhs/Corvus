package protocol_test

import (
	"encoding/json"
	"testing"

	"server/protocol"
)

func TestParseEnvelopeValid(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"recipient_id": "bob", "content": "hello"})
	raw, _ := json.Marshal(protocol.Envelope{
		Version: 1,
		Type:    protocol.TypeMessage,
		Payload: payload,
	})

	env, err := protocol.ParseEnvelope(raw)
	if err != nil {
		t.Fatalf("expected valid envelope, got %v", err)
	}
	if env.Version != 1 || env.Type != protocol.TypeMessage {
		t.Errorf("unexpected envelope content: %+v", env)
	}
}

func TestParseEnvelopeInvalidVersion(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{})
	raw, _ := json.Marshal(protocol.Envelope{
		Version: 2,
		Type:    protocol.TypeMessage,
		Payload: payload,
	})

	_, err := protocol.ParseEnvelope(raw)
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
}

func TestParseEnvelopeInvalidType(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{})
	raw, _ := json.Marshal(protocol.Envelope{
		Version: 1,
		Type:    "unknown_type",
		Payload: payload,
	})

	_, err := protocol.ParseEnvelope(raw)
	if err == nil {
		t.Fatal("expected error for unknown message type")
	}
}
