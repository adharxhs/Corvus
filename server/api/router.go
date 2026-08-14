package api

import (
	"log/slog"
	"net/http"

	"server/auth"
	"server/middleware"
	"server/services"
)

type Router struct {
	mux      *http.ServeMux
	logger   *slog.Logger
	auth     *auth.Service
	services *services.Services
}

// NewRouter creates and initializes the main HTTP router.
func NewRouter(logger *slog.Logger, authSvc *auth.Service, svcs *services.Services, wsHandler http.Handler, corsOrigin string) http.Handler {
	r := &Router{
		mux:      http.NewServeMux(),
		logger:   logger,
		auth:     authSvc,
		services: svcs,
	}

	r.mux.HandleFunc("GET /", r.handleHealth)
	r.mux.HandleFunc("GET /health", r.handleHealth)

	authHandler := NewAuthHandler(authSvc)
	r.mux.HandleFunc("POST /register", authHandler.Register)
	r.mux.HandleFunc("POST /login", authHandler.Login)

	prekeyHandler := NewPrekeyHandler(svcs.Prekeys)

	requireAuth := middleware.RequireAuth(authSvc.JWT())

	// --- Authenticated routes (require valid JWT) ---

	// User resolution
	r.mux.Handle("GET /users/by-username/{username}", requireAuth(http.HandlerFunc(NewUserHandler(svcs.Users).GetByUsername)))
	r.mux.Handle("GET /users/{id}", requireAuth(http.HandlerFunc(NewUserHandler(svcs.Users).GetByID)))
	r.mux.Handle("POST /user/password", requireAuth(http.HandlerFunc(authHandler.ChangePassword)))

	// Chat requests (pending → accepted | rejected)
	relationshipHandler := NewRelationshipHandler(svcs.Relationships)
	r.mux.Handle("POST /chat-request", requireAuth(http.HandlerFunc(relationshipHandler.Send)))
	r.mux.Handle("GET /chat-requests", requireAuth(http.HandlerFunc(relationshipHandler.List)))
	r.mux.Handle("POST /chat-request/{requester_id}/accept", requireAuth(http.HandlerFunc(relationshipHandler.Accept)))
	r.mux.Handle("POST /chat-request/{requester_id}/reject", requireAuth(http.HandlerFunc(relationshipHandler.Reject)))

	// Groups (create, invite, accept, leave)
	groupHandler := NewGroupHandler(svcs.Groups)
	r.mux.Handle("POST /groups", requireAuth(http.HandlerFunc(groupHandler.Create)))
	r.mux.Handle("GET /groups/invites", requireAuth(http.HandlerFunc(groupHandler.ListInvites)))
	r.mux.Handle("GET /groups/{group_id}/members", requireAuth(http.HandlerFunc(groupHandler.ListMembers)))
	r.mux.Handle("POST /groups/{group_id}/invite", requireAuth(http.HandlerFunc(groupHandler.Invite)))
	r.mux.Handle("POST /groups/{group_id}/invite/accept", requireAuth(http.HandlerFunc(groupHandler.AcceptInvite)))
	r.mux.Handle("DELETE /groups/{group_id}/member", requireAuth(http.HandlerFunc(groupHandler.Leave)))
	r.mux.Handle("POST /groups/{group_id}/profile-picture", requireAuth(http.HandlerFunc(groupHandler.UploadPicture)))
	r.mux.Handle("GET /groups/{group_id}/profile-picture", requireAuth(http.HandlerFunc(groupHandler.GetPicture)))

	// Prekey bundles (X3DH, ungated by chat-request status)
	r.mux.Handle("POST /prekey", requireAuth(http.HandlerFunc(prekeyHandler.Upsert)))
	r.mux.Handle("GET /prekey/{id}", requireAuth(http.HandlerFunc(prekeyHandler.Get)))

	// Profile pictures (encrypted client-side; gated behind accepted relationship)
	picHandler := NewProfilePictureHandler(svcs.ProfilePictures)
	r.mux.Handle("GET /profile-picture/{id}", requireAuth(http.HandlerFunc(picHandler.Get)))
	r.mux.Handle("POST /profile-picture", requireAuth(http.HandlerFunc(picHandler.Upload)))

	// WebSocket
	r.mux.Handle("GET /ws", requireAuth(wsHandler))

	return middleware.Chain(
		r.mux,
		middleware.CORS(corsOrigin),
		middleware.Recover(logger),
		middleware.RequestLogger(logger),
	)
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
