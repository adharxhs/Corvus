package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"

	"server/models"
)

type Client struct {
	UserID       string
	Username     string
	conn         *websocket.Conn
	state        models.ConnectionState
	lastActivity time.Time
	mu           sync.Mutex
	send         chan []byte
}

func NewClient(userID, username string, conn *websocket.Conn) *Client {
	return &Client{
		UserID:       userID,
		Username:     username,
		conn:         conn,
		state:        models.StateConnected,
		lastActivity: time.Now(),
		send:         make(chan []byte, 256),
	}
}

func (c *Client) Send(msg []byte) bool {
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == models.StateConnected {
		c.state = models.StateDisconnected
		close(c.send)
		c.conn.Close(websocket.StatusNormalClosure, "closing")
	}
}

func (c *Client) WriteLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
