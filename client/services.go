package client

import "github.com/gorilla/websocket"

func (r *Room) BroadcastMessage(message []byte) {
	for _, client := range r.clients {
		client.SendMessage(message)
	}
}

func (c *Client) SendMessage(message []byte) {
	c.Conn.WriteMessage(websocket.TextMessage, message)
}

func (s *WebSocketServer) MessageToRoom(roomID string, message []byte) {
	s.rooms[roomID].BroadcastMessage(message)
}

func (s *WebSocketServer) MessageToConnection(connID string, message []byte) {
	for _, client := range s.clients {
		if client.Conn.RemoteAddr().String() == connID {
			client.SendMessage(message)
			return
		}
	}
}
