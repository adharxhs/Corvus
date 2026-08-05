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

func (r *Registry) Unregister(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, userID)
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
