package backend

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
)

// ServerRequest is a request from the backend to the ws server to perform a specific action(and might return a response)
type ServerRequest struct {
	RequestID string         `json:"request_id"`
	Action    ServerActions  `json:"action"`
	Params    map[string]any `json:"params"`
}

// BackendAction is a queued request data(it basically extends the ServerRequest struct with the conn field, so replies can be sent with RequestID)
type BackendAction struct {
	ServerRequest
	conn net.Conn
}

// SocketConnectionRequest is a queued request with call back to accept or reject the connection
type BackendRequest struct {
	RequestID string         `json:"request_id"`
	Action    BackendActions `json:"action"`
	Params    map[string]any `json:"params"`
}

// BackendResponse is a response from the server to the backend
type BackendResponse struct {
	RequestID string         `json:"request_id"`
	Action    ServerActions  `json:"action"`
	Params    map[string]any `json:"params"`
}

// TCPServer represents a TCP server instance
type TCPServer struct {
	listener              net.Listener
	Clients               []net.Conn
	mu                    sync.Mutex
	PendingServerRequests chan BackendAction              // Requests from the backend(python) to the server
	PendingResponses      map[string]chan BackendResponse // Responses pending from the backend(python)

	WebsocketActions map[ServerActions]func(map[string]any)
}

// NewTCPServer creates a new TCP server instance
func NewTCPServer() *TCPServer {
	server := &TCPServer{
		Clients:               make([]net.Conn, 0),
		PendingServerRequests: make(chan BackendAction),
		PendingResponses:      make(map[string]chan BackendResponse),
	}
	return server
}

// Start starts the TCP server on the specified address
func (s *TCPServer) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.listener = listener
	log.Printf("TCP Server listening on %s", address)

	go s.acceptConnections()
	go s.HandleRequests()
	return nil
}

// Stop gracefully shuts down the TCP server
func (s *TCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close all client connections
	for _, conn := range s.Clients {
		conn.Close()
	}

	// Close the listener
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// GetListener returns the underlying net.Listener
func (s *TCPServer) GetListener() net.Listener {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener
}

// handleConnection handles a new connection
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		// Remove the connection from the slice
		for i, c := range s.Clients {
			if c == conn {
				s.Clients = append(s.Clients[:i], s.Clients[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		log.Println("Received message", message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		log.Println("Marshalled message", message)
		var request ServerRequest
		err = json.Unmarshal([]byte(message), &request)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			return
		}
		s.PendingServerRequests <- BackendAction{
			ServerRequest: request,
			conn:          conn,
		}
	}
}

// acceptConnections handles incoming connections
func (s *TCPServer) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		s.mu.Lock()
		s.Clients = append(s.Clients, conn)
		s.mu.Unlock()
		log.Println("Accepted connection", len(s.Clients))
		go s.handleConnection(conn)
	}
}

// sendMessage sends a message to the client
func (s *TCPServer) sendMessage(_message BackendRequest) error {
	message, err := json.Marshal(_message)
	if err != nil {
		return err
	}
	conn, ok := s.GetRandomConnection()
	if !ok {
		return errors.New("no connection available")
	}
	writer := bufio.NewWriter(conn)
	writer.WriteString(string(message) + "\n")
	log.Println("Sent message to", conn.RemoteAddr())
	return writer.Flush()
}

// GetRandomConnection returns a random connection from the active clients
func (s *TCPServer) GetRandomConnection() (net.Conn, bool) {
	log.Println("Getting random connection")
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Clients) == 0 {
		log.Println("No connections available")
		return nil, false
	}

	randomIndex := rand.Intn(len(s.Clients))
	log.Println("Random connection", randomIndex)
	return s.Clients[randomIndex], true
}
