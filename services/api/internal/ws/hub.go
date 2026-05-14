package ws

import (
	"log"
	"sync"
)

// BroadcastMessage wraps a room ID and data payload for broadcasting.
type BroadcastMessage struct {
	RoomID string
	Data   []byte
}

// Hub manages WebSocket rooms and client connections.
// Each room is identified by a string (e.g., "run:<runID>") and contains
// a set of connected clients.
type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	mu         sync.RWMutex
}

// NewHub creates and starts a new Hub.
func NewHub() *Hub {
	h := &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage, 256),
	}
	go h.Run()
	return h
}

// Run starts the main event loop for the hub.
// It processes register, unregister, and broadcast events serially.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.RoomID]; !ok {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered to room %s", client.RoomID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.RoomID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered from room %s", client.RoomID)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.rooms[message.RoomID]
			if !ok {
				h.mu.RUnlock()
				continue
			}
			for client := range clients {
				select {
				case client.Send <- message.Data:
				default:
					// Client's send buffer is full; drop the message.
					log.Printf("Dropping message for slow client in room %s", message.RoomID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// BroadcastToRoom sends data to all clients in a room.
// This is thread-safe and non-blocking (drops messages if broadcast channel is full).
func (h *Hub) BroadcastToRoom(roomID string, data []byte) {
	select {
	case h.broadcast <- BroadcastMessage{RoomID: roomID, Data: data}:
	default:
		log.Printf("Broadcast channel full; dropping message for room %s", roomID)
	}
}
