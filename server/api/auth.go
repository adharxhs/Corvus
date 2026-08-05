package api

import (
	"errors"
	"net/http"

	"server/auth"
	"server/models"
)

type AuthHandler struct {
	auth *auth.Service
}

func NewAuthHandler(auth *auth.Service) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	resp, err := h.auth.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRequest) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, auth.ErrDuplicateUsername) {
			writeError(w, http.StatusConflict, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	resp, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
