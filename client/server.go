package client

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	MetaData map[string]any
	RoomID   string
	ClientID string
}

type Room struct {
	ID      string
	clients map[string]*Client
}

// WebSocketServer represents the WebSocket server
type WebSocketServer struct {
	upgrader        websocket.Upgrader
	rooms           map[string]*Room
	clients         map[string]*Client
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
		clients:         make(map[string]*Client),
		onNewConnection: onNewConnection,
		onMessage:       onMessage,
		onDisconnect:    onDisconnect,
	}
}

func (s *WebSocketServer) Start(port string) {
	http.HandleFunc("/ws/", s.HandleConnection)
	http.ListenAndServe(port, nil)
}

func (s *WebSocketServer) Stop() {
	for _, client := range s.clients {
		client.Conn.Close()
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
		ClientID: metaData["clientID"].(string),
	}
	s.clients[client.ClientID] = client
	room.clients[client.ClientID] = client

	log.Printf("New client connected. Total clients: %d", len(s.clients))

	// Handle messages from this client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			s.handleDisconnect(client.ClientID)
			return
		}

		// Handle the received message
		s.onMessage(*client, message)
	}
}

// handleDisconnect handles client disconnection
func (s *WebSocketServer) handleDisconnect(clientID string) {
	client := s.clients[clientID]
	if client != nil {
		s.onDisconnect(*client)
		delete(s.clients, client.ClientID)
		log.Printf("Client disconnected. Total clients: %d", len(s.clients))
		client.Conn.Close()
	}
}

// BroadcastMessage sends a message to all connected clients
func (s *WebSocketServer) BroadcastMessage(message []byte) {
	for _, client := range s.clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
			client.Conn.Close()
			delete(s.clients, client.ClientID)
		}
	}
}

// SendMessage sends a message to a specific client
func (s *WebSocketServer) SendMessage(clientID string, message []byte) error {
	client := s.clients[clientID]
	if client == nil {
		log.Printf("client not found")
		return errors.New("client not found")
	}
	return client.Conn.WriteMessage(websocket.TextMessage, message)
}
