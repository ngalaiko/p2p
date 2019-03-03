package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/ngalayko/p2p/instance"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/logger"
)

// WebSocket serves data to the ui.
type WebSocket struct {
	log      *logger.Logger
	upgrader *websocket.Upgrader
	instance *instance.Instance
}

// New is returns new websocket handler.
func New(
	log *logger.Logger,
	instance *instance.Instance,
) *WebSocket {
	return &WebSocket{
		log:      log.Prefix("ui-ws"),
		upgrader: &websocket.Upgrader{},
		instance: instance,
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

	if err := conn.WriteJSON(newInitMessage(ws.instance.Peer)); err != nil {
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
		case messageTypeTextSent:
			if err := ws.instance.SendText(context.Background(), m.Message.Text, m.Message.To.ID); err != nil {
				ws.log.Error("can't send message: %s", err)
				continue
			}
		}
	}
}

func (ws *WebSocket) watchUpdates(conn *websocket.Conn) {
	for {
		select {
		case <-ws.instance.KnownPeers.Updated():
			for _, peer := range ws.instance.KnownPeers.Map() {
				if err := conn.WriteJSON(newPeerAddedMessage(peer)); err != nil {
					ws.log.Error("error writing new peer message: %s", err)
					return
				}
			}
		case msg := <-ws.instance.Sent():
			switch msg.Type {
			case messages.TypeText:
				if err := conn.WriteJSON(newTextMessageSent(msg)); err != nil {
					ws.log.Error("error writing sent message: %s", err)
					return
				}
			}
		case msg := <-ws.instance.Received():
			switch msg.Type {
			case messages.TypeText:
				if err := conn.WriteJSON(newTextMessageReceived(msg)); err != nil {
					ws.log.Error("error writing received message: %s", err)
					return
				}
			}
		}
	}
}
