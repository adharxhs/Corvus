package websocket

import (
	"context"
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

func (d *Dispatcher) Dispatch(ctx context.Context, sender *Client, env *protocol.Envelope) error {
	switch env.Type {
	case protocol.TypeMessage:
		return d.handleDirectMessage(ctx, sender, env)
	case protocol.TypeGroupMessage:
		return d.handleGroupMessage(ctx, sender, env)
	case protocol.TypeSenderKeyDistribution:
		return d.handleSenderKeyDistribution(ctx, sender, env)
	default:
		return protocol.NewError(protocol.ErrInvalidType, "unsupported message type")
	}
}

func (d *Dispatcher) handleDirectMessage(ctx context.Context, sender *Client, env *protocol.Envelope) error {
	payload, err := protocol.ParseDirectMessage(env)
	if err != nil {
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
			d.logger.Error("failed to queue offline message", "error", err)
			return err
		}
	}
	return nil
}

func (d *Dispatcher) handleGroupMessage(ctx context.Context, sender *Client, env *protocol.Envelope) error {
	payload, err := protocol.ParseGroupMessage(env)
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
			msg := &models.PendingMessage{
				ID:          uuid.New().String(),
				RecipientID: member.UserID,
				Ciphertext:  data,
				Delivered:   false,
			}
			_ = d.repos.Messages.Insert(msg)
		}
	}
	return nil
}

func (d *Dispatcher) handleSenderKeyDistribution(ctx context.Context, sender *Client, env *protocol.Envelope) error {
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
			_ = d.repos.SenderKeys.Enqueue(dist)
		}
	}
	return nil
}
