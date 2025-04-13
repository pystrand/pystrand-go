package backend

import (
	"log"

	"github.com/google/uuid"
)

// ServerActions enum
type ServerActions string
type BackendActions string

const (
	// ServerActions
	ServerActionResponse            ServerActions = "response"
	ServerActionMessageToRoom       ServerActions = "message_to_room"
	ServerActionMessageToConnection ServerActions = "message_to_connection"
	ServerActionBroadcastMessage    ServerActions = "broadcast"

	// BackendActions
	BackendActionConnectionRequest BackendActions = "connection_request"
	BackendActionNewMessage        BackendActions = "new_message"
	BackendActionDisconnected      BackendActions = "disconnected"
)

// tells if the connection should be accepted or not
func (s *TCPServer) NewSocketConnection(headers map[string][]string, url string, remoteAddr string) (map[string]any, error) {
	// random connection from pool server clients
	requestID := uuid.New().String()
	err := s.sendMessage(BackendRequest{
		RequestID: requestID,
		Action:    BackendActionConnectionRequest,
		Params: map[string]any{
			"headers":     headers,
			"url":         url,
			"remote_addr": remoteAddr,
		},
	})
	if err != nil {
		return nil, err
	}
	s.PendingResponses[requestID] = make(chan BackendResponse)
	log.Println("Sent request to backend:", requestID, BackendActionConnectionRequest)
	response := <-s.PendingResponses[requestID]
	log.Println("Received response channel:", response.RequestID, response.Action)
	delete(s.PendingResponses, requestID)
	return response.Params, nil
}

func (s *TCPServer) HandleMessage(metaData map[string]any, message []byte) {
	requestID := uuid.New().String()
	s.sendMessage(BackendRequest{
		RequestID: requestID,
		Action:    BackendActionNewMessage,
		Params:    map[string]any{"message": message, "metaData": metaData},
	})
}

func (s *TCPServer) HandleDisconnect(metaData map[string]any) {
	requestID := uuid.New().String()
	s.sendMessage(BackendRequest{
		RequestID: requestID,
		Action:    BackendActionDisconnected,
		Params:    map[string]any{"metaData": metaData},
	})
}

// Accepts a socket connection
func (s *TCPServer) HandleRequests() {
	for request := range s.PendingServerRequests {
		log.Println("Received request:", request.RequestID, request.Action, request.Params)
		switch request.Action {
		case ServerActionResponse:
			s.PendingResponses[request.RequestID] <- BackendResponse{
				RequestID: request.RequestID,
				Action:    ServerActionResponse,
				Params:    request.Params,
			}
		default:
			if action, ok := s.WebsocketActions[request.Action]; ok {
				action(request.Params)
			}
		}
	}
}
