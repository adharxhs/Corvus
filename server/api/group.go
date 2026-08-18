package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"server/auth"
	"server/models"
	"server/protocol"
	"server/services"
)

type GroupHandler struct {
	groups  *services.GroupService
	notifier Notifier
}

func NewGroupHandler(groups *services.GroupService, notifier Notifier) *GroupHandler {
	return &GroupHandler{groups: groups, notifier: notifier}
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateGroupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	group, err := h.groups.CreateGroup(req.GroupID, req.Name, userID)
	if err != nil {
		switch err := err.(type) {
		case services.ErrInvalidInput:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create group")
		}
		return
	}
	writeJSON(w, http.StatusCreated, group)
}

func (h *GroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}
	group, err := h.groups.GetGroup(groupID)
	if err != nil {
		switch err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to get group")
		}
		return
	}
	writeJSON(w, http.StatusOK, group)
}

func (h *GroupHandler) Rename(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}

	var req models.RenameGroupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.groups.RenameGroup(groupID, userID, req.Name); err != nil {
		switch err := err.(type) {
		case services.ErrNotMember:
			writeError(w, http.StatusForbidden, err.Error())
		case services.ErrInvalidInput:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to rename group")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}
	members, err := h.groups.ListMembers(groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (h *GroupHandler) Invite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}

	var req models.GroupInviteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.groups.Invite(groupID, userID, req.UserID); err != nil {
		switch err := err.(type) {
		case services.ErrNotMember:
			writeError(w, http.StatusForbidden, err.Error())
		case services.ErrNotAccepted:
			writeError(w, http.StatusForbidden, err.Error())
		case services.ErrConflict:
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to send invite")
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *GroupHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groups, err := h.groups.ListByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list groups")
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func (h *GroupHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	invites, err := h.groups.ListInvites(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list invites")
		return
	}
	writeJSON(w, http.StatusOK, invites)
}

func (h *GroupHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}
	username, _ := auth.UsernameFromContext(r)
	if err := h.groups.AcceptInvite(groupID, userID); err != nil {
		switch err := err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case services.ErrConflict:
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to accept invite")
		}
		return
	}

	h.broadcastMemberJoined(groupID, userID, username)
	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

func (h *GroupHandler) RejectInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}
	if err := h.groups.RejectInvite(groupID, userID); err != nil {
		switch err := err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case services.ErrConflict:
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to reject invite")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) Leave(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}
	if err := h.groups.Leave(groupID, userID); err != nil {
		switch err := err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to leave group")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) UploadPicture(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}

	var req models.GroupProfilePictureRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.ImageData == "" {
		writeError(w, http.StatusBadRequest, "image_data is required")
		return
	}

	imageBytes, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		writeError(w, http.StatusBadRequest, "malformed base64")
		return
	}

	if err := h.groups.UploadGroupPicture(groupID, userID, imageBytes, req.Version); err != nil {
		switch err := err.(type) {
		case services.ErrInvalidInput:
			writeError(w, http.StatusBadRequest, err.Error())
		case services.ErrNotMember:
			writeError(w, http.StatusForbidden, err.Error())
		case services.ErrConflict:
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to upload group profile picture")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) GetPicture(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "missing group_id")
		return
	}

	pic, err := h.groups.GetGroupPicture(groupID, userID)
	if err != nil {
		switch err := err.(type) {
		case services.ErrNotMember:
			writeError(w, http.StatusForbidden, err.Error())
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to get group profile picture")
		}
		return
	}

	writeJSON(w, http.StatusOK, models.GroupProfilePictureResponse{
		ImageData: base64.StdEncoding.EncodeToString(pic.ImageData),
		Version:   pic.Version,
	})
}

func (h *GroupHandler) broadcastMemberJoined(groupID, userID, username string) {
	if h.notifier == nil {
		return
	}
	payload := protocol.MemberJoinedPayload{
		GroupID:  groupID,
		UserID:   userID,
		Username: username,
	}
	raw, _ := json.Marshal(payload)
	env := protocol.Envelope{
		Version: protocol.CurrentVersion,
		Type:    protocol.TypeMemberJoined,
	}
	env.Payload = raw
	data, _ := json.Marshal(env)
	h.notifier.BroadcastToGroup(groupID, data)
}
