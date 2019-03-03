package client

import (
	"encoding/json"
	"net/http"

	"github.com/ngalayko/p2p/instance/peers"
)

func healthcheckHandler(peer *peers.Peer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonBytes, err := json.Marshal(peer)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = w.Write(jsonBytes)
		w.WriteHeader(http.StatusOK)
	}
}
