package websocket

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"server/auth"
	"server/models"
	"server/protocol"
	"server/repository"
)

type Server struct {
	registry   *Registry
	dispatcher *Dispatcher
	repos      *repository.Repository
	logger     *slog.Logger
}

func NewServer(repos *repository.Repository, logger *slog.Logger) *Server {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry, repos, logger)
	return &Server{
		registry:   registry,
		dispatcher: dispatcher,
		repos:      repos,
		logger:     logger,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	username, _ := auth.UsernameFromContext(r)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		s.logger.Error("websocket accept failed", "error", err)
		return
	}

	client := NewClient(userID, username, conn)
	s.registry.Register(client)
	s.sendPresenceSnapshot(client)
	s.broadcastPresence(userID, "online")
	defer func() {
		s.registry.UnregisterIfCurrent(userID, client)
		s.broadcastPresence(userID, "offline")
		client.Close()
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go client.WriteLoop(ctx)

	s.deliverPendingMessages(ctx, client)

	s.readLoop(ctx, client)
}

func (s *Server) deliverPendingMessages(ctx context.Context, client *Client) {
	messages, err := s.repos.Messages.FetchByRecipient(client.UserID)
	if err == nil {
		for _, msg := range messages {
			if client.Send(msg.Ciphertext) {
				_ = s.repos.Messages.MarkDelivered(msg.ID)
			}
		}
		_ = s.repos.Messages.DeleteDelivered()
	}

	dists, err := s.repos.SenderKeys.FetchByRecipient(client.UserID)
	if err == nil {
		for _, dist := range dists {
			if client.Send(dist.Ciphertext) {
				_ = s.repos.SenderKeys.MarkDelivered(dist.ID)
			}
		}
		_ = s.repos.SenderKeys.DeleteDelivered()
	}
}

func (s *Server) SendToUser(userID string, msg []byte) {
	if client, ok := s.registry.Get(userID); ok {
		client.Send(msg)
	}
}

// BroadcastToGroup sends msg to every member of the group. Online members
// receive it immediately; offline members have it queued in PendingMessages.
func (s *Server) BroadcastToGroup(groupID string, msg []byte) {
	members, err := s.repos.Groups.ListMembers(groupID)
	if err != nil {
		s.logger.Error("failed to list members for group broadcast",
			"group", groupID, "error", err)
		return
	}
	for _, member := range members {
		if client, ok := s.registry.Get(member.UserID); ok {
			client.Send(msg)
		} else {
			p := &models.PendingMessage{
				ID:          uuid.New().String(),
				RecipientID: member.UserID,
				Ciphertext:  msg,
				Delivered:   false,
			}
			if err := s.repos.Messages.Insert(p); err != nil {
				s.logger.Error("failed to queue offline group broadcast",
					"group", groupID, "recipient", member.UserID, "error", err)
			}
		}
	}
}

func (s *Server) readLoop(ctx context.Context, client *Client) {
	for {
		_, data, err := client.conn.Read(ctx)
		if err != nil {
			return
		}

		env, err := protocol.ParseEnvelope(data)
		if err != nil {
			if pErr, ok := err.(*protocol.Error); ok {
				client.Send(protocol.EncodeError(pErr.Code, pErr.Message))
			} else {
				client.Send(protocol.EncodeError(protocol.ErrMalformedJSON, "invalid frame"))
			}
			continue
		}

		if err := s.dispatcher.Dispatch(client, env); err != nil {
			if pErr, ok := err.(*protocol.Error); ok {
				client.Send(protocol.EncodeError(pErr.Code, pErr.Message))
			} else {
				client.Send(protocol.EncodeError(protocol.ErrDispatch, err.Error()))
			}
		}
	}
}
