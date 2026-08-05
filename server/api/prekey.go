package api

import (
	"encoding/base64"
	"net/http"

	"server/auth"
	"server/models"
	"server/services"
)

type PrekeyHandler struct {
	prekeys *services.PrekeyService
}

func NewPrekeyHandler(prekeys *services.PrekeyService) *PrekeyHandler {
	return &PrekeyHandler{prekeys: prekeys}
}

func (h *PrekeyHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.PrekeyBundleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ik, err1 := base64.StdEncoding.DecodeString(req.IdentityKey)
	spk, err2 := base64.StdEncoding.DecodeString(req.SignedPrekey)
	sig, err3 := base64.StdEncoding.DecodeString(req.SignedPrekeySignature)
	if err1 != nil || err2 != nil || err3 != nil {
		writeError(w, http.StatusBadRequest, "malformed base64")
		return
	}

	bundle := &models.PrekeyBundle{
		UserID:                userID,
		IdentityKey:           ik,
		SignedPrekey:          spk,
		SignedPrekeySignature: sig,
	}

	if req.OneTimePrekey != "" {
		opk, err := base64.StdEncoding.DecodeString(req.OneTimePrekey)
		if err != nil {
			writeError(w, http.StatusBadRequest, "malformed base64")
			return
		}
		bundle.OneTimePrekey = opk
	}

	if err := h.prekeys.Upsert(bundle); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save prekeys")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PrekeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing user ID")
		return
	}

	b, err := h.prekeys.Get(userID)
	if err != nil {
		if _, ok := err.(services.ErrNotFound); ok {
			writeError(w, http.StatusNotFound, "prekeys not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get prekeys")
		}
		return
	}

	resp := models.PrekeyBundleResponse{
		UserID:                b.UserID,
		IdentityKey:           base64.StdEncoding.EncodeToString(b.IdentityKey),
		SignedPrekey:          base64.StdEncoding.EncodeToString(b.SignedPrekey),
		SignedPrekeySignature: base64.StdEncoding.EncodeToString(b.SignedPrekeySignature),
	}
	if len(b.OneTimePrekey) > 0 {
		resp.OneTimePrekey = base64.StdEncoding.EncodeToString(b.OneTimePrekey)
	}

	writeJSON(w, http.StatusOK, resp)
}
