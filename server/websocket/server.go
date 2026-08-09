package websocket

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"

	"server/auth"
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
		s.registry.Unregister(userID)
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
