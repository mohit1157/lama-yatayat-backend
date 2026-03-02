package connection

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all for demo
}

type Client struct {
	Conn     *websocket.Conn
	UserID   string
	RideID   string
	Role     string // "rider" or "driver"
	Send     chan []byte
}

type Hub struct {
	clients    map[string]*Client // userID -> client
	rideRooms  map[string]map[string]*Client // rideID -> {userID -> client}
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
}

type Message struct {
	Type    string          `json:"type"`
	RideID  string          `json:"ride_id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rideRooms:  make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			if client.RideID != "" {
				if h.rideRooms[client.RideID] == nil {
					h.rideRooms[client.RideID] = make(map[string]*Client)
				}
				h.rideRooms[client.RideID][client.UserID] = client
			}
			h.mu.Unlock()
			log.Printf("WS: client connected: %s (ride: %s)", client.UserID, client.RideID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			if client.RideID != "" {
				if room, ok := h.rideRooms[client.RideID]; ok {
					delete(room, client.UserID)
					if len(room) == 0 {
						delete(h.rideRooms, client.RideID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if msg.RideID != "" {
				if room, ok := h.rideRooms[msg.RideID]; ok {
					data, _ := json.Marshal(msg)
					for _, client := range room {
						select {
						case client.Send <- data:
						default:
							close(client.Send)
							delete(room, client.UserID)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToRide(rideID string, msg *Message) {
	msg.RideID = rideID
	h.broadcast <- msg
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	userID := r.URL.Query().Get("user_id")
	rideID := r.URL.Query().Get("ride_id")
	role := r.URL.Query().Get("role")

	client := &Client{
		Conn: conn, UserID: userID, RideID: rideID, Role: role,
		Send: make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	go client.readPump(h)
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// Handle incoming messages (e.g., chat) if needed
	}
}
