// Package bridge provides a bridge between the client and the server.
package bridge

import (
	"net/http"

	"github.com/pystrand/pystrand-server/backend"
	"github.com/pystrand/pystrand-server/client"
)

type Bridge struct {
	_backend  *backend.TCPServer
	webSocket *client.WebSocketServer
}

func NewBridge() *Bridge {
	_backend := backend.NewTCPServer()
	_backend.WebsocketActions = make(map[backend.ServerActions]func(map[string]any))

	onNewConnection := func(r *http.Request) (map[string]any, error) {
		return _backend.NewSocketConnection(r.Header, r.URL.Path, r.RemoteAddr)
	}

	onMessage := func(_client client.Client, message []byte) {
		_backend.HandleMessage(_client.MetaData, message)
	}

	onDisconnect := func(_client client.Client) {
		_backend.HandleDisconnect(_client.MetaData)
	}

	webSocket := client.NewWebSocketServer(
		onNewConnection,
		onMessage,
		onDisconnect,
	)

	// add a new action to the backend
	_backend.WebsocketActions[backend.ServerActionMessageToRoom] = func(params map[string]any) {
		webSocket.MessageToRoom(params["room_id"].(string), params["message"].([]byte))
	}
	_backend.WebsocketActions[backend.ServerActionMessageToConnection] = func(params map[string]any) {
		webSocket.MessageToConnection(params["conn_id"].(string), params["message"].([]byte))
	}
	_backend.WebsocketActions[backend.ServerActionBroadcastMessage] = func(params map[string]any) {
		message := params["message"].(string)
		webSocket.BroadcastMessage([]byte(message))
	}

	return &Bridge{
		_backend:  _backend,
		webSocket: webSocket,
	}
}

func (b *Bridge) Start() {
	b._backend.Start(":8081")
	b.webSocket.Start(":8080")
}

func (b *Bridge) Stop() {
	b._backend.Stop()
	b.webSocket.Stop()
}
