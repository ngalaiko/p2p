package ws

import (
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/ngalayko/p2p/app/logger"
	"github.com/ngalayko/p2p/app/peers"
)

// WebSocket serves data to the ui.
type WebSocket struct {
	log      *logger.Logger
	self     *peers.Peer
	upgrader *websocket.Upgrader
}

// New is returns new websocket handler.
func New(
	log *logger.Logger,
	self *peers.Peer,
) *WebSocket {
	return &WebSocket{
		log:      log.Prefix("ws"),
		self:     self,
		upgrader: &websocket.Upgrader{},
	}
}

// ServeHTTP implements http.Server.
func (ws *WebSocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.log.Error("error upgrading a connection: %s", err)
		return
	}

	go ws.watchUpdates(conn)

	if err := conn.WriteJSON(newInitMessage(ws.self)); err != nil {
		ws.log.Error("error writing init message: %s", err)
		return
	}
}

func (ws *WebSocket) watchUpdates(conn *websocket.Conn) {
	for {
		select {
		case <-ws.self.KnownPeers.Updated():
			if err := conn.WriteJSON(newPeersUpdateMessage(ws.self)); err != nil {
				ws.log.Error("error writing update message: %s", err)
				return
			}
		}
	}
}
