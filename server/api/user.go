package api

import (
	"net/http"

	"server/services"
)

type UserHandler struct {
	users *services.UserService
}

func NewUserHandler(users *services.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// GetByUsername resolves an exact username to a user ID (new-chat discovery).
// Only the ID is returned on match; 404 on no match. Partial/prefix search is
// intentionally absent to prevent user enumeration.
func (h *UserHandler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := h.users.GetByUsername(username)
	if err != nil {
		switch err := err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to resolve user")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": user.ID})
}

// GetByID resolves a user ID to username (ID -> username resolution for the
// client-side contact list).
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	user, err := h.users.GetByID(userID)
	if err != nil {
		switch err := err.(type) {
		case services.ErrNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to resolve user")
		}
		return
	}
	writeJSON(w, http.StatusOK, user)
}
