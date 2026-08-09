package api

import (
	"encoding/base64"
	"net/http"

	"server/auth"
	"server/models"
	"server/services"
)

type ProfilePictureHandler struct {
	pictures *services.ProfilePictureService
}

func NewProfilePictureHandler(pictures *services.ProfilePictureService) *ProfilePictureHandler {
	return &ProfilePictureHandler{pictures: pictures}
}

func (h *ProfilePictureHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.ProfilePictureRequest
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

	if err := h.pictures.Upload(userID, ciphertext, nonce, req.Version); err != nil {
		switch err := err.(type) {
		case services.ErrInvalidInput:
			writeError(w, http.StatusBadRequest, err.Error())
		case services.ErrConflict:
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to upload profile picture")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProfilePictureHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID := r.PathValue("id")
	if targetID == "" {
		writeError(w, http.StatusBadRequest, "missing user ID")
		return
	}

	pic, err := h.pictures.Get(userID, targetID)
	if err != nil {
		switch err := err.(type) {
		case services.ErrNotAccepted:
			writeError(w, http.StatusForbidden, err.Error())
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, "no profile picture")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get profile picture")
		}
		return
	}

	writeJSON(w, http.StatusOK, models.ProfilePictureResponse{
		Ciphertext: base64.StdEncoding.EncodeToString(pic.Ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(pic.Nonce),
		Version:    pic.Version,
	})
}
