package websocket

import (
	"sync"

	"server/models"
)

type Registry struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[string]*Client),
	}
}

func (r *Registry) Register(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, exists := r.clients[c.UserID]; exists {
		old.Close()
	}
	r.clients[c.UserID] = c
}

// UnregisterIfCurrent removes the user from the registry only if the stored
// client is the same one that called this (i.e. it hasn't been superseded by
// a newer reconnection). This prevents a reconnecting user from being
// unregistered by the deferred cleanup of an old connection.
func (r *Registry) UnregisterIfCurrent(userID string, expected *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cur, ok := r.clients[userID]; ok && cur == expected {
		delete(r.clients, userID)
	}
}

func (r *Registry) Get(userID string) (*Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clients[userID]
	if !ok || c.state != models.StateConnected {
		return nil, false
	}
	return c, true
}
