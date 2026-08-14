package websocket

import (
	"log/slog"

	"github.com/google/uuid"

	"server/models"
	"server/protocol"
	"server/repository"
)

type Dispatcher struct {
	registry *Registry
	repos    *repository.Repository
	logger   *slog.Logger
}

func NewDispatcher(registry *Registry, repos *repository.Repository, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		registry: registry,
		repos:    repos,
		logger:   logger,
	}
}

func (d *Dispatcher) Dispatch(sender *Client, env *protocol.Envelope) error {
	switch env.Type {
	case protocol.TypeMessage:
		return d.handleDirectMessage(sender, env)
	case protocol.TypeGroupMessage:
		return d.handleGroupMessage(sender, env)
	case protocol.TypeSenderKeyDistribution:
		return d.handleSenderKeyDistribution(sender, env)
	case protocol.TypeProfilePictureUpdated:
		return d.handleProfilePictureUpdated(sender, env)
	case protocol.TypeGroupProfilePictureUpdated:
		return d.handleGroupProfilePictureUpdated(sender, env)
	default:
		return protocol.NewError(protocol.ErrInvalidType, "unsupported message type")
	}
}

func (d *Dispatcher) handleDirectMessage(sender *Client, env *protocol.Envelope) error {
	payload, err := protocol.ParseDirectMessage(env)
	if err != nil {
		return err
	}

	// Direct messages require an accepted relationship in either direction;
	// discovering a user ID must not grant messaging rights.
	accepted, err := d.repos.Relationships.HasAcceptedBetween(sender.UserID, payload.RecipientID)
	if err != nil {
		d.logger.Error("failed to check relationship for direct message",
			"sender", sender.UserID, "recipient", payload.RecipientID, "error", err)
		return err
	}
	if !accepted {
		return protocol.NewError(protocol.ErrRelationshipRequired,
			"no accepted relationship with recipient; request pending")
	}

	payload.SenderID = sender.UserID
	if err := d.reencodePayload(env, payload); err != nil {
		return err
	}

	data, err := protocol.Encode(env)
	if err != nil {
		return err
	}

	recipient, online := d.registry.Get(payload.RecipientID)
	if online {
		recipient.Send(data)
	} else {
		msg := &models.PendingMessage{
			ID:          uuid.New().String(),
			RecipientID: payload.RecipientID,
			Ciphertext:  data,
			Delivered:   false,
		}
		if err := d.repos.Messages.Insert(msg); err != nil {
			d.logger.Error("failed to queue offline message",
				"recipient", payload.RecipientID, "error", err)
			return err
		}
	}
	return nil
}

func (d *Dispatcher) handleGroupMessage(sender *Client, env *protocol.Envelope) error {
	payload, err := protocol.ParseGroupMessage(env)
	if err != nil {
		return err
	}

	members, err := d.repos.Groups.ListMembers(payload.GroupID)
	if err != nil {
		return err
	}

	payload.SenderID = sender.UserID
	if err := d.reencodePayload(env, payload); err != nil {
		return err
	}

	data, err := protocol.Encode(env)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.UserID == sender.UserID {
			continue
		}
		recipient, online := d.registry.Get(member.UserID)
		if online {
			recipient.Send(data)
		} else {
			msg := &models.PendingMessage{
				ID:          uuid.New().String(),
				RecipientID: member.UserID,
				Ciphertext:  data,
				Delivered:   false,
			}
			if err := d.repos.Messages.Insert(msg); err != nil {
				d.logger.Error("failed to queue offline group message",
					"group", payload.GroupID, "recipient", member.UserID, "error", err)
			}
		}
	}
	return nil
}

func (d *Dispatcher) handleSenderKeyDistribution(sender *Client, env *protocol.Envelope) error {
	payload, err := protocol.ParseSenderKeyDistribution(env)
	if err != nil {
		return err
	}

	members, err := d.repos.Groups.ListMembers(payload.GroupID)
	if err != nil {
		return err
	}

	data, err := protocol.Encode(env)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.UserID == sender.UserID {
			continue
		}
		recipient, online := d.registry.Get(member.UserID)
		if online {
			recipient.Send(data)
		} else {
			dist := &models.SenderKeyDistribution{
				ID:          uuid.New().String(),
				GroupID:     payload.GroupID,
				RecipientID: member.UserID,
				Ciphertext:  data,
				Delivered:   false,
			}
			if err := d.repos.SenderKeys.Enqueue(dist); err != nil {
				d.logger.Error("failed to queue sender key distribution",
					"group", payload.GroupID, "recipient", member.UserID, "error", err)
			}
		}
	}
	return nil
}

func (d *Dispatcher) handleProfilePictureUpdated(sender *Client, env *protocol.Envelope) error {
	payload, err := protocol.ParseProfilePictureUpdated(env)
	if err != nil {
		return err
	}
	if payload.Version <= 0 {
		return protocol.NewError(protocol.ErrInvalidPayload, "profile picture version must be positive")
	}

	peers, err := d.repos.Relationships.AcceptedPeers(sender.UserID)
	if err != nil {
		d.logger.Error("failed to list accepted peers for profile picture update",
			"sender", sender.UserID, "error", err)
		return err
	}

	data, err := protocol.Encode(env)
	if err != nil {
		return err
	}

	for _, peerID := range peers {
		recipient, online := d.registry.Get(peerID)
		if online {
			recipient.Send(data)
		} else {
			msg := &models.PendingMessage{
				ID:          uuid.New().String(),
				RecipientID: peerID,
				Ciphertext:  data,
				Delivered:   false,
			}
			if err := d.repos.Messages.Insert(msg); err != nil {
				d.logger.Error("failed to queue offline profile picture update",
					"recipient", peerID, "error", err)
			}
		}
	}
	return nil
}

// reencodePayload re-marshals a modified payload back into the envelope so
// server-stamped fields (sender_id) reach the recipient.
func (d *Dispatcher) reencodePayload(env *protocol.Envelope, payload any) error {
	raw, err := protocol.Encode(payload)
	if err != nil {
		return err
	}
	env.Payload = raw
	return nil
}

func (d *Dispatcher) handleGroupProfilePictureUpdated(sender *Client, env *protocol.Envelope) error {
	payload, err := protocol.ParseGroupProfilePictureUpdated(env)
	if err != nil {
		return err
	}
	if payload.Version <= 0 {
		return protocol.NewError(protocol.ErrInvalidPayload, "group profile picture version must be positive")
	}

	members, err := d.repos.Groups.ListMembers(payload.GroupID)
	if err != nil {
		d.logger.Error("failed to list members for group profile picture update",
			"sender", sender.UserID, "group", payload.GroupID, "error", err)
		return err
	}

	data, err := protocol.Encode(env)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.UserID == sender.UserID {
			continue
		}
		recipient, online := d.registry.Get(member.UserID)
		if online {
			recipient.Send(data)
		} else {
			msg := &models.PendingMessage{
				ID:          uuid.New().String(),
				RecipientID: member.UserID,
				Ciphertext:  data,
				Delivered:   false,
			}
			if err := d.repos.Messages.Insert(msg); err != nil {
				d.logger.Error("failed to queue offline group profile picture update",
					"recipient", member.UserID, "group", payload.GroupID, "error", err)
			}
		}
	}
	return nil
}
