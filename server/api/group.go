package api

import (
	"encoding/base64"
	"net/http"

	"server/auth"
	"server/models"
	"server/services"
)

type GroupHandler struct {
	groups *services.GroupService
}

func NewGroupHandler(groups *services.GroupService) *GroupHandler {
	return &GroupHandler{groups: groups}
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

	group, err := h.groups.CreateGroup(req.GroupID, userID)
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
	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
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

	ciphertext, err1 := base64.StdEncoding.DecodeString(req.Ciphertext)
	nonce, err2 := base64.StdEncoding.DecodeString(req.Nonce)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "malformed base64")
		return
	}

	if err := h.groups.UploadGroupPicture(groupID, userID, ciphertext, nonce, req.Version); err != nil {
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
		Ciphertext: base64.StdEncoding.EncodeToString(pic.Ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(pic.Nonce),
		Version:    pic.Version,
	})
}
