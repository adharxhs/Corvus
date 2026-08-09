package websocket

import (
	"encoding/json"

	"server/protocol"
)

// sendPresenceSnapshot sends the currently-online accepted contacts of the
// given client immediately after connection. The snapshot is sent BEFORE
// pending messages so FIFO ordering on the client's single WS connection
// guarantees any live presence events arriving after the snapshot resolve
// correctly.
func (s *Server) sendPresenceSnapshot(client *Client) {
	peers, err := s.repos.Relationships.AcceptedPeers(client.UserID)
	if err != nil {
		s.logger.Error("failed to compute presence snapshot", "user", client.UserID, "error", err)
		return
	}

	var online []string
	for _, peerID := range peers {
		if _, isOnline := s.registry.Get(peerID); isOnline {
			online = append(online, peerID)
		}
	}

	payloadBytes, err := json.Marshal(protocol.PresenceSnapshotPayload{Online: online})
	if err != nil {
		s.logger.Error("failed to marshal presence snapshot payload", "user", client.UserID, "error", err)
		return
	}

	env := protocol.Envelope{
		Version: protocol.CurrentVersion,
		Type:    protocol.TypePresenceSnapshot,
		Payload: json.RawMessage(payloadBytes),
	}
	data, err := protocol.Encode(env)
	if err != nil {
		s.logger.Error("failed to encode presence snapshot", "user", client.UserID, "error", err)
		return
	}

	client.Send(data)
}

// broadcastPresence notifies all online accepted contacts of the given user
// that the user's presence changed. Presence is never queued for offline
// contacts — a fresh snapshot on reconnect supersedes any missed events.
func (s *Server) broadcastPresence(userID, status string) {
	peers, err := s.repos.Relationships.AcceptedPeers(userID)
	if err != nil {
		s.logger.Error("failed to compute presence broadcast peers", "user", userID, "error", err)
		return
	}

	payloadBytes, err := json.Marshal(protocol.PresencePayload{UserID: userID, Status: status})
	if err != nil {
		s.logger.Error("failed to marshal presence payload", "user", userID, "error", err)
		return
	}

	env := protocol.Envelope{
		Version: protocol.CurrentVersion,
		Type:    protocol.TypePresence,
		Payload: json.RawMessage(payloadBytes),
	}
	data, err := protocol.Encode(env)
	if err != nil {
		s.logger.Error("failed to encode presence update", "user", userID, "error", err)
		return
	}

	for _, peerID := range peers {
		if peer, ok := s.registry.Get(peerID); ok {
			peer.Send(data)
		}
	}
}
