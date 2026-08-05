package protocol

import "encoding/json"

// DirectMessagePayload is the payload for a direct (1:1) message.
type DirectMessagePayload struct {
	RecipientID string `json:"recipient_id"`
	Content     string `json:"content"`
}

// ParseDirectMessage extracts the typed payload from an envelope.
func ParseDirectMessage(e *Envelope) (*DirectMessagePayload, error) {
	var p DirectMessagePayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return nil, NewError(ErrInvalidPayload, "bad message payload")
	}
	return &p, nil
}

// GroupMessagePayload is the payload for a group message.
type GroupMessagePayload struct {
	GroupID string `json:"group_id"`
	Content string `json:"content"`
}

func ParseGroupMessage(e *Envelope) (*GroupMessagePayload, error) {
	var p GroupMessagePayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return nil, NewError(ErrInvalidPayload, "bad group_message payload")
	}
	return &p, nil
}

// SenderKeyDistributionPayload is the payload for sender key distribution.
type SenderKeyDistributionPayload struct {
	GroupID string `json:"group_id"`
	Content string `json:"content"`
}

func ParseSenderKeyDistribution(e *Envelope) (*SenderKeyDistributionPayload, error) {
	var p SenderKeyDistributionPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return nil, NewError(ErrInvalidPayload, "bad sender_key_distribution payload")
	}
	return &p, nil
}
