package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server/auth"
	"server/models"
	"server/protocol"
	"server/services"
)

// Notifier is used by HTTP handlers to push real-time WebSocket notifications
// to connected clients.
type Notifier interface {
	SendToUser(userID string, msg []byte)
	BroadcastToGroup(groupID string, msg []byte)
}

type RelationshipHandler struct {
	relationships *services.RelationshipService
	notifier      Notifier
}

func NewRelationshipHandler(relationships *services.RelationshipService, notifier Notifier) *RelationshipHandler {
	return &RelationshipHandler{relationships: relationships, notifier: notifier}
}

func (h *RelationshipHandler) notifyChatRequest(targetID, requesterID, recipientID, status string) {
	if h.notifier == nil || targetID == "" {
		return
	}
	payload := protocol.ChatRequestUpdatedPayload{
		RequesterID: requesterID,
		RecipientID: recipientID,
		Status:      status,
	}
	raw, _ := json.Marshal(payload)
	env := protocol.Envelope{
		Version: protocol.CurrentVersion,
		Type:    protocol.TypeChatRequestUpdated,
	}
	env.Payload = raw
	data, _ := json.Marshal(env)
	h.notifier.SendToUser(targetID, data)
}

func (h *RelationshipHandler) notifyRequestParties(requesterID, recipientID, status string) {
	// The recipient must see newly-pending requests; on accept/reject both
	// parties should be refreshed in real time.
	h.notifyChatRequest(recipientID, requesterID, recipientID, status)
	if status == string(models.RelationshipAccepted) || status == string(models.RelationshipRejected) {
		h.notifyChatRequest(requesterID, requesterID, recipientID, status)
	}
}

func (h *RelationshipHandler) Send(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.ChatRequestRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.RecipientID == "" {
		writeError(w, http.StatusBadRequest, "recipient_id is required")
		return
	}

	rel, err := h.relationships.SendRequest(userID, req.RecipientID)
	if err != nil {
		switch err := err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case services.ErrConflict:
			writeError(w, http.StatusConflict, err.Error())
		case services.ErrCooldownActive:
			w.Header().Set("Retry-After", strconv.FormatInt(int64(err.RetryAfter.Seconds()), 10))
			writeError(w, http.StatusTooManyRequests, err.Error())
		case services.ErrInvalidInput:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to send chat request")
		}
		return
	}

	writeJSON(w, http.StatusCreated, rel)
	h.notifyRequestParties(userID, req.RecipientID, "pending")
}

func (h *RelationshipHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rels, err := h.relationships.ListPending(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list chat requests")
		return
	}
	writeJSON(w, http.StatusOK, rels)
}

func (h *RelationshipHandler) Accept(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, models.RelationshipAccepted)
}

func (h *RelationshipHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, models.RelationshipRejected)
}

func (h *RelationshipHandler) respond(w http.ResponseWriter, r *http.Request, status models.RelationshipStatus) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	requesterID := r.PathValue("requester_id")
	if requesterID == "" {
		writeError(w, http.StatusBadRequest, "missing requester_id")
		return
	}
	rel, err := h.relationships.Respond(userID, requesterID, status)
	if err != nil {
		switch err := err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case services.ErrConflict:
			writeError(w, http.StatusConflict, err.Error())
		case services.ErrInvalidInput:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to respond to chat request")
		}
		return
	}
	writeJSON(w, http.StatusOK, rel)
	h.notifyRequestParties(requesterID, userID, string(status))
}
