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
func NewRouter(logger *slog.Logger, authSvc *auth.Service, svcs *services.Services, wsHandler http.Handler) http.Handler {
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

	r.mux.Handle("GET /users", requireAuth(http.HandlerFunc(r.handlePlaceholder)))
	r.mux.Handle("GET /messages", requireAuth(http.HandlerFunc(r.handlePlaceholder)))
	r.mux.Handle("GET /groups", requireAuth(http.HandlerFunc(r.handlePlaceholder)))
	
	r.mux.Handle("POST /prekey", requireAuth(http.HandlerFunc(prekeyHandler.Upsert)))
	r.mux.Handle("GET /prekey/{id}", requireAuth(http.HandlerFunc(prekeyHandler.Get)))

	r.mux.Handle("GET /ws", requireAuth(wsHandler))

	return middleware.Chain(
		r.mux,
		middleware.Recover(logger),
		middleware.RequestLogger(logger),
	)
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handlePlaceholder(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
}
