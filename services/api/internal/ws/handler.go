package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// ServeWS handles WebSocket upgrade requests.
// It expects a URL parameter "id" (the run ID) and creates a client
// registered to room "run:<id>".
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["id"]
	if runID == "" {
		http.Error(w, "Missing run ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		RoomID: "run:" + runID,
	}

	hub.Register(client)

	// Start goroutines for reading and writing.
	go client.WritePump()
	go client.ReadPump()
}
