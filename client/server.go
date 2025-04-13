package client

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	MetaData map[string]any
	RoomID   string
}

type Room struct {
	ID      string
	clients map[string]*Client
}

// WebSocketServer represents the WebSocket server
type WebSocketServer struct {
	upgrader        websocket.Upgrader
	rooms           map[string]*Room
	clients         map[*websocket.Conn]*Client
	onNewConnection func(r *http.Request) (map[string]any, error)
	onMessage       func(_client Client, message []byte)
	onDisconnect    func(_client Client)
}

// NewWebSocketServer creates a new WebSocket server instance
func NewWebSocketServer(
	onNewConnection func(r *http.Request) (map[string]any, error),
	onMessage func(_client Client, message []byte),
	onDisconnect func(_client Client),
) *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		rooms:           make(map[string]*Room),
		clients:         make(map[*websocket.Conn]*Client),
		onNewConnection: onNewConnection,
		onMessage:       onMessage,
		onDisconnect:    onDisconnect,
	}
}

func (s *WebSocketServer) Start(port string) {
	fmt.Println("Starting WebSocket server on port", port)
	http.HandleFunc("/ws/", s.HandleConnection)
	http.ListenAndServe(port, nil)
}

func (s *WebSocketServer) Stop() {
	for conn := range s.clients {
		conn.Close()
	}
}

// HandleConnection handles new WebSocket connections
func (s *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {
	metaData, err := s.onNewConnection(r)
	if err != nil {
		log.Printf("Failed to handle new connection: %v", err)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	if metaData["accepted"] == false {
		log.Println("Connection rejected by Backend")
		conn.Close()
		return
	}

	roomID := metaData["roomID"].(string)

	room, ok := s.rooms[roomID]
	if !ok {
		room = &Room{
			ID:      roomID,
			clients: make(map[string]*Client),
		}
		s.rooms[roomID] = room
	}

	// Add client to the map
	client := &Client{
		Conn:     conn,
		MetaData: metaData,
		RoomID:   roomID,
	}
	s.clients[conn] = client
	room.clients[metaData["clientID"].(string)] = client

	log.Printf("New client connected. Total clients: %d", len(s.clients))

	// Handle messages from this client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			s.handleDisconnect(conn)
			return
		}

		// Handle the received message
		s.onMessage(*client, message)
	}
}

// handleDisconnect handles client disconnection
func (s *WebSocketServer) handleDisconnect(conn *websocket.Conn) {
	client := s.clients[conn]
	if client != nil {
		s.onDisconnect(*client)
		delete(s.clients, conn)
		log.Printf("Client disconnected. Total clients: %d", len(s.clients))
		conn.Close()
	}
}

// BroadcastMessage sends a message to all connected clients
func (s *WebSocketServer) BroadcastMessage(message []byte) {
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

// SendMessage sends a message to a specific client
func (s *WebSocketServer) SendMessage(conn *websocket.Conn, message []byte) error {
	return conn.WriteMessage(websocket.TextMessage, message)
}
