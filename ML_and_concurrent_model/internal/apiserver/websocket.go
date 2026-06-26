package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins (dev)
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Hub maintains the set of active WebSocket clients and broadcasts messages to them.
type Hub struct {
	mu      sync.Mutex
	clients map[*wsClient]struct{}
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

func newHub() *Hub {
	return &Hub{clients: make(map[*wsClient]struct{})}
}

// Broadcast sends a JSON-encoded message to every connected client.
func (h *Hub) Broadcast(msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			// Slow client — drop message rather than block the broadcast.
		}
	}
}

func (h *Hub) register(c *wsClient) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send) // guarded: only closed once
	}
}

// serveWS upgrades the HTTP connection to WebSocket and starts read/write pumps.
func (h *Hub) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	c := &wsClient{conn: conn, send: make(chan []byte, 32)}
	h.register(c)

	go c.writePump(h)
	go c.readPump(h)
}

// writePump relays messages from the hub to the WebSocket connection.
func (c *wsClient) writePump(h *Hub) {
	defer func() {
		h.unregister(c)
		c.conn.Close()
	}()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// readPump discards incoming messages (we only push from server to client)
// and handles disconnections.
func (c *wsClient) readPump(h *Hub) {
	defer func() {
		h.unregister(c)
		c.conn.Close()
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
