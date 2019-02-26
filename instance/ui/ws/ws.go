package ws

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/peers"
)

// WebSocket serves data to the ui.
type WebSocket struct {
	log        *logger.Logger
	self       *peers.Peer
	upgrader   *websocket.Upgrader
	msgHandler *messages.Handler
}

// New is returns new websocket handler.
func New(
	log *logger.Logger,
	self *peers.Peer,
	msgHandler *messages.Handler,
) *WebSocket {
	return &WebSocket{
		log:        log.Prefix("ui-ws"),
		self:       self,
		upgrader:   &websocket.Upgrader{},
		msgHandler: msgHandler,
	}
}

// ServeHTTP implements http.Server.
func (ws *WebSocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	ws.log.Info("new connection from %s", origin)
	defer ws.log.Info("connection from %s closed", origin)

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.log.Error("error upgrading a connection from %s: %s", origin, err)
		return
	}

	go ws.watchUpdates(conn)

	if err := conn.WriteJSON(newInitMessage(ws.self)); err != nil {
		ws.log.Error("error writing init message to %s: %s", origin, err)
		return
	}

	for {
		_, data, err := conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			ws.log.Error("error reading message from %s: %s", origin, err)
			continue
		}

		m := &message{}
		if err := json.Unmarshal(data, m); err != nil {
			continue
		}

		switch m.Type {
		case messageTypeText:
			if err := ws.msgHandler.SendText(r.Context(), m.Text, m.Peer.ID); err != nil {
				ws.log.Error("can't send message: %s", err)
				continue
			}
		}
	}
}

func (ws *WebSocket) watchUpdates(conn *websocket.Conn) {
	for {
		select {
		case new := <-ws.self.KnownPeers.Updated():
			if err := conn.WriteJSON(newPeersAddedMessage(new)); err != nil {
				ws.log.Error("error writing update message: %s", err)
				return
			}
		}
	}
}
